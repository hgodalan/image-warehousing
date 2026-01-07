package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourcompany/image-warehousing/internal/api"
	"github.com/yourcompany/image-warehousing/internal/config"
	"github.com/yourcompany/image-warehousing/internal/service"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting Image Warehousing Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	logger.Info("Configuration loaded successfully")

	// Initialize services
	logger.Info("Initializing services...")

	// Storage service
	storageService := service.NewStorageService(cfg.DataDir)
	if err := storageService.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize storage service: %v", err)
	}
	logger.Info("Storage service initialized")

	// Index service
	indexService := service.NewIndexService(cfg.DataDir)
	if err := indexService.InitializeIndex(); err != nil {
		logger.Fatalf("Failed to initialize index: %v", err)
	}
	logger.Info("Index service initialized")

	// AI service
	aiService, err := service.NewAIService(cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		logger.Fatalf("Failed to initialize AI service: %v", err)
	}
	defer aiService.Close()
	logger.Infof("AI service initialized (model: %s)", cfg.GeminiModel)

	// Image service (with workers)
	imageService := service.NewImageService(storageService, aiService, indexService, logger)
	imageService.StartWorkers(3) // Start 3 worker goroutines

	// Search service
	searchService := service.NewSearchService(indexService, aiService, logger)
	logger.Info("Search service initialized")

	// Create router
	router := api.NewRouter(cfg, storageService, imageService, searchService, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Server listening on port %s", cfg.ServerPort)
		logger.Infof("API base URL: http://localhost:%s/api/v1", cfg.ServerPort)
		logger.Info("Ready to accept requests!")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped gracefully")
}
