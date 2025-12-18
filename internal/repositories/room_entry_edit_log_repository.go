package repositories

import (
	"context"
	"time"

	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomEntryEditLogRepository struct {
	DB *pgxpool.Pool
}

func NewRoomEntryEditLogRepository(db *pgxpool.Pool) *RoomEntryEditLogRepository {
	return &RoomEntryEditLogRepository{DB: db}
}

// CreateEditLog records a room entry edit
func (r *RoomEntryEditLogRepository) CreateEditLog(ctx context.Context, log *models.RoomEntryEditLog) error {
	query := `
		INSERT INTO room_entry_edit_logs (
			room_entry_id, edited_by_user_id, old_room_no, new_room_no,
			old_floor, new_floor, old_gate_no, new_gate_no,
			old_quantity, new_quantity, old_remark, new_remark, edited_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
	`

	_, err := r.DB.Exec(ctx, query,
		log.RoomEntryID, log.EditedByUserID,
		log.OldRoomNo, log.NewRoomNo,
		log.OldFloor, log.NewFloor,
		log.OldGateNo, log.NewGateNo,
		log.OldQuantity, log.NewQuantity,
		log.OldRemark, log.NewRemark,
	)

	return err
}

// ListAllEditLogs retrieves all room entry edit logs with user details
func (r *RoomEntryEditLogRepository) ListAllEditLogs(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			el.id,
			el.room_entry_id,
			re.thock_number,
			el.edited_by_user_id,
			u.name as editor_name,
			u.email as editor_email,
			el.old_room_no,
			el.new_room_no,
			el.old_floor,
			el.new_floor,
			el.old_gate_no,
			el.new_gate_no,
			el.old_quantity,
			el.new_quantity,
			el.old_remark,
			el.new_remark,
			el.edited_at
		FROM room_entry_edit_logs el
		JOIN room_entries re ON el.room_entry_id = re.id
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
			id             int
			roomEntryID    int
			thockNumber    string
			editedByUserID int
			editorName     string
			editorEmail    string
			oldRoomNo      *string
			newRoomNo      *string
			oldFloor       *string
			newFloor       *string
			oldGateNo      *string
			newGateNo      *string
			oldQuantity    *int
			newQuantity    *int
			oldRemark      *string
			newRemark      *string
			editedAt       time.Time
		)

		if err := rows.Scan(
			&id, &roomEntryID, &thockNumber, &editedByUserID, &editorName, &editorEmail,
			&oldRoomNo, &newRoomNo, &oldFloor, &newFloor,
			&oldGateNo, &newGateNo, &oldQuantity, &newQuantity,
			&oldRemark, &newRemark, &editedAt,
		); err != nil {
			return nil, err
		}

		log := map[string]interface{}{
			"id":               id,
			"room_entry_id":    roomEntryID,
			"thock_number":     thockNumber,
			"edited_by_user_id": editedByUserID,
			"editor_name":      editorName,
			"editor_email":     editorEmail,
			"edited_at":        editedAt,
		}

		if oldRoomNo != nil {
			log["old_room_no"] = *oldRoomNo
		}
		if newRoomNo != nil {
			log["new_room_no"] = *newRoomNo
		}
		if oldFloor != nil {
			log["old_floor"] = *oldFloor
		}
		if newFloor != nil {
			log["new_floor"] = *newFloor
		}
		if oldGateNo != nil {
			log["old_gate_no"] = *oldGateNo
		}
		if newGateNo != nil {
			log["new_gate_no"] = *newGateNo
		}
		if oldQuantity != nil {
			log["old_quantity"] = *oldQuantity
		}
		if newQuantity != nil {
			log["new_quantity"] = *newQuantity
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
