package batcher

import (
	"context"

	"github.com/shamil/proxy_track_service-1/pkg/models"
)

type BatcherInterface interface {
	AddRequest(ctx context.Context, trackCode string) <-chan models.TrackResponse
	Start(ctx context.Context) error
	Stop() error
	Health(ctx context.Context) error
	Flush()

	addToBatch(req batchRequest)
	flushBatchLocked()
	processBatch(items []batchItem)
	worker(ctx context.Context)
}
