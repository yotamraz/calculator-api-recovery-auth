package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load configuration from environment variables.
	cfg := LoadConfig()

	// Initialize SQLite database via GORM.
	db, err := gorm.Open(sqlite.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate database schema.
	if err := db.AutoMigrate(&User{}, &Calculation{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Build shared dependencies.
	deps := &Deps{
		DB:     db,
		Config: cfg,
	}

	// Set up Chi router.
	r := chi.NewRouter()

	// Public routes.
	r.Get("/health", deps.HealthHandler)

	// Protected routes will be added in future milestones.

	// Create HTTP server.
	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	// Graceful shutdown: listen for interrupt signals in a goroutine.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("server starting on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until signal received.
	<-done
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server stopped")
}
