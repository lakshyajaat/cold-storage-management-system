package repositories

import (
	"context"
	"time"

	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminActionLogRepository struct {
	DB *pgxpool.Pool
}

func NewAdminActionLogRepository(db *pgxpool.Pool) *AdminActionLogRepository {
	return &AdminActionLogRepository{DB: db}
}

// CreateActionLog records an admin action
func (r *AdminActionLogRepository) CreateActionLog(ctx context.Context, log *models.AdminActionLog) error {
	query := `
		INSERT INTO admin_action_logs (
			admin_user_id, action_type, target_type, target_id,
			description, old_value, new_value, ip_address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`

	_, err := r.DB.Exec(ctx, query,
		log.AdminUserID, log.ActionType, log.TargetType, log.TargetID,
		log.Description, log.OldValue, log.NewValue, log.IPAddress,
	)

	return err
}

// ListAllActionLogs retrieves all admin action logs with admin details
func (r *AdminActionLogRepository) ListAllActionLogs(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			al.id,
			al.admin_user_id,
			u.name as admin_name,
			u.email as admin_email,
			u.role as admin_role,
			al.action_type,
			al.target_type,
			al.target_id,
			al.description,
			al.old_value,
			al.new_value,
			al.ip_address,
			al.created_at
		FROM admin_action_logs al
		JOIN users u ON al.admin_user_id = u.id
		ORDER BY al.created_at DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var (
			id            int
			adminUserID   int
			adminName     string
			adminEmail    string
			adminRole     string
			actionType    string
			targetType    string
			targetID      *int
			description   string
			oldValue      *string
			newValue      *string
			ipAddress     *string
			createdAt     time.Time
		)

		if err := rows.Scan(
			&id, &adminUserID, &adminName, &adminEmail, &adminRole,
			&actionType, &targetType, &targetID, &description,
			&oldValue, &newValue, &ipAddress, &createdAt,
		); err != nil {
			return nil, err
		}

		log := map[string]interface{}{
			"id":             id,
			"admin_user_id":  adminUserID,
			"admin_name":     adminName,
			"admin_email":    adminEmail,
			"admin_role":     adminRole,
			"action_type":    actionType,
			"target_type":    targetType,
			"description":    description,
			"created_at":     createdAt,
		}

		if targetID != nil {
			log["target_id"] = *targetID
		}
		if oldValue != nil {
			log["old_value"] = *oldValue
		}
		if newValue != nil {
			log["new_value"] = *newValue
		}
		if ipAddress != nil {
			log["ip_address"] = *ipAddress
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}
