package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
	"cold-backend/internal/services"
)

type AuthHandler struct {
	Service        *services.UserService
	LoginLogRepo   *repositories.LoginLogRepository
}

func NewAuthHandler(s *services.UserService, loginLogRepo *repositories.LoginLogRepository) *AuthHandler {
	return &AuthHandler{
		Service:      s,
		LoginLogRepo: loginLogRepo,
	}
}

// Signup handles user registration
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authResp, err := h.Service.Signup(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResp)
}

// Login handles user authentication (step 1 - may require 2FA)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	loginResult, err := h.Service.Login(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Check if 2FA is required
	if loginResult.Requires2FA {
		json.NewEncoder(w).Encode(loginResult.Step1Response)
		return
	}

	// No 2FA - log the successful login and return full token
	ipAddress := getIPAddress(r)
	userAgent := r.UserAgent()
	if _, err := h.LoginLogRepo.CreateLoginLog(context.Background(), loginResult.AuthResponse.User.ID, ipAddress, userAgent); err != nil {
		// Log error but don't fail the login
	}

	json.NewEncoder(w).Encode(loginResult.AuthResponse)
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
