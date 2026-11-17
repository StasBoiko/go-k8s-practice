// Package model contains core domain data structures and request/response models.
package model

import "time"

// Note represents a user note in the system.
type Note struct {
	ID        int       `json:"id" example:"1"`
	Title     string    `json:"title" example:"My first note"`
	Content   string    `json:"content" example:"Hello, Kubernetes!"`
	CreatedAt time.Time `json:"created_at" example:"2026-11-17T06:00:00Z"`
}

// CreateNoteRequest defines the payload required to create a new note.
type CreateNoteRequest struct {
	Title   string `json:"title" example:"My first note"`
	Content string `json:"content" example:"Hello, Kubernetes!"`
}

// ErrorResponse represents a standard error response payload.
type ErrorResponse struct {
	Error string `json:"error" example:"title is required"`
}
