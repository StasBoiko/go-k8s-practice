package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/StasBoiko/go-k8s-practice/internal/model"
	"github.com/StasBoiko/go-k8s-practice/internal/repository"
)

// NoteHandler handles HTTP requests for notes.
type NoteHandler struct {
	repo repository.NoteRepository
}

// NewNoteHandler creates a new NoteHandler.
func NewNoteHandler(repo repository.NoteRepository) *NoteHandler {
	return &NoteHandler{repo: repo}
}

// Create creates a new note.
// @Summary Create a new note
// @Description Creates a new note record with the provided title and content
// @Tags notes
// @Accept json
// @Produce json
// @Param request body model.CreateNoteRequest true "Note creation payload"
// @Success 201 {object} model.Note
// @Failure 400 {object} model.ErrorResponse "Invalid input data or empty title"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /notes [post]
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		writeJSONError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}

	note, err := h.repo.Create(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(note)
}

// List retrieves all notes.
// @Summary List all notes
// @Description Returns an array of all saved notes
// @Tags notes
// @Produce json
// @Success 200 {array} model.Note
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /notes [get]
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	notes, err := h.repo.List(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(notes)
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}
