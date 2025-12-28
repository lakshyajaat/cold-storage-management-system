package models

import "time"

type Customer struct {
	ID                   int        `json:"id"`
	Name                 string     `json:"name"`
	Phone                string     `json:"phone"`
	SO                   string     `json:"so"`
	Village              string     `json:"village"`
	Address              string     `json:"address"`
	Status               string     `json:"status"`                  // 'active', 'merged', 'inactive'
	MergedIntoCustomerID *int       `json:"merged_into_customer_id"` // If merged, points to target customer
	MergedAt             *time.Time `json:"merged_at"`               // When merge happened
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
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
	TargetCustomer *Customer     `json:"target_customer"`
	EntriesMoved   int           `json:"entries_moved"`
	PaymentsMoved  int           `json:"payments_moved"`
	MergeDetails   *MergeDetails `json:"merge_details,omitempty"`
	Message        string        `json:"message"`
}

// MergeDetails contains detailed information about what was transferred during a merge
type MergeDetails struct {
	Entries  []MergeEntryDetail   `json:"entries"`
	Payments []MergePaymentDetail `json:"payments"`
}

// MergeEntryDetail contains info about a single entry that was transferred
type MergeEntryDetail struct {
	ID               int    `json:"id"`
	ThockNumber      string `json:"thock_number"`
	ExpectedQuantity int    `json:"expected_quantity"`
	ThockCategory    string `json:"thock_category"`
}

// MergePaymentDetail contains info about a single payment that was transferred
type MergePaymentDetail struct {
	ID            int     `json:"id"`
	Amount        float64 `json:"amount"`
	ReceiptNumber string  `json:"receipt_number"`
	PaymentDate   string  `json:"payment_date"`
}
