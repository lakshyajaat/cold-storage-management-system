package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/repositories"
	"cold-backend/pkg/utils"
)

type RoomEntryEditLogHandler struct {
	Repo *repositories.RoomEntryEditLogRepository
}

func NewRoomEntryEditLogHandler(repo *repositories.RoomEntryEditLogRepository) *RoomEntryEditLogHandler {
	return &RoomEntryEditLogHandler{Repo: repo}
}

// ListEditLogs returns all room entry edit logs
func (h *RoomEntryEditLogHandler) ListEditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Repo.ListAllEditLogs(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to retrieve edit logs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
