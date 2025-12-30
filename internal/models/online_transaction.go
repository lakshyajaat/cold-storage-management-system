package models

import "time"

// OnlineTransactionStatus represents the status of an online payment
type OnlineTransactionStatus string

const (
	OnlineTxStatusPending  OnlineTransactionStatus = "pending"
	OnlineTxStatusSuccess  OnlineTransactionStatus = "success"
	OnlineTxStatusFailed   OnlineTransactionStatus = "failed"
	OnlineTxStatusRefunded OnlineTransactionStatus = "refunded"
)

// PaymentScope represents what the payment is for
type PaymentScope string

const (
	PaymentScopeTruck        PaymentScope = "truck"
	PaymentScopeFamilyMember PaymentScope = "family_member"
	PaymentScopeAccount      PaymentScope = "account"
)

// OnlineTransaction represents a Razorpay payment transaction
type OnlineTransaction struct {
	ID                int                     `json:"id"`
	RazorpayOrderID   string                  `json:"razorpay_order_id"`
	RazorpayPaymentID string                  `json:"razorpay_payment_id,omitempty"`
	RazorpaySignature string                  `json:"-"` // Don't expose signature in JSON

	// Customer info
	CustomerID    int    `json:"customer_id"`
	CustomerPhone string `json:"customer_phone"`
	CustomerName  string `json:"customer_name"`

	// Payment scope
	EntryID          *int   `json:"entry_id,omitempty"`
	FamilyMemberID   *int   `json:"family_member_id,omitempty"`
	ThockNumber      string `json:"thock_number,omitempty"`
	FamilyMemberName string `json:"family_member_name,omitempty"`
	PaymentScope     string `json:"payment_scope"`

	// Amounts (in rupees)
	Amount      float64 `json:"amount"`       // Original payment amount
	FeeAmount   float64 `json:"fee_amount"`   // Transaction fee
	TotalAmount float64 `json:"total_amount"` // What customer pays (amount + fee)

	// Payment details from Razorpay
	UTRNumber     string `json:"utr_number,omitempty"`
	PaymentMethod string `json:"payment_method,omitempty"` // upi, card, netbanking, wallet
	Bank          string `json:"bank,omitempty"`
	VPA           string `json:"vpa,omitempty"`      // UPI ID
	CardLast4     string `json:"card_last4,omitempty"`
	CardNetwork   string `json:"card_network,omitempty"`

	// Status
	Status        OnlineTransactionStatus `json:"status"`
	FailureReason string                  `json:"failure_reason,omitempty"`

	// Linked records
	RentPaymentID *int `json:"rent_payment_id,omitempty"`
	LedgerEntryID *int `json:"ledger_entry_id,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CreateOnlinePaymentRequest is the request from customer portal to initiate payment
type CreateOnlinePaymentRequest struct {
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	EntryID          *int    `json:"entry_id,omitempty"`
	FamilyMemberID   *int    `json:"family_member_id,omitempty"`
	FamilyMemberName string  `json:"family_member_name,omitempty"`
	ThockNumber      string  `json:"thock_number,omitempty"`
	PaymentScope     string  `json:"payment_scope" validate:"required,oneof=truck family_member account"`
}

// CreateOrderResponse is returned to frontend for Razorpay checkout
type CreateOrderResponse struct {
	OrderID       string  `json:"order_id"`
	Amount        int     `json:"amount"`        // In paise
	FeeAmount     int     `json:"fee_amount"`    // In paise
	TotalAmount   int     `json:"total_amount"`  // In paise
	Currency      string  `json:"currency"`
	KeyID         string  `json:"key_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	CustomerEmail string  `json:"customer_email,omitempty"`
	FeePercent    float64 `json:"fee_percent"`
}

// VerifyPaymentRequest is sent from frontend after Razorpay callback
type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" validate:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" validate:"required"`
	RazorpaySignature string `json:"razorpay_signature" validate:"required"`
}

// PaymentStatusResponse is returned when checking if online payments are enabled
type PaymentStatusResponse struct {
	Enabled    bool    `json:"enabled"`
	FeePercent float64 `json:"fee_percent"`
	KeyID      string  `json:"key_id,omitempty"`
}

// RazorpayWebhookPayload represents the webhook payload from Razorpay
type RazorpayWebhookPayload struct {
	Event     string                 `json:"event"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt int64                  `json:"created_at"`
}

// OnlineTransactionFilter is used for listing/filtering transactions
type OnlineTransactionFilter struct {
	CustomerPhone string     `json:"customer_phone,omitempty"`
	CustomerID    int        `json:"customer_id,omitempty"`
	Status        string     `json:"status,omitempty"`
	PaymentScope  string     `json:"payment_scope,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// OnlinePaymentSummary is for admin reports
type OnlinePaymentSummary struct {
	TotalTransactions   int     `json:"total_transactions"`
	SuccessfulPayments  int     `json:"successful_payments"`
	FailedTransactions  int     `json:"failed_transactions"`
	PendingTransactions int     `json:"pending_transactions"`
	TotalAmount         float64 `json:"total_amount"`       // Sum of successful payment amounts
	TotalFees           float64 `json:"total_fees"`         // Sum of fees collected
	TotalCollected      float64 `json:"total_collected"`    // Total amount + fees
}
