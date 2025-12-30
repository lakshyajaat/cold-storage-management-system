package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"time"

	"cold-backend/internal/models"
	"cold-backend/internal/repositories"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	issuer            = "ColdStorage"
	backupCodeCount   = 10
	backupCodeLength  = 8
	maxFailedAttempts = 5
	rateLimitWindow   = 15 * time.Minute
)

type TOTPService struct {
	userRepo *repositories.UserRepository
	totpRepo *repositories.TOTPRepository
}

func NewTOTPService(userRepo *repositories.UserRepository, totpRepo *repositories.TOTPRepository) *TOTPService {
	return &TOTPService{
		userRepo: userRepo,
		totpRepo: totpRepo,
	}
}

// GenerateSetup creates a new TOTP secret and QR code for a user
func (s *TOTPService) GenerateSetup(ctx context.Context, user *models.User) (*models.TOTPSetupResponse, error) {
	// Generate new TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.Email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, err
	}

	// Store the secret (not yet enabled)
	err = s.userRepo.SetTOTPSecret(ctx, user.ID, key.Secret())
	if err != nil {
		return nil, err
	}

	// Generate QR code image
	qrImage, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}

	// Convert to base64 PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, qrImage)
	if err != nil {
		return nil, err
	}
	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &models.TOTPSetupResponse{
		Secret:      key.Secret(),
		QRCode:      "data:image/png;base64," + qrBase64,
		Issuer:      issuer,
		AccountName: user.Email,
	}, nil
}

// VerifyAndEnable verifies a TOTP code and enables 2FA for the user
func (s *TOTPService) VerifyAndEnable(ctx context.Context, userID int, code string, ipAddress string) (*models.BackupCodesResponse, error) {
	// Check rate limiting
	if exceeded, err := s.isRateLimited(ctx, userID, ipAddress); err != nil {
		return nil, err
	} else if exceeded {
		return nil, ErrTooManyAttempts
	}

	// Get user with TOTP secret
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.TOTPSecret == "" {
		return nil, ErrNoTOTPSecret
	}

	// Verify the code
	valid := totp.Validate(code, user.TOTPSecret)
	if !valid {
		// Log failed attempt
		s.totpRepo.LogVerificationAttempt(ctx, userID, ipAddress, false)
		return nil, ErrInvalidTOTPCode
	}

	// Log successful attempt
	s.totpRepo.LogVerificationAttempt(ctx, userID, ipAddress, true)

	// Enable TOTP
	err = s.userRepo.EnableTOTP(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate backup codes
	codes, err := s.generateBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.BackupCodesResponse{Codes: codes}, nil
}

// Verify validates a TOTP code or backup code during login
func (s *TOTPService) Verify(ctx context.Context, userID int, code string, ipAddress string) (bool, error) {
	// Check rate limiting
	if exceeded, err := s.isRateLimited(ctx, userID, ipAddress); err != nil {
		return false, err
	} else if exceeded {
		return false, ErrTooManyAttempts
	}

	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return false, err
	}

	if !user.TOTPEnabled || user.TOTPSecret == "" {
		return false, ErrTOTPNotEnabled
	}

	// Try TOTP code first
	if totp.Validate(code, user.TOTPSecret) {
		s.totpRepo.LogVerificationAttempt(ctx, userID, ipAddress, true)
		return true, nil
	}

	// Try backup code
	if s.verifyAndConsumeBackupCode(ctx, userID, code, user.BackupCodes) {
		s.totpRepo.LogVerificationAttempt(ctx, userID, ipAddress, true)
		return true, nil
	}

	// Log failed attempt
	s.totpRepo.LogVerificationAttempt(ctx, userID, ipAddress, false)
	return false, ErrInvalidTOTPCode
}

// Disable disables 2FA for a user after verifying password and current TOTP code
func (s *TOTPService) Disable(ctx context.Context, userID int, password, code string) error {
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidPassword
	}

	// Verify TOTP code
	if !totp.Validate(code, user.TOTPSecret) {
		return ErrInvalidTOTPCode
	}

	// Disable TOTP
	return s.userRepo.DisableTOTP(ctx, userID)
}

