package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cold-backend/internal/cache"
	"cold-backend/internal/middleware"
	"cold-backend/internal/models"
	"cold-backend/internal/services"

	"github.com/gorilla/mux"
)

const settingsCacheTTL = 24 * time.Hour

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
	ctx := r.Context()
	cacheKey := "settings:list"

	// Try cache first
	if data, ok := cache.GetCached(ctx, cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(data)
		return
	}

	settings, err := h.Service.ListSettings(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the response
	data, _ := json.Marshal(settings)
	cache.SetCached(ctx, cacheKey, data, settingsCacheTTL)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(data)
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

	// Invalidate settings cache
	cache.InvalidateSettingCaches(r.Context())

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Setting updated successfully"})
}

// SkipRange represents a range of thock numbers to skip
type SkipRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// SkipRangesRequest represents the request body for skip ranges
type SkipRangesRequest struct {
	SeedRanges []SkipRange `json:"seed_ranges"`
	SellRanges []SkipRange `json:"sell_ranges"`
}

// GetSkipThockRanges returns the skip ranges for both categories
func (h *SystemSettingHandler) GetSkipThockRanges(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get seed ranges
	seedSetting, _ := h.Service.GetSetting(ctx, "skip_thock_ranges_seed")
	sellSetting, _ := h.Service.GetSetting(ctx, "skip_thock_ranges_sell")

	var seedRanges, sellRanges []SkipRange

	if seedSetting != nil && seedSetting.SettingValue != "" {
		json.Unmarshal([]byte(seedSetting.SettingValue), &seedRanges)
	}
	if sellSetting != nil && sellSetting.SettingValue != "" {
		json.Unmarshal([]byte(sellSetting.SettingValue), &sellRanges)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"seed_ranges": seedRanges,
		"sell_ranges": sellRanges,
	})
}

// UpdateSkipThockRanges updates the skip ranges for both categories
func (h *SystemSettingHandler) UpdateSkipThockRanges(w http.ResponseWriter, r *http.Request) {
	var req SkipRangesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Save seed ranges
	seedData, _ := json.Marshal(req.SeedRanges)
	if err := h.Service.UpsertSetting(ctx, "skip_thock_ranges_seed", string(seedData), "Thock number ranges to skip for SEED category", userID); err != nil {
		http.Error(w, "Failed to save seed ranges: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save sell ranges
	sellData, _ := json.Marshal(req.SellRanges)
	if err := h.Service.UpsertSetting(ctx, "skip_thock_ranges_sell", string(sellData), "Thock number ranges to skip for SELL category", userID); err != nil {
		http.Error(w, "Failed to save sell ranges: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Invalidate settings cache
	cache.InvalidateSettingCaches(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Skip ranges updated successfully"})
}

// GetOperationMode returns the current system operation mode
func (h *SystemSettingHandler) GetOperationMode(w http.ResponseWriter, r *http.Request) {
	// Try to get from database, fallback to default
	setting, err := h.Service.GetSetting(context.Background(), "operation_mode")

	mode := "loading" // Default to loading mode
	message := "System is in loading mode - items being stored"

	if err == nil && setting != nil {
		mode = setting.SettingValue
		switch mode {
		case "loading":
			message = "System is in loading mode - items being stored"
		case "unloading":
			message = "System is in unloading mode - items being dispatched"
		case "maintenance":
			message = "System is in maintenance mode"
		case "readonly":
			message = "System is in read-only mode"
		case "emergency":
			message = "System is in emergency mode"
		default:
			mode = "loading"
			message = "System is in loading mode - items being stored"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mode":    mode,
		"message": message,
	})
}
