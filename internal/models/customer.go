package models

import "time"

type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Village   string    `json:"village"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCustomerRequest represents the request body for creating a customer
type CreateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Village string `json:"village"`
	Address string `json:"address"`
}

// UpdateCustomerRequest represents the request body for updating a customer
type UpdateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Village string `json:"village"`
	Address string `json:"address"`
}
