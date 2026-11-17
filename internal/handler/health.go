// Package handler provides HTTP handlers for the application.
package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/StasBoiko/go-k8s-practice/internal/config"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	cfg *config.Config
	db  *sql.DB
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(cfg *config.Config, db *sql.DB) *HealthHandler {
	return &HealthHandler{
		cfg: cfg,
		db:  db,
	}
}

// HealthResponse represents the response body for health checks.
type HealthResponse struct {
	Status   string `json:"status" example:"ok"`
	Env      string `json:"env" example:"local"`
	Version  string `json:"version" example:"0.1.0"`
	DBStatus string `json:"db_status" example:"up"`
}

// Check returns the system health status.
// @Summary Health check
// @Description Returns the system health status, environment, version, and database connectivity status
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	dbStatus := "up"
	httpStatusCode := http.StatusOK

	if h.db == nil {
		status = "degraded"
		dbStatus = "down"
		httpStatusCode = http.StatusServiceUnavailable
	} else {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			status = "degraded"
			dbStatus = "down"
			httpStatusCode = http.StatusServiceUnavailable
		}
	}

	resp := HealthResponse{
		Status:   status,
		Env:      h.cfg.AppEnv,
		Version:  h.cfg.AppVersion,
		DBStatus: dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
