package models

import "time"

type PendingSettingChangeStatus string

const (
	PendingSettingStatusPending  PendingSettingChangeStatus = "pending"
	PendingSettingStatusApproved PendingSettingChangeStatus = "approved"
	PendingSettingStatusRejected PendingSettingChangeStatus = "rejected"
	PendingSettingStatusExpired  PendingSettingChangeStatus = "expired"
)

type PendingSettingChange struct {
	ID              int                        `json:"id"`
	SettingKey      string                     `json:"setting_key"`
	OldValue        string                     `json:"old_value,omitempty"`
	NewValue        string                     `json:"new_value"`
	RequestedBy     int                        `json:"requested_by"`
	RequestedByName string                     `json:"requested_by_name,omitempty"`
	RequestedAt     time.Time                  `json:"requested_at"`
	Reason          string                     `json:"reason,omitempty"`
	ApprovedBy      *int                       `json:"approved_by,omitempty"`
	ApprovedByName  string                     `json:"approved_by_name,omitempty"`
	ApprovedAt      *time.Time                 `json:"approved_at,omitempty"`
	Status          PendingSettingChangeStatus `json:"status"`
	RejectionReason string                     `json:"rejection_reason,omitempty"`
	ExpiresAt       time.Time                  `json:"expires_at"`
}

type RequestSettingChangeRequest struct {
	SettingKey string `json:"setting_key"`
	NewValue   string `json:"new_value"`
	Reason     string `json:"reason,omitempty"`
}

type ApproveSettingChangeRequest struct {
	Password string `json:"password"`
}

type RejectSettingChangeRequest struct {
	Reason string `json:"reason"`
}

// ProtectedSetting represents a setting that requires dual admin approval
type ProtectedSetting struct {
	SettingKey  string `json:"setting_key"`
	Description string `json:"description"`
}

// IsSensitiveSetting checks if a setting key contains sensitive data that should be masked
func IsSensitiveSetting(key string) bool {
	sensitiveKeys := map[string]bool{
		"razorpay_key_secret":     true,
		"razorpay_webhook_secret": true,
	}
	return sensitiveKeys[key]
}

// MaskSensitiveValue masks sensitive values for display
func MaskSensitiveValue(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}
