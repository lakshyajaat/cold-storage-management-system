package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/auth"
	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
	"cold-backend/internal/services"
)

type TOTPHandler struct {
	TOTPService *services.TOTPService
	UserRepo    *repositories.UserRepository
	JWTManager  *auth.JWTManager
}

func NewTOTPHandler(totpService *services.TOTPService, userRepo *repositories.UserRepository, jwtManager *auth.JWTManager) *TOTPHandler {
	return &TOTPHandler{
		TOTPService: totpService,
		UserRepo:    userRepo,
		JWTManager:  jwtManager,
	}
}

// SetupTOTP initiates 2FA setup - returns secret and QR code
func (h *TOTPHandler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	// Only admins can set up 2FA
	if claims.Role != "admin" {
		http.Error(w, "Only admin users can enable 2FA", http.StatusForbidden)
		return
	}

	user, err := h.UserRepo.Get(context.Background(), claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if already enabled
	if user.TOTPEnabled {
		http.Error(w, "2FA is already enabled", http.StatusBadRequest)
		return
	}

	response, err := h.TOTPService.GenerateSetup(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to generate 2FA setup", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EnableTOTP verifies the code and enables 2FA - returns backup codes
func (h *TOTPHandler) EnableTOTP(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	var req models.TOTPEnableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, "Verification code is required", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	response, err := h.TOTPService.VerifyAndEnable(context.Background(), claims.UserID, req.Code, ipAddress)
	if err != nil {
		if _, ok := err.(*services.TOTPError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to enable 2FA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DisableTOTP turns off 2FA after verifying password and code
func (h *TOTPHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	var req models.TOTPDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" || req.Code == "" {
		http.Error(w, "Password and verification code are required", http.StatusBadRequest)
		return
	}

	err := h.TOTPService.Disable(context.Background(), claims.UserID, req.Password, req.Code)
	if err != nil {
		if _, ok := err.(*services.TOTPError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to disable 2FA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "2FA disabled successfully"})
}

// GetStatus returns the 2FA status for the current user
func (h *TOTPHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	status, err := h.TOTPService.GetStatus(context.Background(), claims.UserID)
	if err != nil {
		http.Error(w, "Failed to get 2FA status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// RegenerateBackupCodes creates new backup codes (requires password)
func (h *TOTPHandler) RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	var req models.RegenerateBackupCodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	response, err := h.TOTPService.RegenerateBackupCodes(context.Background(), claims.UserID, req.Password)
	if err != nil {
		if _, ok := err.(*services.TOTPError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to regenerate backup codes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyTOTP handles 2FA verification during login (step 2)
func (h *TOTPHandler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	var req models.TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TempToken == "" || req.Code == "" {
		http.Error(w, "Temp token and verification code are required", http.StatusBadRequest)
		return
	}

	// Validate temp token
	tempClaims, err := h.JWTManager.ValidateTempToken(req.TempToken)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.UserRepo.Get(context.Background(), tempClaims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify TOTP code
	ipAddress := getIPAddress(r)
	valid, err := h.TOTPService.Verify(context.Background(), user.ID, req.Code, ipAddress)
	if err != nil {
		if _, ok := err.(*services.TOTPError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Verification failed", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid verification code", http.StatusUnauthorized)
		return
	}

	// Generate full JWT token
	token, err := h.JWTManager.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := &models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
