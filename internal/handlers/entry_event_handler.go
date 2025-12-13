package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/middleware"
	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
)

type EntryEventHandler struct {
	EntryEventRepo *repositories.EntryEventRepository
}

func NewEntryEventHandler(repo *repositories.EntryEventRepository) *EntryEventHandler {
	return &EntryEventHandler{EntryEventRepo: repo}
}

func (h *EntryEventHandler) CreateEntryEvent(w http.ResponseWriter, r *http.Request) {
	var req models.CreateEntryEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	event := &models.EntryEvent{
		EntryID:         req.EntryID,
		EventType:       req.EventType,
		Status:          req.Status,
		Notes:           req.Notes,
		CreatedByUserID: userID,
	}

	if err := h.EntryEventRepo.Create(context.Background(), event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}