// RegenerateBackupCodes creates new backup codes (invalidates old ones)
func (s *TOTPService) RegenerateBackupCodes(ctx context.Context, userID int, password string) (*models.BackupCodesResponse, error) {
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	if !user.TOTPEnabled {
		return nil, ErrTOTPNotEnabled
	}

	codes, err := s.generateBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.BackupCodesResponse{Codes: codes}, nil
}

// GetStatus returns the 2FA status for a user
func (s *TOTPService) GetStatus(ctx context.Context, userID int) (*models.User2FAStatus, error) {
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.User2FAStatus{
		Enabled:        user.TOTPEnabled,
		EnabledAt:      user.TOTPVerifiedAt,
		HasBackupCodes: user.BackupCodes != "" && user.BackupCodes != "[]",
	}, nil
}

// generateBackupCodes creates 10 random backup codes
func (s *TOTPService) generateBackupCodes(ctx context.Context, userID int) ([]string, error) {
	codes := make([]string, backupCodeCount)
	hashedCodes := make([]string, backupCodeCount)

	for i := 0; i < backupCodeCount; i++ {
		code := generateRandomCode(backupCodeLength)
		codes[i] = code

		// Hash the code for storage
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashedCodes[i] = string(hash)
	}

	// Store hashed codes as JSON
	hashedJSON, err := json.Marshal(hashedCodes)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.SetBackupCodes(ctx, userID, string(hashedJSON))
	if err != nil {
		return nil, err
	}

	return codes, nil
}

// verifyAndConsumeBackupCode checks if code matches any backup code and removes it
func (s *TOTPService) verifyAndConsumeBackupCode(ctx context.Context, userID int, code, storedCodes string) bool {
	if storedCodes == "" {
		return false
	}

	var hashedCodes []string
	if err := json.Unmarshal([]byte(storedCodes), &hashedCodes); err != nil {
		return false
	}

	for i, hash := range hashedCodes {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			// Remove the used code
			hashedCodes = append(hashedCodes[:i], hashedCodes[i+1:]...)
			updatedJSON, _ := json.Marshal(hashedCodes)
			s.userRepo.SetBackupCodes(ctx, userID, string(updatedJSON))
			return true
		}
	}

	return false
}

// isRateLimited checks if user/IP has exceeded failed attempt limit
func (s *TOTPService) isRateLimited(ctx context.Context, userID int, ipAddress string) (bool, error) {
	// Check user-based rate limit
	userAttempts, err := s.totpRepo.GetRecentFailedAttempts(ctx, userID, rateLimitWindow)
	if err != nil {
		return false, err
	}
	if userAttempts >= maxFailedAttempts {
		return true, nil
	}

	// Check IP-based rate limit
	ipAttempts, err := s.totpRepo.GetRecentFailedAttemptsByIP(ctx, ipAddress, rateLimitWindow)
	if err != nil {
		return false, err
	}
	if ipAttempts >= maxFailedAttempts*2 { // Allow more for shared IPs
		return true, nil
	}

	return false, nil
}

// generateRandomCode creates a random alphanumeric code
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes similar chars: I, O, 0, 1
	code := make([]byte, length)
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	for i := range code {
		code[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(code)
}

// Custom errors
var (
	ErrTooManyAttempts = &TOTPError{Message: "too many failed attempts, please try again later"}
	ErrNoTOTPSecret    = &TOTPError{Message: "2FA setup not initiated"}
	ErrInvalidTOTPCode = &TOTPError{Message: "invalid verification code"}
	ErrTOTPNotEnabled  = &TOTPError{Message: "2FA is not enabled"}
	ErrInvalidPassword = &TOTPError{Message: "invalid password"}
)

type TOTPError struct {
	Message string
}

func (e *TOTPError) Error() string {
	return e.Message
}
