package batcher

import (
	"context"

	"github.com/shamil/proxy_track_service-1/internal/models"
)

type batchItem struct {
	trackCode    string
	responseChan chan models.TrackResponse
}

type batchRequest struct {
	ctx         context.Context
	trackCode   string
	respChannel chan models.TrackResponse
}
