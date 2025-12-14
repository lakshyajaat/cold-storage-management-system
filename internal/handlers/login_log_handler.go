package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/repositories"
	"cold-backend/pkg/utils"
)

type LoginLogHandler struct {
	Repo *repositories.LoginLogRepository
}

func NewLoginLogHandler(repo *repositories.LoginLogRepository) *LoginLogHandler {
	return &LoginLogHandler{Repo: repo}
}

// ListLoginLogs returns all login/logout logs
func (h *LoginLogHandler) ListLoginLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Repo.ListAllLoginLogs(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to retrieve login logs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// Logout handles user logout and records logout time
func (h *LoginLogHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Update logout time for the user's most recent login
	if err := h.Repo.UpdateLogoutTimeByUser(context.Background(), userID); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to record logout")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout recorded successfully",
	})
}
