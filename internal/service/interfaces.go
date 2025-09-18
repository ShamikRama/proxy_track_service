package service

import (
	"context"

	"github.com/shamil/proxy_track_service-1/internal/config"
	"github.com/shamil/proxy_track_service-1/pkg/models"
)

type TrackingService interface {
	TrackPackage(ctx context.Context, trackCode string) <-chan models.TrackResponse
	Start(ctx context.Context) error
	Stop() error
	Health(ctx context.Context) error
}

type ServiceConfig struct {
	BatcherConfig config.BatcherConfig
	ClientConfig  config.ExternalConfig
}
