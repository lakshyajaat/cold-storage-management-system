package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
	"cold-backend/internal/services"

	"golang.org/x/crypto/bcrypt"
)

type PendingSettingHandler struct {
	pendingRepo       *repositories.PendingSettingChangeRepository
	systemSettingRepo *repositories.SystemSettingRepository
	userRepo          *repositories.UserRepository
	totpService       *services.TOTPService
}

func NewPendingSettingHandler(
	pendingRepo *repositories.PendingSettingChangeRepository,
	systemSettingRepo *repositories.SystemSettingRepository,
	userRepo *repositories.UserRepository,
	totpService *services.TOTPService,
) *PendingSettingHandler {
	return &PendingSettingHandler{
		pendingRepo:       pendingRepo,
		systemSettingRepo: systemSettingRepo,
		userRepo:          userRepo,
		totpService:       totpService,
	}
}

// RequestChange initiates a setting change request
// POST /api/admin/setting-changes
func (h *PendingSettingHandler) RequestChange(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.RequestSettingChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SettingKey == "" || req.NewValue == "" {
		http.Error(w, "Setting key and new value are required", http.StatusBadRequest)
		return
	}

	// Check if this is a protected setting
	isProtected, err := h.pendingRepo.IsProtectedSetting(r.Context(), req.SettingKey)
	if err != nil {
		log.Printf("[PendingSetting] Error checking protected: %v", err)
		http.Error(w, "Failed to check setting", http.StatusInternalServerError)
		return
	}

	if !isProtected {
		http.Error(w, "This setting does not require dual approval", http.StatusBadRequest)
		return
	}

	// Check if there's already a pending request for this setting
	existing, _ := h.pendingRepo.GetPendingBySettingKey(r.Context(), req.SettingKey)
	if existing != nil {
		http.Error(w, "There is already a pending request for this setting", http.StatusConflict)
		return
	}

	// Get current value
	oldValue := ""
	if setting, err := h.systemSettingRepo.Get(r.Context(), req.SettingKey); err == nil && setting != nil {
		oldValue = setting.SettingValue
	}

	// Create pending change
	change := &models.PendingSettingChange{
		SettingKey:  req.SettingKey,
		OldValue:    oldValue,
		NewValue:    req.NewValue,
		RequestedBy: userID,
		Reason:      req.Reason,
		Status:      models.PendingSettingStatusPending,
	}

	if err := h.pendingRepo.Create(r.Context(), change); err != nil {
		log.Printf("[PendingSetting] Error creating: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Setting change request created. Awaiting approval from another admin.",
		"change":  change,
	})
}

