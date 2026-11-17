package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StasBoiko/go-k8s-practice/internal/handler"
	"github.com/StasBoiko/go-k8s-practice/internal/model"
)

type mockNoteRepository struct {
	createFunc func(ctx context.Context, req model.CreateNoteRequest) (*model.Note, error)
	listFunc   func(ctx context.Context) ([]model.Note, error)
}

func (m *mockNoteRepository) Create(ctx context.Context, req model.CreateNoteRequest) (*model.Note, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockNoteRepository) List(ctx context.Context) ([]model.Note, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func TestNoteHandler_Create_Success(t *testing.T) {
	mockRepo := &mockNoteRepository{
		createFunc: func(_ context.Context, req model.CreateNoteRequest) (*model.Note, error) {
			return &model.Note{
				ID:        1,
				Title:     req.Title,
				Content:   req.Content,
				CreatedAt: time.Date(2026, 11, 17, 6, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	h := handler.NewNoteHandler(mockRepo)

	payload := model.CreateNoteRequest{
		Title:   "Test Note",
		Content: "This is a test content",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, rr.Code)
	}

	var created model.Note
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.ID != 1 || created.Title != payload.Title || created.Content != payload.Content {
		t.Errorf("unexpected note returned: %+v", created)
	}
}

func TestNoteHandler_Create_ValidationError_EmptyTitle(t *testing.T) {
	mockRepo := &mockNoteRepository{}
	h := handler.NewNoteHandler(mockRepo)

	payload := model.CreateNoteRequest{
		Title:   "   ", // whitespace only
		Content: "Content without title",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var errResp model.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if errResp.Error != "title cannot be empty" {
		t.Errorf("expected error 'title cannot be empty', got '%s'", errResp.Error)
	}
}

func TestNoteHandler_List_Success(t *testing.T) {
	expectedNotes := []model.Note{
		{
			ID:        1,
			Title:     "Note 1",
			Content:   "Content 1",
			CreatedAt: time.Date(2026, 11, 17, 5, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Title:     "Note 2",
			Content:   "Content 2",
			CreatedAt: time.Date(2026, 11, 17, 6, 0, 0, 0, time.UTC),
		},
	}

	mockRepo := &mockNoteRepository{
		listFunc: func(_ context.Context) ([]model.Note, error) {
			return expectedNotes, nil
		},
	}

	h := handler.NewNoteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var notes []model.Note
	if err := json.Unmarshal(rr.Body.Bytes(), &notes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}

	if notes[0].Title != expectedNotes[0].Title || notes[1].Title != expectedNotes[1].Title {
		t.Errorf("notes mismatch: %+v", notes)
	}
}
