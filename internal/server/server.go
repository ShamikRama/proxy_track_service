package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shamil/proxy_track_service-1/internal/config"
	"github.com/shamil/proxy_track_service-1/internal/service"
)

type Server struct {
	httpServer *http.Server
	config     config.ServerConfig
}

func NewServer(config config.ServerConfig, trackingService service.TrackingService) *Server {
	router := SetupRoutes(trackingService)

	httpServer := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		config:     config,
	}
}

func (s *Server) Start() error {
	fmt.Printf("Starting server on port %s\n", s.config.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) StartTLS(certFile, keyFile string) error {
	fmt.Printf("Starting TLS server on port %s\n", s.config.Port)
	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Stopping server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}
