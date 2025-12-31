package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/repositories"
	"cold-backend/pkg/utils"
)

type LoginLogHandler struct {
	Repo    *repositories.LoginLogRepository
	OTPRepo *repositories.OTPRepository
}

func NewLoginLogHandler(repo *repositories.LoginLogRepository) *LoginLogHandler {
	return &LoginLogHandler{Repo: repo}
}

// SetOTPRepo sets the OTP repository for customer login logs
func (h *LoginLogHandler) SetOTPRepo(otpRepo *repositories.OTPRepository) {
	h.OTPRepo = otpRepo
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

// ListCustomerLoginLogs returns customer portal login logs (OTP verifications)
func (h *LoginLogHandler) ListCustomerLoginLogs(w http.ResponseWriter, r *http.Request) {
	if h.OTPRepo == nil {
		utils.RespondError(w, http.StatusInternalServerError, "Customer login logs not configured")
		return
	}

	logs, err := h.OTPRepo.GetLoginLogs(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to retrieve customer login logs")
		return
	}

	if logs == nil {
		logs = []repositories.CustomerLoginLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
