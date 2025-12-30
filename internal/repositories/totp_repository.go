package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TOTPRepository struct {
	DB *pgxpool.Pool
}

func NewTOTPRepository(db *pgxpool.Pool) *TOTPRepository {
	return &TOTPRepository{DB: db}
}

// LogVerificationAttempt records a 2FA verification attempt for rate limiting
func (r *TOTPRepository) LogVerificationAttempt(ctx context.Context, userID int, ipAddress string, success bool) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO totp_verification_attempts (user_id, ip_address, success) VALUES ($1, $2, $3)`,
		userID, ipAddress, success)
	return err
}

// GetRecentFailedAttempts returns failed attempt count for a user in time window
func (r *TOTPRepository) GetRecentFailedAttempts(ctx context.Context, userID int, window time.Duration) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM totp_verification_attempts
		 WHERE user_id = $1 AND success = false AND created_at > $2`,
		userID, time.Now().Add(-window)).Scan(&count)
	return count, err
}

// GetRecentFailedAttemptsByIP returns failed attempts from an IP address in time window
func (r *TOTPRepository) GetRecentFailedAttemptsByIP(ctx context.Context, ip string, window time.Duration) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM totp_verification_attempts
		 WHERE ip_address = $1 AND success = false AND created_at > $2`,
		ip, time.Now().Add(-window)).Scan(&count)
	return count, err
}

// CleanupOldAttempts removes attempts older than 24 hours
func (r *TOTPRepository) CleanupOldAttempts(ctx context.Context) error {
	_, err := r.DB.Exec(ctx,
		`DELETE FROM totp_verification_attempts WHERE created_at < NOW() - INTERVAL '24 hours'`)
	return err
}
