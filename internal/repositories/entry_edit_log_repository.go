package repositories

import (
	"context"
	"time"

	"cold-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryEditLogRepository struct {
	DB *pgxpool.Pool
}

func NewEntryEditLogRepository(db *pgxpool.Pool) *EntryEditLogRepository {
	return &EntryEditLogRepository{DB: db}
}

// CreateEditLog records an entry edit
func (r *EntryEditLogRepository) CreateEditLog(ctx context.Context, log *models.EntryEditLog) error {
	query := `
		INSERT INTO entry_edit_logs (
			entry_id, edited_by_user_id,
			old_name, new_name,
			old_phone, new_phone,
			old_village, new_village,
			old_so, new_so,
			old_expected_quantity, new_expected_quantity,
			old_thock_category, new_thock_category,
			old_remark, new_remark,
			edited_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())
	`

	_, err := r.DB.Exec(ctx, query,
		log.EntryID, log.EditedByUserID,
		log.OldName, log.NewName,
		log.OldPhone, log.NewPhone,
		log.OldVillage, log.NewVillage,
		log.OldSO, log.NewSO,
		log.OldExpectedQuantity, log.NewExpectedQuantity,
		log.OldThockCategory, log.NewThockCategory,
		log.OldRemark, log.NewRemark,
	)

	return err
}

