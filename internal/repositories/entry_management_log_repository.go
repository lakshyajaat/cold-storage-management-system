package repositories

import (
	"context"

	"cold-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryManagementLogRepository struct {
	DB *pgxpool.Pool
}

func NewEntryManagementLogRepository(db *pgxpool.Pool) *EntryManagementLogRepository {
	return &EntryManagementLogRepository{DB: db}
}

// CreateReassignLog logs an entry reassignment
func (r *EntryManagementLogRepository) CreateReassignLog(ctx context.Context, log *models.EntryManagementLog) error {
	query := `
		INSERT INTO entry_management_logs (
			action_type, performed_by_id,
			entry_id, thock_number,
			old_customer_id, old_customer_name, old_customer_phone,
			new_customer_id, new_customer_name, new_customer_phone
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	return r.DB.QueryRow(ctx, query,
		"reassign", log.PerformedByID,
		log.EntryID, log.ThockNumber,
		log.OldCustomerID, log.OldCustomerName, log.OldCustomerPhone,
		log.NewCustomerID, log.NewCustomerName, log.NewCustomerPhone,
	).Scan(&log.ID, &log.CreatedAt)
}

// CreateMergeLog logs a customer merge
func (r *EntryManagementLogRepository) CreateMergeLog(ctx context.Context, log *models.EntryManagementLog) error {
	query := `
		INSERT INTO entry_management_logs (
			action_type, performed_by_id,
			source_customer_id, source_customer_name, source_customer_phone,
			target_customer_id, target_customer_name, target_customer_phone,
			entries_moved
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return r.DB.QueryRow(ctx, query,
		"merge", log.PerformedByID,
		log.SourceCustomerID, log.SourceCustomerName, log.SourceCustomerPhone,
		log.TargetCustomerID, log.TargetCustomerName, log.TargetCustomerPhone,
		log.EntriesMoved,
	).Scan(&log.ID, &log.CreatedAt)
}

// List returns all entry management logs
func (r *EntryManagementLogRepository) List(ctx context.Context) ([]*models.EntryManagementLog, error) {
	query := `
		SELECT
			eml.id, eml.action_type, eml.performed_by_id, u.name as performed_by_name,
			eml.entry_id, eml.thock_number,
			eml.old_customer_id, eml.old_customer_name, eml.old_customer_phone,
			eml.new_customer_id, eml.new_customer_name, eml.new_customer_phone,
			eml.source_customer_id, eml.source_customer_name, eml.source_customer_phone,
			eml.target_customer_id, eml.target_customer_name, eml.target_customer_phone,
			eml.entries_moved, eml.created_at
		FROM entry_management_logs eml
		JOIN users u ON eml.performed_by_id = u.id
		ORDER BY eml.created_at DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.EntryManagementLog
	for rows.Next() {
		var log models.EntryManagementLog
		err := rows.Scan(
			&log.ID, &log.ActionType, &log.PerformedByID, &log.PerformedByName,
			&log.EntryID, &log.ThockNumber,
			&log.OldCustomerID, &log.OldCustomerName, &log.OldCustomerPhone,
			&log.NewCustomerID, &log.NewCustomerName, &log.NewCustomerPhone,
			&log.SourceCustomerID, &log.SourceCustomerName, &log.SourceCustomerPhone,
			&log.TargetCustomerID, &log.TargetCustomerName, &log.TargetCustomerPhone,
			&log.EntriesMoved, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// ListByType returns entry management logs filtered by action type
func (r *EntryManagementLogRepository) ListByType(ctx context.Context, actionType string) ([]*models.EntryManagementLog, error) {
	query := `
		SELECT
			eml.id, eml.action_type, eml.performed_by_id, u.name as performed_by_name,
			eml.entry_id, eml.thock_number,
			eml.old_customer_id, eml.old_customer_name, eml.old_customer_phone,
			eml.new_customer_id, eml.new_customer_name, eml.new_customer_phone,
			eml.source_customer_id, eml.source_customer_name, eml.source_customer_phone,
			eml.target_customer_id, eml.target_customer_name, eml.target_customer_phone,
			eml.entries_moved, eml.created_at
		FROM entry_management_logs eml
		JOIN users u ON eml.performed_by_id = u.id
		WHERE eml.action_type = $1
		ORDER BY eml.created_at DESC
	`

	rows, err := r.DB.Query(ctx, query, actionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.EntryManagementLog
	for rows.Next() {
		var log models.EntryManagementLog
		err := rows.Scan(
			&log.ID, &log.ActionType, &log.PerformedByID, &log.PerformedByName,
			&log.EntryID, &log.ThockNumber,
			&log.OldCustomerID, &log.OldCustomerName, &log.OldCustomerPhone,
			&log.NewCustomerID, &log.NewCustomerName, &log.NewCustomerPhone,
			&log.SourceCustomerID, &log.SourceCustomerName, &log.SourceCustomerPhone,
			&log.TargetCustomerID, &log.TargetCustomerName, &log.TargetCustomerPhone,
			&log.EntriesMoved, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, nil
}
