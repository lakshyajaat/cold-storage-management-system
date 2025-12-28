package models

import "time"

type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	SO        string    `json:"so"`
	Village   string    `json:"village"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCustomerRequest represents the request body for creating a customer
type CreateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	SO      string `json:"so"`
	Village string `json:"village"`
	Address string `json:"address"`
}

// UpdateCustomerRequest represents the request body for updating a customer
type UpdateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	SO      string `json:"so"`
	Village string `json:"village"`
	Address string `json:"address"`
}

// MergeCustomersRequest represents the request body for merging two customers
type MergeCustomersRequest struct {
	SourceCustomerID int `json:"source_customer_id"` // Customer to merge FROM (will be deleted)
	TargetCustomerID int `json:"target_customer_id"` // Customer to merge INTO (will keep)
}

// MergeCustomersResponse represents the response after merging customers
type MergeCustomersResponse struct {
	TargetCustomer *Customer `json:"target_customer"`
	EntriesMoved   int       `json:"entries_moved"`
	Message        string    `json:"message"`
}
