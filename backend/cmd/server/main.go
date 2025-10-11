package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/alerts"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/api"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/auth"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/config"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/logs"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/metrics"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/storage"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize JWT utilities with config
	utils.InitConfig(cfg)

	// Initialize database
	db, err := storage.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	authService := auth.NewService(db.GetDB())
	logAnalyzer := logs.NewLogAnalyzer()
	metricsCollector := metrics.NewCollector(db.GetDB(), cfg.Metrics.CollectionInterval)
	alertService := alerts.NewService(db.GetDB())

	// Initialize metric thresholds
	if err := metricsCollector.InitializeThresholds(); err != nil {
		log.Fatalf("Failed to initialize thresholds: %v", err)
	}

	// Initialize API handlers
	handlers := api.NewHandlers(authService, logAnalyzer, metricsCollector, alertService)

	// Setup Gin router
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	api.SetupRoutes(router, handlers, authService)

	// Start metrics collection in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting metrics collection...")
		metricsCollector.Start(ctx)
	}()

	// Start alert monitoring
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentMetrics, err := metricsCollector.GetCurrentMetrics()
				if err != nil {
					log.Printf("Failed to get current metrics for alert checking: %v", err)
					continue
				}

				if err := alertService.CheckThresholds(currentMetrics); err != nil {
					log.Printf("Failed to check alert thresholds: %v", err)
				}
			}
		}
	}()

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 CodeXray Observability Service starting on port %s", cfg.Server.Port)
		log.Printf("📊 Metrics collection interval: %v", cfg.Metrics.CollectionInterval)
		log.Printf("🔥 CPU threshold: %.1f%%", cfg.Metrics.CPUThreshold)
		log.Printf("💾 Memory threshold: %.1f%%", cfg.Metrics.MemoryThreshold)
		log.Printf("📁 Database: %s", cfg.GetDatabaseDSN())

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Cancel background processes
	cancel()

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}
