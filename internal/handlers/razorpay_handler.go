package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"cold-backend/internal/middleware"
	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
	"cold-backend/internal/services"
)

type RazorpayHandler struct {
	Service      *services.RazorpayService
	CustomerRepo *repositories.CustomerRepository
}

func NewRazorpayHandler(service *services.RazorpayService, customerRepo *repositories.CustomerRepository) *RazorpayHandler {
	return &RazorpayHandler{
		Service:      service,
		CustomerRepo: customerRepo,
	}
}

// CheckPaymentStatus returns whether online payments are enabled and fee info
// GET /api/payment/status
func (h *RazorpayHandler) CheckPaymentStatus(w http.ResponseWriter, r *http.Request) {
	status := h.Service.GetPaymentStatus(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// CreateOrder creates a Razorpay order for payment
// POST /api/payment/create-order
func (h *RazorpayHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Get customer ID from context (set by auth middleware)
	customerID, ok := middleware.GetCustomerIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req models.CreateOnlinePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.PaymentScope == "" {
		req.PaymentScope = "account"
	}

	// Get customer
	customer, err := h.CustomerRepo.Get(r.Context(), customerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	// Create order
	response, err := h.Service.CreateOrder(r.Context(), customer, &req)
	if err != nil {
		log.Printf("[Razorpay] CreateOrder error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyPayment verifies the payment after Razorpay callback
// POST /api/payment/verify
func (h *RazorpayHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	// Get customer ID from context
	customerID, ok := middleware.GetCustomerIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req models.VerifyPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.RazorpayOrderID == "" || req.RazorpayPaymentID == "" || req.RazorpaySignature == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Verify payment
	tx, err := h.Service.VerifyPayment(r.Context(), &req)
	if err != nil {
		log.Printf("[Razorpay] VerifyPayment error for customer %d: %v", customerID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "Payment verified successfully",
		"transaction": tx,
	})
}

// GetMyTransactions returns customer's online payment history
// GET /api/payment/transactions
func (h *RazorpayHandler) GetMyTransactions(w http.ResponseWriter, r *http.Request) {
	customerID, ok := middleware.GetCustomerIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	transactions, err := h.Service.GetTransactionHistory(r.Context(), customerID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// HandleWebhook processes Razorpay webhook events
// POST /api/payment/webhook
func (h *RazorpayHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Razorpay] Failed to read webhook body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Verify signature
	signature := r.Header.Get("X-Razorpay-Signature")
	if !h.Service.VerifyWebhookSignature(r.Context(), body, signature) {
		log.Printf("[Razorpay] Invalid webhook signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[Razorpay] Failed to parse webhook: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	event, _ := payload["event"].(string)
	payloadData, _ := payload["payload"].(map[string]interface{})

	log.Printf("[Razorpay] Received webhook: %s", event)

	// Process webhook
	if err := h.Service.ProcessWebhook(r.Context(), event, payloadData); err != nil {
		log.Printf("[Razorpay] Webhook processing error: %v", err)
		// Return 200 anyway to prevent retries for known errors
	}

	// Always return 200 to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// GetAllTransactions returns all online transactions (admin)
// GET /api/admin/online-transactions
func (h *RazorpayHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	filter := &models.OnlineTransactionFilter{
		Limit:  50,
		Offset: 0,
	}

	// Parse query params
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			filter.Limit = n
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			filter.Offset = n
		}
	}

	if phone := r.URL.Query().Get("phone"); phone != "" {
		filter.CustomerPhone = phone
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}

	if scope := r.URL.Query().Get("scope"); scope != "" {
		filter.PaymentScope = scope
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = &t
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			filter.EndDate = &endOfDay
		}
	}

	transactions, total, err := h.Service.GetAllTransactions(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": transactions,
		"total":        total,
		"limit":        filter.Limit,
		"offset":       filter.Offset,
	})
}

// GetTransactionSummary returns summary for reports (admin)
// GET /api/admin/online-transactions/summary
func (h *RazorpayHandler) GetTransactionSummary(w http.ResponseWriter, r *http.Request) {
	var startDate, endDate *time.Time

	if s := r.URL.Query().Get("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			startDate = &t
		}
	}

	if e := r.URL.Query().Get("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			endDate = &endOfDay
		}
	}

	summary, err := h.Service.GetSummary(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
