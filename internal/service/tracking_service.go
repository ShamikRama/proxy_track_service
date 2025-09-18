package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/shamil/proxy_track_service-1/internal/batcher"
	"github.com/shamil/proxy_track_service-1/internal/client"
	"github.com/shamil/proxy_track_service-1/internal/repository"
	"github.com/shamil/proxy_track_service-1/pkg/models"
)

type trackingService struct {
	batcher batcher.BatcherInterface
	cache   repository.CacheRepository
	client  client.ExternalAPIClient
	config  ServiceConfig

	mu     sync.RWMutex
	active bool
}

func NewTrackingService(
	config ServiceConfig,
	cache repository.CacheRepository,
	client client.ExternalAPIClient,
) TrackingService {
	batcherInstance := batcher.NewBatcher(config.BatcherConfig, cache, client)

	return &trackingService{
		batcher: batcherInstance,
		cache:   cache,
		client:  client,
		config:  config,
		active:  false,
	}
}

func (s *trackingService) TrackPackage(ctx context.Context, trackCode string) <-chan models.TrackResponse {
	s.mu.RLock()
	if !s.active {
		s.mu.RUnlock()
		errorChan := make(chan models.TrackResponse, 1)
		errorChan <- models.TrackResponse{
			Status: false,
			Error:  "service is not running",
		}
		return errorChan
	}
	s.mu.RUnlock()

	if cachedData, err := s.cache.GetTrackData(ctx, trackCode); err == nil && cachedData != nil {
		log.Printf("Cache hit for track code: %s", trackCode)
		cachedChan := make(chan models.TrackResponse, 1)
		cachedChan <- models.TrackResponse{
			Status: true,
			Data:   cachedData,
		}
		return cachedChan
	}

	log.Printf("Adding track code to batch: %s", trackCode)
	return s.batcher.AddRequest(ctx, trackCode)
}

func (s *trackingService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		return fmt.Errorf("service is already running")
	}

	if err := s.batcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start batcher: %w", err)
	}

	s.active = true
	log.Println("Tracking service started successfully")

	return nil
}

func (s *trackingService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return fmt.Errorf("service is not running")
	}

	if err := s.batcher.Stop(); err != nil {
		return fmt.Errorf("failed to stop batcher: %w", err)
	}

	s.active = false
	log.Println("Tracking service stopped successfully")

	return nil
}

func (s *trackingService) Health(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.active {
		return fmt.Errorf("service is not running")
	}

	if err := s.batcher.Health(ctx); err != nil {
		return fmt.Errorf("batcher health check failed: %w", err)
	}

	if err := s.cache.Health(ctx); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	return nil
}