// ListAllEditLogs retrieves all entry edit logs with user and entry details
func (r *EntryEditLogRepository) ListAllEditLogs(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			el.id,
			el.entry_id,
			e.thock_number,
			e.name as customer_name,
			el.edited_by_user_id,
			u.name as editor_name,
			u.email as editor_email,
			el.old_name,
			el.new_name,
			el.old_phone,
			el.new_phone,
			el.old_village,
			el.new_village,
			el.old_so,
			el.new_so,
			el.old_expected_quantity,
			el.new_expected_quantity,
			el.old_thock_category,
			el.new_thock_category,
			el.old_remark,
			el.new_remark,
			el.edited_at
		FROM entry_edit_logs el
		JOIN entries e ON el.entry_id = e.id
		JOIN users u ON el.edited_by_user_id = u.id
		ORDER BY el.edited_at DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var (
			id                  int
			entryID             int
			thockNumber         string
			customerName        string
			editedByUserID      int
			editorName          string
			editorEmail         string
			oldName             *string
			newName             *string
			oldPhone            *string
			newPhone            *string
			oldVillage          *string
			newVillage          *string
			oldSO               *string
			newSO               *string
			oldExpectedQuantity *int
			newExpectedQuantity *int
			oldThockCategory    *string
			newThockCategory    *string
			oldRemark           *string
			newRemark           *string
			editedAt            time.Time
		)

		if err := rows.Scan(
			&id, &entryID, &thockNumber, &customerName,
			&editedByUserID, &editorName, &editorEmail,
			&oldName, &newName,
			&oldPhone, &newPhone,
			&oldVillage, &newVillage,
			&oldSO, &newSO,
			&oldExpectedQuantity, &newExpectedQuantity,
			&oldThockCategory, &newThockCategory,
			&oldRemark, &newRemark,
			&editedAt,
		); err != nil {
			return nil, err
		}

		log := map[string]interface{}{
			"id":                id,
			"entry_id":          entryID,
			"thock_number":      thockNumber,
			"customer_name":     customerName,
			"edited_by_user_id": editedByUserID,
			"editor_name":       editorName,
			"editor_email":      editorEmail,
			"edited_at":         editedAt,
			"type":              "entry", // To distinguish from room entry edits
		}

		// Only include changed fields
		if oldName != nil {
			log["old_name"] = *oldName
		}
		if newName != nil {
			log["new_name"] = *newName
		}
		if oldPhone != nil {
			log["old_phone"] = *oldPhone
		}
		if newPhone != nil {
			log["new_phone"] = *newPhone
		}
		if oldVillage != nil {
			log["old_village"] = *oldVillage
		}
		if newVillage != nil {
			log["new_village"] = *newVillage
		}
		if oldSO != nil {
			log["old_so"] = *oldSO
		}
		if newSO != nil {
			log["new_so"] = *newSO
		}
		if oldExpectedQuantity != nil {
			log["old_expected_quantity"] = *oldExpectedQuantity
		}
		if newExpectedQuantity != nil {
			log["new_expected_quantity"] = *newExpectedQuantity
		}
		if oldThockCategory != nil {
			log["old_thock_category"] = *oldThockCategory
		}
		if newThockCategory != nil {
			log["new_thock_category"] = *newThockCategory
		}
		if oldRemark != nil {
			log["old_remark"] = *oldRemark
		}
		if newRemark != nil {
			log["new_remark"] = *newRemark
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// ListByEntryID retrieves edit logs for a specific entry
func (r *EntryEditLogRepository) ListByEntryID(ctx context.Context, entryID int) ([]map[string]interface{}, error) {
	query := `
		SELECT
			el.id,
			el.entry_id,
			e.thock_number,
			el.edited_by_user_id,
			u.name as editor_name,
			u.email as editor_email,
			el.old_name,
			el.new_name,
			el.old_phone,
			el.new_phone,
			el.old_village,
			el.new_village,
			el.old_so,
			el.new_so,
			el.old_expected_quantity,
			el.new_expected_quantity,
			el.old_thock_category,
			el.new_thock_category,
			el.old_remark,
			el.new_remark,
			el.edited_at
		FROM entry_edit_logs el
		JOIN entries e ON el.entry_id = e.id
		JOIN users u ON el.edited_by_user_id = u.id
		WHERE el.entry_id = $1
		ORDER BY el.edited_at DESC
	`

	rows, err := r.DB.Query(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var (
			id                  int
			entryIDResult       int
			thockNumber         string
			editedByUserID      int
			editorName          string
			editorEmail         string
			oldName             *string
			newName             *string
			oldPhone            *string
			newPhone            *string
			oldVillage          *string
			newVillage          *string
			oldSO               *string
			newSO               *string
			oldExpectedQuantity *int
			newExpectedQuantity *int
			oldThockCategory    *string
			newThockCategory    *string
			oldRemark           *string
			newRemark           *string
			editedAt            time.Time
		)

		if err := rows.Scan(
			&id, &entryIDResult, &thockNumber,
			&editedByUserID, &editorName, &editorEmail,
			&oldName, &newName,
			&oldPhone, &newPhone,
			&oldVillage, &newVillage,
			&oldSO, &newSO,
			&oldExpectedQuantity, &newExpectedQuantity,
			&oldThockCategory, &newThockCategory,
			&oldRemark, &newRemark,
			&editedAt,
		); err != nil {
			return nil, err
		}

		log := map[string]interface{}{
			"id":                id,
			"entry_id":          entryIDResult,
			"thock_number":      thockNumber,
			"edited_by_user_id": editedByUserID,
			"editor_name":       editorName,
			"editor_email":      editorEmail,
			"edited_at":         editedAt,
		}

		if oldName != nil {
			log["old_name"] = *oldName
		}
		if newName != nil {
			log["new_name"] = *newName
		}
		if oldPhone != nil {
			log["old_phone"] = *oldPhone
		}
		if newPhone != nil {
			log["new_phone"] = *newPhone
		}
		if oldVillage != nil {
			log["old_village"] = *oldVillage
		}
		if newVillage != nil {
			log["new_village"] = *newVillage
		}
		if oldSO != nil {
			log["old_so"] = *oldSO
		}
		if newSO != nil {
			log["new_so"] = *newSO
		}
		if oldExpectedQuantity != nil {
			log["old_expected_quantity"] = *oldExpectedQuantity
		}
		if newExpectedQuantity != nil {
			log["new_expected_quantity"] = *newExpectedQuantity
		}
		if oldThockCategory != nil {
			log["old_thock_category"] = *oldThockCategory
		}
		if newThockCategory != nil {
			log["new_thock_category"] = *newThockCategory
		}
		if oldRemark != nil {
			log["old_remark"] = *oldRemark
		}
		if newRemark != nil {
			log["new_remark"] = *newRemark
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}
