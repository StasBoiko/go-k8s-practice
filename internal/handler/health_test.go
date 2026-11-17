package handler_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/StasBoiko/go-k8s-practice/internal/config"
	"github.com/StasBoiko/go-k8s-practice/internal/handler"
)

func TestHealthCheck_DegradedWhenDBUnavailable(t *testing.T) {
	cfg := &config.Config{
		AppEnv:     "test",
		AppVersion: "0.1.0-test",
		HTTPPort:   "8080",
	}

	// Passing nil *sql.DB simulates an unavailable database
	h := handler.NewHealthHandler(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.Check(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}

	var resp handler.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", resp.Status)
	}

	if resp.DBStatus != "down" {
		t.Errorf("expected db_status 'down', got '%s'", resp.DBStatus)
	}

	if resp.Env != "test" {
		t.Errorf("expected env 'test', got '%s'", resp.Env)
	}

	if resp.Version != "0.1.0-test" {
		t.Errorf("expected version '0.1.0-test', got '%s'", resp.Version)
	}
}

func TestHealthCheck_SuccessWhenDBAvailable(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("skipping integration test: DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	cfg := &config.Config{
		AppEnv:      "test",
		AppVersion:  "0.1.0-test",
		HTTPPort:    "8080",
		DatabaseURL: dbURL,
	}

	h := handler.NewHealthHandler(cfg, db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.Check(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var resp handler.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}

	if resp.DBStatus != "up" {
		t.Errorf("expected db_status 'up', got '%s'", resp.DBStatus)
	}
}
