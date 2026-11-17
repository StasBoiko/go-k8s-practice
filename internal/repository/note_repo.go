// Package repository provides data storage interfaces and database implementations.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/StasBoiko/go-k8s-practice/internal/model"
)

// NoteRepository defines methods for accessing and managing notes data.
type NoteRepository interface {
	Create(ctx context.Context, req model.CreateNoteRequest) (*model.Note, error)
	List(ctx context.Context) ([]model.Note, error)
}

// PostgresNoteRepository provides a PostgreSQL implementation of NoteRepository.
type PostgresNoteRepository struct {
	db *sql.DB
}

// NewPostgresNoteRepository creates a new PostgresNoteRepository instance.
func NewPostgresNoteRepository(db *sql.DB) *PostgresNoteRepository {
	return &PostgresNoteRepository{db: db}
}

// Create inserts a new note record into PostgreSQL and returns the created note.
func (r *PostgresNoteRepository) Create(ctx context.Context, req model.CreateNoteRequest) (*model.Note, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is not connected")
	}

	query := `INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id, title, content, created_at`

	var note model.Note
	err := r.db.QueryRowContext(ctx, query, req.Title, req.Content).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return &note, nil
}

// List returns all notes from PostgreSQL ordered by newest first.
func (r *PostgresNoteRepository) List(ctx context.Context) ([]model.Note, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is not connected")
	}

	query := `SELECT id, title, content, created_at FROM notes ORDER BY id DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	notes := make([]model.Note, 0)
	for rows.Next() {
		var note model.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error during notes iteration: %w", err)
	}

	return notes, nil
}
