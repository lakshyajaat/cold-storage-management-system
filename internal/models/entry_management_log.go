package models

import "time"

// EntryManagementLog records entry reassignments and customer merges
type EntryManagementLog struct {
	ID              int       `json:"id"`
	ActionType      string    `json:"action_type"` // "reassign" or "merge"
	PerformedByID   int       `json:"performed_by_id"`
	PerformedByName string    `json:"performed_by_name"`

	// For reassignment
	EntryID          *int    `json:"entry_id,omitempty"`
	ThockNumber      *string `json:"thock_number,omitempty"`
	OldCustomerID    *int    `json:"old_customer_id,omitempty"`
	OldCustomerName  *string `json:"old_customer_name,omitempty"`
	OldCustomerPhone *string `json:"old_customer_phone,omitempty"`
	NewCustomerID    *int    `json:"new_customer_id,omitempty"`
	NewCustomerName  *string `json:"new_customer_name,omitempty"`
	NewCustomerPhone *string `json:"new_customer_phone,omitempty"`

	// For merge
	SourceCustomerID      *int          `json:"source_customer_id,omitempty"`
	SourceCustomerName    *string       `json:"source_customer_name,omitempty"`
	SourceCustomerPhone   *string       `json:"source_customer_phone,omitempty"`
	SourceCustomerVillage string        `json:"source_customer_village,omitempty"`
	SourceCustomerSO      string        `json:"source_customer_so,omitempty"`
	TargetCustomerID      *int          `json:"target_customer_id,omitempty"`
	TargetCustomerName    *string       `json:"target_customer_name,omitempty"`
	TargetCustomerPhone   *string       `json:"target_customer_phone,omitempty"`
	TargetCustomerVillage string        `json:"target_customer_village,omitempty"`
	TargetCustomerSO      string        `json:"target_customer_so,omitempty"`
	EntriesMoved          *int          `json:"entries_moved,omitempty"`
	PaymentsMoved         int           `json:"payments_moved,omitempty"`
	MergeDetails          *MergeDetails `json:"merge_details,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}
