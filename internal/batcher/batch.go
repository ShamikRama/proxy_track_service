package batcher

import (
	"context"
	"log"
	"sync"
	"time"

	"errors"

	"github.com/shamil/proxy_track_service-1/internal/client"
	"github.com/shamil/proxy_track_service-1/internal/config"
	"github.com/shamil/proxy_track_service-1/internal/erors"
	"github.com/shamil/proxy_track_service-1/internal/models"
	"github.com/shamil/proxy_track_service-1/internal/repository"
)

type Batcher struct {
	config config.BatcherConfig
	cache  repository.CacheRepository
	client client.ExternalAPIClient

	mu          sync.Mutex
	batch       []batchItem
	batchTimer  *time.Timer
	inputChan   chan batchRequest
	workerChan  chan []batchItem
	flushSignal chan struct{}
	stopChan    chan struct{}
}

func NewBatcher(config config.BatcherConfig, cache repository.CacheRepository, client client.ExternalAPIClient) BatcherInterface {
	b := &Batcher{
		config:      config,
		cache:       cache,
		client:      client,
		inputChan:   make(chan batchRequest, config.BatchSize*2),
		workerChan:  make(chan []batchItem, config.Workers),
		flushSignal: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
		batch:       make([]batchItem, 0, config.BatchSize),
	}

	b.batchTimer = time.NewTimer(0)
	if !b.batchTimer.Stop() {
		<-b.batchTimer.C
	}

	return b
}

func (b *Batcher) AddRequest(ctx context.Context, trackCode string) <-chan models.TrackResponse {
	respChan := make(chan models.TrackResponse, 1)

	select {
	case b.inputChan <- batchRequest{
		ctx:         ctx,
		trackCode:   trackCode,
		respChannel: respChan,
	}:
		return respChan

	case <-ctx.Done():
		respChan <- models.TrackResponse{
			Status: false,
			Error:  "request cancelled",
		}
		return respChan

	default:
		respChan <- models.TrackResponse{
			Status: false,
			Error:  "service busy, try again later",
		}
		return respChan
	}
}

func (b *Batcher) Start(ctx context.Context) error {
	for i := 0; i < b.config.Workers; i++ {
		go b.worker(ctx)
	}

	go b.timerManager(ctx)

	go b.mainLoop(ctx)

	return nil
}

func (b *Batcher) mainLoop(ctx context.Context) {
	for {
		select {
		case req := <-b.inputChan:
			b.addToBatch(req)

		case <-b.flushSignal:
			b.mu.Lock()
			if len(b.batch) > 0 {
				b.flushBatchLocked()
			}
			b.mu.Unlock()

		case <-ctx.Done():
			b.mu.Lock()
			if len(b.batch) > 0 {
				b.flushBatchLocked()
			}
			b.mu.Unlock()
			return

		case <-b.stopChan:
			return
		}
	}
}

func (b *Batcher) timerManager(ctx context.Context) {
	for {
		b.mu.Lock()
		hasItems := len(b.batch) > 0
		b.mu.Unlock()

		if hasItems {
			b.batchTimer.Reset(b.config.BatchTimeout)

			select {
			case <-b.batchTimer.C:
				select {
				case b.flushSignal <- struct{}{}:
				default:
				}
			case <-ctx.Done():
				return
			case <-b.stopChan:
				return
			}
		} else {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return
			case <-b.stopChan:
				return
			}
		}
	}
}

func (b *Batcher) Stop() error {
	close(b.stopChan)

	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.batchTimer.Stop() {
		select {
		case <-b.batchTimer.C:
		default:
		}
	}

	for _, item := range b.batch {
		select {
		case item.responseChan <- models.TrackResponse{
			Status: false,
			Error:  "service shutting down",
		}:
		default:
		}
	}
	b.batch = b.batch[:0]

	return nil
}

func (b *Batcher) Health(ctx context.Context) error {
	select {
	case b.inputChan <- batchRequest{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		return errors.New("batcher input channel blocked")
	}
}

func (b *Batcher) Flush() {
	select {
	case b.flushSignal <- struct{}{}:
	default:
	}
}

func (b *Batcher) addToBatch(req batchRequest) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if req.ctx != nil {
		select {
		case <-req.ctx.Done():
			req.respChannel <- models.TrackResponse{
				Status: false,
				Error:  "request cancelled",
			}
			return
		default:
		}
	}

	if len(b.batch) == 0 {
		b.batchTimer.Reset(b.config.BatchTimeout)
	}

	b.batch = append(b.batch, batchItem{
		trackCode:    req.trackCode,
		responseChan: req.respChannel,
	})

	if len(b.batch) >= b.config.BatchSize {
		b.flushBatchLocked()
	}
}

func (b *Batcher) flushBatchLocked() {
	if len(b.batch) == 0 {
		return
	}

	batchToSend := make([]batchItem, len(b.batch))
	copy(batchToSend, b.batch)
	b.batch = b.batch[:0]

	if !b.batchTimer.Stop() {
		select {
		case <-b.batchTimer.C:
		default:
		}
	}

	select {
	case b.workerChan <- batchToSend:
	default:
		go b.processBatch(batchToSend)
	}
}

func (b *Batcher) worker(ctx context.Context) {
	for {
		select {
		case batch, ok := <-b.workerChan:
			if !ok {
				return
			}
			b.processBatch(batch)
		case <-ctx.Done():
			return
		}
	}
}

func (b *Batcher) processBatch(items []batchItem) {
	if len(items) == 0 {
		return
	}

	trackCodes := make([]string, 0, len(items))
	for _, item := range items {
		trackCodes = append(trackCodes, item.trackCode)
	}

	results, err := b.client.TrackPackagesBatch(context.Background(), trackCodes)
	if err != nil {
		log.Printf("batcher.processBatch.APIError: %v", err)

		var errorMsg string
		if erors.IsClientError(err) {
			errorMsg = err.Error()
		} else {
			errorMsg = "tracking service temporarily unavailable"
		}

		errorResponse := models.TrackResponse{
			Status: false,
			Error:  errorMsg,
		}

		for _, item := range items {
			select {
			case item.responseChan <- errorResponse:
			case <-time.After(100 * time.Millisecond):
				log.Printf("Client timeout for track code: %s", item.trackCode)
			}
		}
		return
	}

	for _, item := range items {
		if trackData, exists := results[item.trackCode]; exists {
			if err := b.cache.SetTrackData(context.Background(), item.trackCode, trackData, 5*time.Minute); err != nil {
				log.Printf("Cache set error for %s: %v", item.trackCode, err)
			}

			successResponse := models.TrackResponse{
				Status: true,
				Data:   trackData,
			}

			select {
			case item.responseChan <- successResponse:
			case <-time.After(100 * time.Millisecond):
				log.Printf("Client timeout for successful response: %s", item.trackCode)
			}

		} else {
			notFoundResponse := models.TrackResponse{
				Status: false,
				Error:  "tracking code not found in external system",
			}

			select {
			case item.responseChan <- notFoundResponse:
			case <-time.After(100 * time.Millisecond):
				log.Printf("Client timeout for not-found response: %s", item.trackCode)
			}

			log.Printf("No data found for track code: %s", item.trackCode)
		}
	}

	successful := 0
	for _, item := range items {
		if _, exists := results[item.trackCode]; exists {
			successful++
		}
	}

	log.Printf("Batch processing completed: %d/%d successful", successful, len(items))
}
