package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shamil/proxy_track_service-1/internal/client/fourpx"
	"github.com/shamil/proxy_track_service-1/internal/config"
	"github.com/shamil/proxy_track_service-1/internal/repository"
	"github.com/shamil/proxy_track_service-1/internal/server"
	"github.com/shamil/proxy_track_service-1/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("Starting proxy tracking service...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache, err := repository.NewRedisCache(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
	defer cache.Close()

	externalClient := fourpx.NewFourPXClient(cfg.External.BaseURL, cfg.External.HashPattern, cfg.External.Timeout)

	serviceConfig := service.ServiceConfig{
		BatcherConfig: cfg.Batcher,
		ClientConfig:  cfg.External,
	}

	trackingService := service.NewTrackingService(serviceConfig, cache, externalClient)

	if err := trackingService.Start(ctx); err != nil {
		log.Fatalf("Failed to start tracking service: %v", err)
	}
	defer trackingService.Stop()

	router := server.SetupRoutes(trackingService)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