// GetPendingChanges returns all pending setting changes
// GET /api/admin/setting-changes/pending
func (h *PendingSettingHandler) GetPendingChanges(w http.ResponseWriter, r *http.Request) {
	// Expire old pending changes first
	_ = h.pendingRepo.ExpireOld(r.Context())

	changes, err := h.pendingRepo.GetAllPending(r.Context())
	if err != nil {
		log.Printf("[PendingSetting] Error getting pending: %v", err)
		http.Error(w, "Failed to get pending changes", http.StatusInternalServerError)
		return
	}

	// Mask sensitive values
	for _, change := range changes {
		if models.IsSensitiveSetting(change.SettingKey) {
			change.OldValue = models.MaskSensitiveValue(change.OldValue)
			change.NewValue = models.MaskSensitiveValue(change.NewValue)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(changes)
}

// GetChange returns a specific pending change
// GET /api/admin/setting-changes/{id}
func (h *PendingSettingHandler) GetChange(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		// Try URL query
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	change, err := h.pendingRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Change not found", http.StatusNotFound)
		return
	}

	// Mask sensitive values
	if models.IsSensitiveSetting(change.SettingKey) {
		change.OldValue = models.MaskSensitiveValue(change.OldValue)
		change.NewValue = models.MaskSensitiveValue(change.NewValue)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(change)
}

// ApproveChange approves a pending setting change
// POST /api/admin/setting-changes/{id}/approve
func (h *PendingSettingHandler) ApproveChange(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get the pending change
	change, err := h.pendingRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Change not found", http.StatusNotFound)
		return
	}

	if change.Status != models.PendingSettingStatusPending {
		http.Error(w, "Change has already been processed", http.StatusBadRequest)
		return
	}

	// Check that approver is different from requester
	if change.RequestedBy == userID {
		http.Error(w, "You cannot approve your own request. Another admin must approve.", http.StatusForbidden)
		return
	}

	// Parse and verify password
	var req models.ApproveSettingChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required for approval", http.StatusBadRequest)
		return
	}

	// Verify approver's password
	user, err := h.userRepo.Get(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// If approver has 2FA enabled, verify TOTP code
	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			http.Error(w, "2FA code is required for approval", http.StatusBadRequest)
			return
		}

		ipAddress := getIPAddress(r)
		valid, err := h.totpService.Verify(r.Context(), userID, req.TOTPCode, ipAddress)
		if err != nil {
			if _, ok := err.(*services.TOTPError); ok {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "2FA verification failed", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "Invalid 2FA code", http.StatusUnauthorized)
			return
		}
	}

	// Approve the change
	if err := h.pendingRepo.Approve(r.Context(), id, userID); err != nil {
		log.Printf("[PendingSetting] Error approving: %v", err)
		http.Error(w, "Failed to approve", http.StatusInternalServerError)
		return
	}

	// Get the full change to apply it
	change, _ = h.pendingRepo.GetByID(r.Context(), id)

	// Apply the setting change
	if err := h.systemSettingRepo.Update(r.Context(), change.SettingKey, change.NewValue, userID); err != nil {
		log.Printf("[PendingSetting] Error applying setting: %v", err)
		http.Error(w, "Failed to apply setting", http.StatusInternalServerError)
		return
	}

	log.Printf("[PendingSetting] Setting '%s' changed by user %d (approved by %d)", change.SettingKey, change.RequestedBy, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Setting change approved and applied",
	})
}

// RejectChange rejects a pending setting change
// POST /api/admin/setting-changes/{id}/reject
func (h *PendingSettingHandler) RejectChange(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.RejectSettingChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.pendingRepo.Reject(r.Context(), id, userID, req.Reason); err != nil {
		log.Printf("[PendingSetting] Error rejecting: %v", err)
		http.Error(w, "Failed to reject", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Setting change request rejected",
	})
}

// GetHistory returns recent setting change history
// GET /api/admin/setting-changes/history
func (h *PendingSettingHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	changes, err := h.pendingRepo.GetHistory(r.Context(), limit)
	if err != nil {
		log.Printf("[PendingSetting] Error getting history: %v", err)
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	// Mask sensitive values
	for _, change := range changes {
		if models.IsSensitiveSetting(change.SettingKey) {
			change.OldValue = models.MaskSensitiveValue(change.OldValue)
			change.NewValue = models.MaskSensitiveValue(change.NewValue)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(changes)
}

// GetProtectedSettings returns list of settings requiring dual approval
// GET /api/admin/setting-changes/protected
func (h *PendingSettingHandler) GetProtectedSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.pendingRepo.GetProtectedSettings(r.Context())
	if err != nil {
		log.Printf("[PendingSetting] Error getting protected: %v", err)
		http.Error(w, "Failed to get protected settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// CheckPendingForSetting checks if there's a pending change for a specific setting
// GET /api/admin/setting-changes/check?key=setting_key
func (h *PendingSettingHandler) CheckPendingForSetting(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Setting key is required", http.StatusBadRequest)
		return
	}

	change, err := h.pendingRepo.GetPendingBySettingKey(r.Context(), key)
	hasPending := err == nil && change != nil

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"has_pending": hasPending,
		"change":      change,
	})
}
