package handlers

import (
	"encoding/json"
	"net/http"

	"crm/middleware"
	"crm/models"
	"crm/repositories"
	"crm/utils"
)

// NoteHandler handles note-related HTTP requests
type NoteHandler struct {
	noteRepo  *repositories.NoteRepository
	auditRepo *repositories.AuditRepository
}

// NewNoteHandler creates a new NoteHandler
func NewNoteHandler(noteRepo *repositories.NoteRepository, auditRepo *repositories.AuditRepository) *NoteHandler {
	return &NoteHandler{noteRepo: noteRepo, auditRepo: auditRepo}
}

// ListByEntity handles GET /api/notes?entity_type=lead&entity_id=xxx
func (h *NoteHandler) ListByEntity(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	if entityType == "" || entityID == "" {
		utils.BadRequest(w, "entity_type and entity_id are required")
		return
	}

	notes, err := h.noteRepo.ListByEntity(entityType, entityID)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch notes")
		return
	}
	if notes == nil {
		notes = []*models.Note{}
	}
	utils.Success(w, notes)
}

// Create handles POST /api/notes
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Body == "" {
		utils.BadRequest(w, "Note body is required")
		return
	}

	claims := middleware.GetClaims(r)
	ownerID := ""
	if claims != nil {
		ownerID = claims.UserID
	}

	note, err := h.noteRepo.Create(&req, ownerID)
	if err != nil {
		utils.InternalServerError(w, "Failed to create note")
		return
	}

	utils.Created(w, note)
}

// Update handles PUT /api/notes/:id
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/notes/")

	var req models.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	note, err := h.noteRepo.Update(id, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update note")
		return
	}

	utils.Success(w, note)
}

// Delete handles DELETE /api/notes/:id
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/notes/")

	if err := h.noteRepo.SoftDelete(id); err != nil {
		utils.InternalServerError(w, "Failed to delete note")
		return
	}

	utils.Success(w, map[string]string{"message": "Note deleted"})
}
