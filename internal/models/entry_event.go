package models

import "time"

type EntryEvent struct {
	ID              int       `json:"id"`
	EntryID         int       `json:"entry_id"`
	EventType       string    `json:"event_type"`
	Status          string    `json:"status"`
	Notes           string    `json:"notes"`
	CreatedByUserID int       `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateEntryEventRequest struct {
	EntryID   int    `json:"entry_id"`
	EventType string `json:"event_type"`
	Status    string `json:"status"`
	Notes     string `json:"notes"`
}

// Event type constants
const (
	EventTypeCreated      = "CREATED"
	EventTypeInStorage    = "IN_STORAGE"
	EventTypeProcessing   = "PROCESSING"
	EventTypeQualityCheck = "QUALITY_CHECK"
	EventTypeReady        = "READY"
	EventTypeCompleted    = "COMPLETED"
)

// Status constants
const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusInStorage  = "IN_STORAGE"
	StatusCompleted  = "COMPLETED"
	StatusOnHold     = "ON_HOLD"
)
