package repositories

import (
	"context"
	"fmt"
	"time"

	"cold-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PendingSettingChangeRepository struct {
	DB *pgxpool.Pool
}

func NewPendingSettingChangeRepository(db *pgxpool.Pool) *PendingSettingChangeRepository {
	return &PendingSettingChangeRepository{DB: db}
}

// Create creates a new pending setting change request
func (r *PendingSettingChangeRepository) Create(ctx context.Context, change *models.PendingSettingChange) error {
	query := `
		INSERT INTO pending_setting_changes (
			setting_key, old_value, new_value, requested_by, reason
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, requested_at, status, expires_at
	`

	err := r.DB.QueryRow(ctx, query,
		change.SettingKey,
		change.OldValue,
		change.NewValue,
		change.RequestedBy,
		change.Reason,
	).Scan(&change.ID, &change.RequestedAt, &change.Status, &change.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to create pending setting change: %w", err)
	}

	return nil
}

// GetByID retrieves a pending change by ID
func (r *PendingSettingChangeRepository) GetByID(ctx context.Context, id int) (*models.PendingSettingChange, error) {
	query := `
		SELECT
			psc.id, psc.setting_key, COALESCE(psc.old_value, ''), psc.new_value,
			psc.requested_by, COALESCE(u1.name, u1.email) as requested_by_name,
			psc.requested_at, COALESCE(psc.reason, ''),
			psc.approved_by, COALESCE(u2.name, u2.email, '') as approved_by_name,
			psc.approved_at, psc.status, COALESCE(psc.rejection_reason, ''),
			psc.expires_at
		FROM pending_setting_changes psc
		LEFT JOIN users u1 ON psc.requested_by = u1.id
		LEFT JOIN users u2 ON psc.approved_by = u2.id
		WHERE psc.id = $1
	`

	change := &models.PendingSettingChange{}
	var approvedByName *string

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&change.ID, &change.SettingKey, &change.OldValue, &change.NewValue,
		&change.RequestedBy, &change.RequestedByName,
		&change.RequestedAt, &change.Reason,
		&change.ApprovedBy, &approvedByName,
		&change.ApprovedAt, &change.Status, &change.RejectionReason,
		&change.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	if approvedByName != nil {
		change.ApprovedByName = *approvedByName
	}

	return change, nil
}

