package repository

import (
	"context"
	"time"

	"github.com/shamil/proxy_track_service-1/pkg/models"
)

type CacheRepository interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetTrackData(ctx context.Context, trackCode string) (*models.TrackData, error)
	SetTrackData(ctx context.Context, trackCode string, data *models.TrackData, ttl time.Duration) error
	Health(ctx context.Context) error
	Close() error
}
