package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LoginLogRepository struct {
	DB *pgxpool.Pool
}

func NewLoginLogRepository(db *pgxpool.Pool) *LoginLogRepository {
	return &LoginLogRepository{DB: db}
}

// CreateLoginLog records a new login event
func (r *LoginLogRepository) CreateLoginLog(ctx context.Context, userID int, ipAddress, userAgent string) (int, error) {
	query := `
		INSERT INTO login_logs (user_id, login_time, ip_address, user_agent)
		VALUES ($1, NOW(), $2, $3)
		RETURNING id
	`

	var logID int
	err := r.DB.QueryRow(ctx, query, userID, ipAddress, userAgent).Scan(&logID)
	if err != nil {
		return 0, err
	}

	return logID, nil
}

// UpdateLogoutTime records when a user logs out
func (r *LoginLogRepository) UpdateLogoutTime(ctx context.Context, logID int) error {
	query := `
		UPDATE login_logs
		SET logout_time = NOW()
		WHERE id = $1
	`

	_, err := r.DB.Exec(ctx, query, logID)
	return err
}

// UpdateLogoutTimeByUser records logout for the most recent login of a user
func (r *LoginLogRepository) UpdateLogoutTimeByUser(ctx context.Context, userID int) error {
	query := `
		UPDATE login_logs
		SET logout_time = NOW()
		WHERE id = (
			SELECT id FROM login_logs
			WHERE user_id = $1 AND logout_time IS NULL
			ORDER BY login_time DESC
			LIMIT 1
		)
	`

	_, err := r.DB.Exec(ctx, query, userID)
	return err
}

// ListAllLoginLogs retrieves all login/logout logs with user details
func (r *LoginLogRepository) ListAllLoginLogs(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			ll.id,
			ll.user_id,
			u.name as user_name,
			u.email,
			u.role,
			ll.login_time,
			ll.logout_time,
			ll.ip_address,
			ll.user_agent
		FROM login_logs ll
		JOIN users u ON ll.user_id = u.id
		ORDER BY ll.login_time DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var (
			id         int
			userID     int
			userName   string
			email      string
			role       string
			loginTime  time.Time
			logoutTime *time.Time
			ipAddress  *string
			userAgent  *string
		)

		if err := rows.Scan(&id, &userID, &userName, &email, &role, &loginTime, &logoutTime, &ipAddress, &userAgent); err != nil {
			return nil, err
		}

		log := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"user_name":  userName,
			"email":      email,
			"role":       role,
			"login_time": loginTime,
		}

		if logoutTime != nil {
			log["logout_time"] = *logoutTime
		}
		if ipAddress != nil {
			log["ip_address"] = *ipAddress
		}
		if userAgent != nil {
			log["user_agent"] = *userAgent
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}
