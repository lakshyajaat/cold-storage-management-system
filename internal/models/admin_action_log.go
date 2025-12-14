package models

import "time"

type AdminActionLog struct {
	ID            int       `json:"id" db:"id"`
	AdminUserID   int       `json:"admin_user_id" db:"admin_user_id"`
	ActionType    string    `json:"action_type" db:"action_type"`
	TargetType    string    `json:"target_type" db:"target_type"`
	TargetID      *int      `json:"target_id,omitempty" db:"target_id"`
	Description   string    `json:"description" db:"description"`
	OldValue      *string   `json:"old_value,omitempty" db:"old_value"`
	NewValue      *string   `json:"new_value,omitempty" db:"new_value"`
	IPAddress     *string   `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
