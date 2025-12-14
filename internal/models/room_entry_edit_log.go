package models

import "time"

type RoomEntryEditLog struct {
	ID             int       `json:"id" db:"id"`
	RoomEntryID    int       `json:"room_entry_id" db:"room_entry_id"`
	EditedByUserID int       `json:"edited_by_user_id" db:"edited_by_user_id"`
	OldRoomNo      *string   `json:"old_room_no,omitempty" db:"old_room_no"`
	NewRoomNo      *string   `json:"new_room_no,omitempty" db:"new_room_no"`
	OldFloor       *string   `json:"old_floor,omitempty" db:"old_floor"`
	NewFloor       *string   `json:"new_floor,omitempty" db:"new_floor"`
	OldGateNo      *string   `json:"old_gate_no,omitempty" db:"old_gate_no"`
	NewGateNo      *string   `json:"new_gate_no,omitempty" db:"new_gate_no"`
	OldQuantity    *int      `json:"old_quantity,omitempty" db:"old_quantity"`
	NewQuantity    *int      `json:"new_quantity,omitempty" db:"new_quantity"`
	OldRemark      *string   `json:"old_remark,omitempty" db:"old_remark"`
	NewRemark      *string   `json:"new_remark,omitempty" db:"new_remark"`
	EditedAt       time.Time `json:"edited_at" db:"edited_at"`
}
