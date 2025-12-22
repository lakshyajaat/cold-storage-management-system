package models

import "time"

// EntryEditLog records changes made to an entry
type EntryEditLog struct {
	ID                  int       `json:"id"`
	EntryID             int       `json:"entry_id"`
	EditedByUserID      int       `json:"edited_by_user_id"`
	OldName             *string   `json:"old_name,omitempty"`
	NewName             *string   `json:"new_name,omitempty"`
	OldPhone            *string   `json:"old_phone,omitempty"`
	NewPhone            *string   `json:"new_phone,omitempty"`
	OldVillage          *string   `json:"old_village,omitempty"`
	NewVillage          *string   `json:"new_village,omitempty"`
	OldSO               *string   `json:"old_so,omitempty"`
	NewSO               *string   `json:"new_so,omitempty"`
	OldExpectedQuantity *int      `json:"old_expected_quantity,omitempty"`
	NewExpectedQuantity *int      `json:"new_expected_quantity,omitempty"`
	OldThockCategory    *string   `json:"old_thock_category,omitempty"`
	NewThockCategory    *string   `json:"new_thock_category,omitempty"`
	OldRemark           *string   `json:"old_remark,omitempty"`
	NewRemark           *string   `json:"new_remark,omitempty"`
	EditedAt            time.Time `json:"edited_at"`
}
