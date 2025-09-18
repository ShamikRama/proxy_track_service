package client

import (
	"context"

	"github.com/shamil/proxy_track_service-1/internal/models"
)

type ExternalAPIClient interface {
	TrackPackage(ctx context.Context, trackCode string) (*models.TrackData, error)
	TrackPackagesBatch(ctx context.Context, trackCodes []string) (map[string]*models.TrackData, error)
}
