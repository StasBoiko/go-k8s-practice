package migrations_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/StasBoiko/go-k8s-practice/migrations"
)

func TestMigrations_Run(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("skipping migration test: DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := migrations.Run(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
}
