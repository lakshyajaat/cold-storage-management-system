package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"cold-backend/internal/middleware"
	"cold-backend/internal/models"
	"cold-backend/internal/services"
	"github.com/gorilla/mux"
)

type SystemSettingHandler struct {
	Service *services.SystemSettingService
}

func NewSystemSettingHandler(service *services.SystemSettingService) *SystemSettingHandler {
	return &SystemSettingHandler{Service: service}
}

func (h *SystemSettingHandler) GetSetting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	setting, err := h.Service.GetSetting(context.Background(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setting)
}

func (h *SystemSettingHandler) ListSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Service.ListSettings(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *SystemSettingHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var req models.UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if err := h.Service.UpdateSetting(context.Background(), key, req.SettingValue, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Setting updated successfully"})
}
