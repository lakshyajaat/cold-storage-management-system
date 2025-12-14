package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/repositories"
	"cold-backend/pkg/utils"
)

type AdminActionLogHandler struct {
	Repo *repositories.AdminActionLogRepository
}

func NewAdminActionLogHandler(repo *repositories.AdminActionLogRepository) *AdminActionLogHandler {
	return &AdminActionLogHandler{Repo: repo}
}

// ListActionLogs returns all admin action logs
func (h *AdminActionLogHandler) ListActionLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Repo.ListAllActionLogs(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to retrieve admin action logs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
