package models

import "time"

// TOTPSetupResponse returned when initiating 2FA setup
type TOTPSetupResponse struct {
	Secret      string `json:"secret"`       // Base32 secret for manual entry
	QRCode      string `json:"qr_code"`      // Base64 encoded PNG QR code
	Issuer      string `json:"issuer"`       // "ColdStorage"
	AccountName string `json:"account_name"` // User's email
}

// TOTPEnableRequest to verify and enable 2FA
type TOTPEnableRequest struct {
	Code string `json:"code"` // 6-digit TOTP code
}

// TOTPVerifyRequest for login 2FA verification
type TOTPVerifyRequest struct {
	TempToken string `json:"temp_token"` // Temporary token from step 1
	Code      string `json:"code"`       // 6-digit TOTP code or backup code
}

// TOTPDisableRequest to disable 2FA
type TOTPDisableRequest struct {
	Password string `json:"password"` // User's password for verification
	Code     string `json:"code"`     // Current TOTP code
}

// LoginStep1Response when 2FA is required after password verification
type LoginStep1Response struct {
	Requires2FA bool   `json:"requires_2fa"`
	TempToken   string `json:"temp_token,omitempty"` // Short-lived token for step 2
	Message     string `json:"message,omitempty"`
}

// BackupCodesResponse returned after generating backup codes
type BackupCodesResponse struct {
	Codes []string `json:"codes"` // Plaintext codes (shown once, user must save)
}

// User2FAStatus for user profile/settings
type User2FAStatus struct {
	Enabled        bool       `json:"enabled"`
	EnabledAt      *time.Time `json:"enabled_at,omitempty"`
	HasBackupCodes bool       `json:"has_backup_codes"`
}

// RegenerateBackupCodesRequest requires password verification
type RegenerateBackupCodesRequest struct {
	Password string `json:"password"`
}

// TOTPVerificationAttempt for rate limiting
type TOTPVerificationAttempt struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}