// GetPendingBySettingKey gets pending changes for a specific setting
func (r *PendingSettingChangeRepository) GetPendingBySettingKey(ctx context.Context, settingKey string) (*models.PendingSettingChange, error) {
	query := `
		SELECT
			psc.id, psc.setting_key, COALESCE(psc.old_value, ''), psc.new_value,
			psc.requested_by, COALESCE(u1.name, u1.email) as requested_by_name,
			psc.requested_at, COALESCE(psc.reason, ''),
			psc.approved_by, COALESCE(u2.name, u2.email, '') as approved_by_name,
			psc.approved_at, psc.status, COALESCE(psc.rejection_reason, ''),
			psc.expires_at
		FROM pending_setting_changes psc
		LEFT JOIN users u1 ON psc.requested_by = u1.id
		LEFT JOIN users u2 ON psc.approved_by = u2.id
		WHERE psc.setting_key = $1 AND psc.status = 'pending' AND psc.expires_at > NOW()
		ORDER BY psc.requested_at DESC
		LIMIT 1
	`

	change := &models.PendingSettingChange{}
	var approvedByName *string

	err := r.DB.QueryRow(ctx, query, settingKey).Scan(
		&change.ID, &change.SettingKey, &change.OldValue, &change.NewValue,
		&change.RequestedBy, &change.RequestedByName,
		&change.RequestedAt, &change.Reason,
		&change.ApprovedBy, &approvedByName,
		&change.ApprovedAt, &change.Status, &change.RejectionReason,
		&change.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	if approvedByName != nil {
		change.ApprovedByName = *approvedByName
	}

	return change, nil
}

// GetAllPending gets all pending setting changes
func (r *PendingSettingChangeRepository) GetAllPending(ctx context.Context) ([]*models.PendingSettingChange, error) {
	query := `
		SELECT
			psc.id, psc.setting_key, COALESCE(psc.old_value, ''), psc.new_value,
			psc.requested_by, COALESCE(u1.name, u1.email) as requested_by_name,
			psc.requested_at, COALESCE(psc.reason, ''),
			psc.expires_at
		FROM pending_setting_changes psc
		LEFT JOIN users u1 ON psc.requested_by = u1.id
		WHERE psc.status = 'pending' AND psc.expires_at > NOW()
		ORDER BY psc.requested_at DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*models.PendingSettingChange
	for rows.Next() {
		change := &models.PendingSettingChange{Status: models.PendingSettingStatusPending}
		err := rows.Scan(
			&change.ID, &change.SettingKey, &change.OldValue, &change.NewValue,
			&change.RequestedBy, &change.RequestedByName,
			&change.RequestedAt, &change.Reason,
			&change.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// Approve marks a pending change as approved and returns it
func (r *PendingSettingChangeRepository) Approve(ctx context.Context, id int, approverID int) error {
	now := time.Now()
	query := `
		UPDATE pending_setting_changes
		SET status = 'approved', approved_by = $2, approved_at = $3
		WHERE id = $1 AND status = 'pending' AND expires_at > NOW()
	`

	result, err := r.DB.Exec(ctx, query, id, approverID, now)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending change not found or already processed")
	}

	return nil
}

// Reject marks a pending change as rejected
func (r *PendingSettingChangeRepository) Reject(ctx context.Context, id int, rejectorID int, reason string) error {
	query := `
		UPDATE pending_setting_changes
		SET status = 'rejected', approved_by = $2, approved_at = NOW(), rejection_reason = $3
		WHERE id = $1 AND status = 'pending'
	`

	result, err := r.DB.Exec(ctx, query, id, rejectorID, reason)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending change not found or already processed")
	}

	return nil
}

// ExpireOld marks old pending changes as expired
func (r *PendingSettingChangeRepository) ExpireOld(ctx context.Context) error {
	query := `
		UPDATE pending_setting_changes
		SET status = 'expired'
		WHERE status = 'pending' AND expires_at <= NOW()
	`

	_, err := r.DB.Exec(ctx, query)
	return err
}

// IsProtectedSetting checks if a setting requires dual admin approval
func (r *PendingSettingChangeRepository) IsProtectedSetting(ctx context.Context, settingKey string) (bool, error) {
	var count int
	err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM protected_settings WHERE setting_key = $1", settingKey).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetProtectedSettings returns all protected settings
func (r *PendingSettingChangeRepository) GetProtectedSettings(ctx context.Context) ([]models.ProtectedSetting, error) {
	query := `SELECT setting_key, description FROM protected_settings ORDER BY setting_key`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.ProtectedSetting
	for rows.Next() {
		var s models.ProtectedSetting
		if err := rows.Scan(&s.SettingKey, &s.Description); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}

	return settings, nil
}

// GetHistory returns recent setting change history
func (r *PendingSettingChangeRepository) GetHistory(ctx context.Context, limit int) ([]*models.PendingSettingChange, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT
			psc.id, psc.setting_key, COALESCE(psc.old_value, ''), psc.new_value,
			psc.requested_by, COALESCE(u1.name, u1.email) as requested_by_name,
			psc.requested_at, COALESCE(psc.reason, ''),
			psc.approved_by, COALESCE(u2.name, u2.email, '') as approved_by_name,
			psc.approved_at, psc.status, COALESCE(psc.rejection_reason, ''),
			psc.expires_at
		FROM pending_setting_changes psc
		LEFT JOIN users u1 ON psc.requested_by = u1.id
		LEFT JOIN users u2 ON psc.approved_by = u2.id
		ORDER BY psc.requested_at DESC
		LIMIT $1
	`

	rows, err := r.DB.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*models.PendingSettingChange
	for rows.Next() {
		change := &models.PendingSettingChange{}
		var approvedByName *string

		err := rows.Scan(
			&change.ID, &change.SettingKey, &change.OldValue, &change.NewValue,
			&change.RequestedBy, &change.RequestedByName,
			&change.RequestedAt, &change.Reason,
			&change.ApprovedBy, &approvedByName,
			&change.ApprovedAt, &change.Status, &change.RejectionReason,
			&change.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		if approvedByName != nil {
			change.ApprovedByName = *approvedByName
		}

		changes = append(changes, change)
	}

	return changes, nil
}
