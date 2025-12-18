package g

import (
	"time"
)

// ============================================
// Legacy Gallery Models (for backward compatibility)
// ============================================

// Item represents an inventory item (legacy)
type Item struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku,omitempty"`
	Floor      int       `json:"floor"`
	CurrentQty int       `json:"current_qty"`
	UnitCost   float64   `json:"unit_cost"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Transaction represents a stock movement (legacy)
type Txn struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	Type      string    `json:"type"` // "in" or "out"
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Total     float64   `json:"total"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ItemName  string    `json:"item_name,omitempty"` // joined field
}

// AccessLog represents an access attempt
type AccessLog struct {
	ID         int       `json:"id"`
	DeviceHash string    `json:"device_hash"`
	IPAddress  string    `json:"ip_address"`
	Success    bool      `json:"success"`
	FailReason string    `json:"fail_reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Session represents an active session
type Session struct {
	ID         int       `json:"id"`
	Token      string    `json:"token"`
	DeviceHash string    `json:"device_hash"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// Summary represents the accounting summary (legacy)
type Summary struct {
	TotalItems     int            `json:"total_items"`
	TotalQty       int            `json:"total_qty"`
	TotalInvested  float64        `json:"total_invested"`
	TotalSold      float64        `json:"total_sold"`
	ProfitLoss     float64        `json:"profit_loss"`
	CurrentValue   float64        `json:"current_value"`
	FloorBreakdown []FloorSummary `json:"floor_breakdown"`
}

// FloorSummary represents summary for a single floor (legacy)
type FloorSummary struct {
	Floor     int     `json:"floor"`
	ItemCount int     `json:"item_count"`
	TotalQty  int     `json:"total_qty"`
	Value     float64 `json:"value"`
}

// Legacy Request/Response types
type AddItemRequest struct {
	Name     string  `json:"name"`
	SKU      string  `json:"sku,omitempty"`
	Floor    int     `json:"floor"`
	Qty      int     `json:"qty"`
	UnitCost float64 `json:"unit_cost"`
}

type UpdateItemRequest struct {
	Name     string  `json:"name,omitempty"`
	SKU      string  `json:"sku,omitempty"`
	Floor    int     `json:"floor,omitempty"`
	UnitCost float64 `json:"unit_cost,omitempty"`
}

type StockInRequest struct {
	ItemID    int     `json:"item_id"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
	Reason    string  `json:"reason,omitempty"`
}

type StockOutRequest struct {
	ItemID    int     `json:"item_id"`
	Qty       int     `json:"qty"`
	SalePrice float64 `json:"sale_price,omitempty"`
	Reason    string  `json:"reason"` // sold, damaged, personal, other
}

type AuthRequest struct {
	Pin1       string `json:"p1"`
	Pin2       string `json:"p2"`
	DeviceHash string `json:"dh"`
}

type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// ============================================
// Full Main System Models (mirroring cold_db)
// ============================================

// Customer represents a customer in the system
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

type CreateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	SO      string `json:"so"`
	Village string `json:"village"`
	Address string `json:"address"`
}

type UpdateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	SO      string `json:"so"`
	Village string `json:"village"`
	Address string `json:"address"`
}

// Entry represents a thock/truck entry
type Entry struct {
	ID               int       `json:"id"`
	CustomerID       int       `json:"customer_id"`
	Phone            string    `json:"phone"`
	Name             string    `json:"name"`
	Village          string    `json:"village"`
	SO               string    `json:"so"`
	ExpectedQuantity int       `json:"expected_quantity"`
	ThockCategory    string    `json:"thock_category"` // 'seed' or 'sell'
	ThockNumber      string    `json:"thock_number"`
	CreatedByUserID  int       `json:"created_by_user_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateEntryRequest struct {
	CustomerID       int    `json:"customer_id"`
	Phone            string `json:"phone"`
	Name             string `json:"name"`
	Village          string `json:"village"`
	SO               string `json:"so"`
	ExpectedQuantity int    `json:"expected_quantity"`
	ThockCategory    string `json:"thock_category"`
}

// RoomEntry represents a location assignment
type RoomEntry struct {
	ID              int       `json:"id"`
	EntryID         int       `json:"entry_id"`
	ThockNumber     string    `json:"thock_number"`
	RoomNo          string    `json:"room_no"`
	Floor           string    `json:"floor"`
	GateNo          string    `json:"gate_no"`
	Remark          string    `json:"remark"`
	Quantity        int       `json:"quantity"`
	CreatedByUserID int       `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateRoomEntryRequest struct {
	EntryID     int    `json:"entry_id"`
	ThockNumber string `json:"thock_number"`
	RoomNo      string `json:"room_no"`
	Floor       string `json:"floor"`
	GateNo      string `json:"gate_no"`
	Remark      string `json:"remark"`
	Quantity    int    `json:"quantity"`
}

type UpdateRoomEntryRequest struct {
	RoomNo   string `json:"room_no"`
	Floor    string `json:"floor"`
	GateNo   string `json:"gate_no"`
	Remark   string `json:"remark"`
	Quantity int    `json:"quantity"`
}

// GatePass represents a gate pass for stock out
type GatePass struct {
	ID                    int        `json:"id"`
	CustomerID            int        `json:"customer_id"`
	ThockNumber           string     `json:"thock_number"`
	EntryID               *int       `json:"entry_id,omitempty"`
	RequestedQuantity     int        `json:"requested_quantity"`
	ApprovedQuantity      *string    `json:"approved_quantity,omitempty"`
	FinalApprovedQuantity *int       `json:"final_approved_quantity,omitempty"`
	GateNo                *string    `json:"gate_no,omitempty"`
	Status                string     `json:"status"`
	PaymentVerified       bool       `json:"payment_verified"`
	PaymentAmount         *float64   `json:"payment_amount,omitempty"`
	TotalPickedUp         int        `json:"total_picked_up"`
	IssuedByUserID        *int       `json:"issued_by_user_id,omitempty"`
	ApprovedByUserID      *int       `json:"approved_by_user_id,omitempty"`
	CreatedByCustomerID   *int       `json:"created_by_customer_id,omitempty"`
	RequestSource         string     `json:"request_source"`
	IssuedAt              time.Time  `json:"issued_at"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	ApprovalExpiresAt     *time.Time `json:"approval_expires_at,omitempty"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	Remarks               *string    `json:"remarks,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CreateGatePassRequest struct {
	CustomerID        int     `json:"customer_id"`
	ThockNumber       string  `json:"thock_number"`
	EntryID           *int    `json:"entry_id"`
	RequestedQuantity int     `json:"requested_quantity"`
	PaymentVerified   bool    `json:"payment_verified"`
	PaymentAmount     float64 `json:"payment_amount"`
	Remarks           string  `json:"remarks"`
}

type UpdateGatePassRequest struct {
	ApprovedQuantity int    `json:"approved_quantity"`
	GateNo           string `json:"gate_no"`
	Status           string `json:"status"`
	RequestSource    string `json:"request_source,omitempty"`
	Remarks          string `json:"remarks"`
}

type RecordPickupRequest struct {
	GatePassID     int    `json:"gate_pass_id"`
	PickupQuantity int    `json:"pickup_quantity"`
	RoomNo         string `json:"room_no"`
	Floor          string `json:"floor"`
	GatarNo        string `json:"gatar_no"`
	Remarks        string `json:"remarks"`
}

// GatePassPickup represents a pickup record
type GatePassPickup struct {
	ID               int       `json:"id"`
	GatePassID       int       `json:"gate_pass_id"`
	Quantity         int       `json:"quantity"`
	RoomNo           *string   `json:"room_no,omitempty"`
	Floor            *string   `json:"floor,omitempty"`
	GatarNo          *string   `json:"gatar_no,omitempty"`
	CreatedByUserID  int       `json:"created_by_user_id"`
	CreatedAt        time.Time `json:"created_at"`
}

// EntryEvent represents an audit event
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
	EventTypeGatePassIssued = "GATE_PASS_ISSUED"
	EventTypeItemsOut     = "ITEMS_OUT"
)

