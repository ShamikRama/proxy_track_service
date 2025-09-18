package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/shamil/proxy_track_service-1/internal/models"
	"github.com/shamil/proxy_track_service-1/internal/service"
)

type TrackHandler struct {
	trackingService service.TrackingService
}

func NewTrackHandler(trackingService service.TrackingService) *TrackHandler {
	return &TrackHandler{
		trackingService: trackingService,
	}
}

func (h *TrackHandler) GetTrackStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "only GET method is supported")
		return
	}

	vars := mux.Vars(r)
	trackCode := strings.TrimSpace(vars["trackCode"])

	if trackCode == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "track_code is required")
		return
	}

	responseChan := h.trackingService.TrackPackage(r.Context(), trackCode)

	select {
	case response := <-responseChan:
		if !response.Status && response.Error != "" {
			statusCode := h.getStatusCodeFromError(response.Error)
			h.writeErrorResponse(w, statusCode, response.Error)
			return
		}
		h.writeJSONResponse(w, http.StatusOK, response)
	case <-r.Context().Done():
		h.writeErrorResponse(w, http.StatusRequestTimeout, "request cancelled by client")
		return
	}
}

func (h *TrackHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "only GET method is supported")
		return
	}

	err := h.trackingService.Health(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "service unhealthy: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"status":  true,
		"message": "service is healthy",
		"service": "proxy_track_service",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *TrackHandler) getStatusCodeFromError(errorMsg string) int {
	switch {
	case strings.Contains(errorMsg, "tracking code not found"):
		return http.StatusNotFound
	case strings.Contains(errorMsg, "invalid tracking code format"):
		return http.StatusBadRequest
	case strings.Contains(errorMsg, "tracking service temporarily unavailable"):
		return http.StatusServiceUnavailable
	case strings.Contains(errorMsg, "request timeout"):
		return http.StatusRequestTimeout
	case strings.Contains(errorMsg, "too many requests"):
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func (h *TrackHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
	}
}

func (h *TrackHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := models.TrackResponse{
		Status: false,
		Error:  message,
	}

	h.writeJSONResponse(w, statusCode, response)
}
