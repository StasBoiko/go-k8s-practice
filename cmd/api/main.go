// Package main is the entry point for the API service.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/StasBoiko/go-k8s-practice/docs"
	"github.com/StasBoiko/go-k8s-practice/internal/config"
	"github.com/StasBoiko/go-k8s-practice/internal/handler"
	"github.com/StasBoiko/go-k8s-practice/internal/repository"
	"github.com/StasBoiko/go-k8s-practice/migrations"
)

// @title Go K8s Practice API
// @version 0.2.0
// @description REST API service for Kubernetes practice project.
// @BasePath /
func main() {
	// Initialize structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()
	slog.Info("configuration loaded",
		"env", cfg.AppEnv,
		"version", cfg.AppVersion,
		"port", cfg.HTTPPort,
	)

	// Connect to PostgreSQL (if configured)
	var db *sql.DB
	if cfg.DatabaseURL != "" {
		var err error
		db, err = sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			slog.Error("failed to open database connection", "error", err)
			os.Exit(1)
		}
		defer func() {
			_ = db.Close()
		}()

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := db.PingContext(ctx); err != nil {
			slog.Warn("database ping failed on startup, running in degraded state", "error", err)
		} else {
			slog.Info("connected to PostgreSQL successfully")
			if err := migrations.Run(db); err != nil {
				slog.Error("failed to run database migrations", "error", err)
				os.Exit(1)
			}
			slog.Info("database migrations applied successfully")
		}
		cancel()
	} else {
		slog.Warn("DATABASE_URL is not set; running without database connection")
	}

	// Setup HTTP router
	mux := http.NewServeMux()

	healthHandler := handler.NewHealthHandler(cfg, db)
	mux.HandleFunc("GET /health", healthHandler.Check)

	noteRepo := repository.NewPostgresNoteRepository(db)
	noteHandler := handler.NewNoteHandler(noteRepo)
	mux.HandleFunc("POST /notes", noteHandler.Create)
	mux.HandleFunc("GET /notes", noteHandler.List)

	// Swagger documentation routes
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

	// Setup HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		slog.Info("HTTP server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("received termination signal, initiating graceful shutdown", "signal", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	} else {
		slog.Info("server exited cleanly")
	}
}