// Status constants
const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusInStorage  = "IN_STORAGE"
	StatusCompleted  = "COMPLETED"
	StatusOnHold     = "ON_HOLD"
)

// RentPayment represents a rent payment record
type RentPayment struct {
	ID                int       `json:"id"`
	ReceiptNumber     string    `json:"receipt_number"`
	EntryID           int       `json:"entry_id"`
	CustomerName      string    `json:"customer_name"`
	CustomerPhone     string    `json:"customer_phone"`
	TotalRent         float64   `json:"total_rent"`
	AmountPaid        float64   `json:"amount_paid"`
	Balance           float64   `json:"balance"`
	PaymentDate       time.Time `json:"payment_date"`
	ProcessedByUserID int       `json:"processed_by_user_id"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateRentPaymentRequest struct {
	EntryID       int     `json:"entry_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	TotalRent     float64 `json:"total_rent"`
	AmountPaid    float64 `json:"amount_paid"`
	Balance       float64 `json:"balance"`
	Notes         string  `json:"notes"`
}

// SystemSetting represents a system configuration
type SystemSetting struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateSettingRequest struct {
	Value string `json:"value"`
}

// DashboardSummary represents the dashboard data
type DashboardSummary struct {
	TotalCustomers    int     `json:"total_customers"`
	TotalEntries      int     `json:"total_entries"`
	TotalQuantity     int     `json:"total_quantity"`
	PendingGatePasses int     `json:"pending_gate_passes"`
	TodayEntries      int     `json:"today_entries"`
	TodayGatePasses   int     `json:"today_gate_passes"`
	TotalRentCollected float64 `json:"total_rent_collected"`
	RoomBreakdown     []RoomSummary `json:"room_breakdown"`
}

type RoomSummary struct {
	RoomNo    string `json:"room_no"`
	TotalQty  int    `json:"total_qty"`
	EntryCount int   `json:"entry_count"`
}
