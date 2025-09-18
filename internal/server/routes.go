package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shamil/proxy_track_service-1/internal/handler"
	"github.com/shamil/proxy_track_service-1/internal/service"
)

func SetupRoutes(trackingService service.TrackingService) *mux.Router {
	router := mux.NewRouter()
	router.Use(handler.LoggingMiddleware)
	router.Use(handler.CORSMiddleware)
	router.Use(handler.RecoveryMiddleware)

	trackHandler := handler.NewTrackHandler(trackingService)

	setupAPIRoutes(router, trackHandler)
	setupGeneralRoutes(router)

	return router
}

func setupAPIRoutes(router *mux.Router, trackHandler *handler.TrackHandler) {
	router.HandleFunc("/track/{trackCode}", trackHandler.GetTrackStatus).Methods("GET")
	router.HandleFunc("/health", trackHandler.HealthCheck).Methods("GET")
}

func setupGeneralRoutes(router *mux.Router) {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"message": "Proxy Track Service",
			"version": "1.0.0",
			"endpoints": {
				"track": "GET /track/{trackCode}",
				"health": "GET /health"
			}
		}`)
	}).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{
			"error": "not found",
			"path": "%s",
			"available_endpoints": [
				"GET /track/{trackCode}",
				"GET /health"
			]
		}`, r.URL.Path)
	})
}
