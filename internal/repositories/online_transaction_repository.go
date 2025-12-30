package repositories

import (
	"context"
	"fmt"
	"time"

	"cold-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OnlineTransactionRepository struct {
	DB *pgxpool.Pool
}

func NewOnlineTransactionRepository(db *pgxpool.Pool) *OnlineTransactionRepository {
	return &OnlineTransactionRepository{DB: db}
}

// Create creates a new online transaction record
func (r *OnlineTransactionRepository) Create(ctx context.Context, tx *models.OnlineTransaction) error {
	query := `
		INSERT INTO online_transactions (
			razorpay_order_id, customer_id, customer_phone, customer_name,
			entry_id, family_member_id, thock_number, family_member_name, payment_scope,
			amount, fee_amount, total_amount, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err := r.DB.QueryRow(ctx, query,
		tx.RazorpayOrderID,
		tx.CustomerID,
		tx.CustomerPhone,
		tx.CustomerName,
		tx.EntryID,
		tx.FamilyMemberID,
		tx.ThockNumber,
		tx.FamilyMemberName,
		tx.PaymentScope,
		tx.Amount,
		tx.FeeAmount,
		tx.TotalAmount,
		models.OnlineTxStatusPending,
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create online transaction: %w", err)
	}

	tx.Status = models.OnlineTxStatusPending
	return nil
}

// GetByOrderID retrieves a transaction by Razorpay order ID
func (r *OnlineTransactionRepository) GetByOrderID(ctx context.Context, orderID string) (*models.OnlineTransaction, error) {
	query := `
		SELECT id, razorpay_order_id, COALESCE(razorpay_payment_id, ''), COALESCE(razorpay_signature, ''),
		       customer_id, customer_phone, customer_name,
		       entry_id, family_member_id, COALESCE(thock_number, ''), COALESCE(family_member_name, ''), payment_scope,
		       amount, fee_amount, total_amount,
		       COALESCE(utr_number, ''), COALESCE(payment_method, ''), COALESCE(bank, ''), COALESCE(vpa, ''),
		       COALESCE(card_last4, ''), COALESCE(card_network, ''),
		       status, COALESCE(failure_reason, ''),
		       rent_payment_id, ledger_entry_id,
		       created_at, completed_at
		FROM online_transactions
		WHERE razorpay_order_id = $1
	`

	tx := &models.OnlineTransaction{}
	err := r.DB.QueryRow(ctx, query, orderID).Scan(
		&tx.ID, &tx.RazorpayOrderID, &tx.RazorpayPaymentID, &tx.RazorpaySignature,
		&tx.CustomerID, &tx.CustomerPhone, &tx.CustomerName,
		&tx.EntryID, &tx.FamilyMemberID, &tx.ThockNumber, &tx.FamilyMemberName, &tx.PaymentScope,
		&tx.Amount, &tx.FeeAmount, &tx.TotalAmount,
		&tx.UTRNumber, &tx.PaymentMethod, &tx.Bank, &tx.VPA,
		&tx.CardLast4, &tx.CardNetwork,
		&tx.Status, &tx.FailureReason,
		&tx.RentPaymentID, &tx.LedgerEntryID,
		&tx.CreatedAt, &tx.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// GetByPaymentID retrieves a transaction by Razorpay payment ID
func (r *OnlineTransactionRepository) GetByPaymentID(ctx context.Context, paymentID string) (*models.OnlineTransaction, error) {
	query := `
		SELECT id, razorpay_order_id, COALESCE(razorpay_payment_id, ''), COALESCE(razorpay_signature, ''),
		       customer_id, customer_phone, customer_name,
		       entry_id, family_member_id, COALESCE(thock_number, ''), COALESCE(family_member_name, ''), payment_scope,
		       amount, fee_amount, total_amount,
		       COALESCE(utr_number, ''), COALESCE(payment_method, ''), COALESCE(bank, ''), COALESCE(vpa, ''),
		       COALESCE(card_last4, ''), COALESCE(card_network, ''),
		       status, COALESCE(failure_reason, ''),
		       rent_payment_id, ledger_entry_id,
		       created_at, completed_at
		FROM online_transactions
		WHERE razorpay_payment_id = $1
	`

	tx := &models.OnlineTransaction{}
	err := r.DB.QueryRow(ctx, query, paymentID).Scan(
		&tx.ID, &tx.RazorpayOrderID, &tx.RazorpayPaymentID, &tx.RazorpaySignature,
		&tx.CustomerID, &tx.CustomerPhone, &tx.CustomerName,
		&tx.EntryID, &tx.FamilyMemberID, &tx.ThockNumber, &tx.FamilyMemberName, &tx.PaymentScope,
		&tx.Amount, &tx.FeeAmount, &tx.TotalAmount,
		&tx.UTRNumber, &tx.PaymentMethod, &tx.Bank, &tx.VPA,
		&tx.CardLast4, &tx.CardNetwork,
		&tx.Status, &tx.FailureReason,
		&tx.RentPaymentID, &tx.LedgerEntryID,
		&tx.CreatedAt, &tx.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// UpdatePaymentSuccess updates the transaction with successful payment details
func (r *OnlineTransactionRepository) UpdatePaymentSuccess(ctx context.Context, orderID, paymentID, signature, utr, method, bank, vpa, cardLast4, cardNetwork string) error {
	now := time.Now()
	query := `
		UPDATE online_transactions
		SET razorpay_payment_id = $2,
		    razorpay_signature = $3,
		    utr_number = $4,
		    payment_method = $5,
		    bank = $6,
		    vpa = $7,
		    card_last4 = $8,
		    card_network = $9,
		    status = $10,
		    completed_at = $11
		WHERE razorpay_order_id = $1
	`

	_, err := r.DB.Exec(ctx, query,
		orderID, paymentID, signature, utr, method, bank, vpa, cardLast4, cardNetwork,
		models.OnlineTxStatusSuccess, now,
	)

	return err
}

// UpdatePaymentFailed marks the transaction as failed
func (r *OnlineTransactionRepository) UpdatePaymentFailed(ctx context.Context, orderID, reason string) error {
	now := time.Now()
	query := `
		UPDATE online_transactions
		SET status = $2, failure_reason = $3, completed_at = $4
		WHERE razorpay_order_id = $1
	`

	_, err := r.DB.Exec(ctx, query, orderID, models.OnlineTxStatusFailed, reason, now)
	return err
}

// LinkToRentPayment links the transaction to created rent payment and ledger entry
func (r *OnlineTransactionRepository) LinkToRentPayment(ctx context.Context, orderID string, rentPaymentID, ledgerEntryID int) error {
	query := `
		UPDATE online_transactions
		SET rent_payment_id = $2, ledger_entry_id = $3
		WHERE razorpay_order_id = $1
	`

	_, err := r.DB.Exec(ctx, query, orderID, rentPaymentID, ledgerEntryID)
	return err
}

// GetByCustomer returns online transactions for a customer
func (r *OnlineTransactionRepository) GetByCustomer(ctx context.Context, customerID int, limit, offset int) ([]*models.OnlineTransaction, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, razorpay_order_id, COALESCE(razorpay_payment_id, ''),
		       customer_id, customer_phone, customer_name,
		       entry_id, family_member_id, COALESCE(thock_number, ''), COALESCE(family_member_name, ''), payment_scope,
		       amount, fee_amount, total_amount,
		       COALESCE(utr_number, ''), COALESCE(payment_method, ''),
		       status, COALESCE(failure_reason, ''),
		       created_at, completed_at
		FROM online_transactions
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.DB.Query(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.OnlineTransaction
	for rows.Next() {
		tx := &models.OnlineTransaction{}
		err := rows.Scan(
			&tx.ID, &tx.RazorpayOrderID, &tx.RazorpayPaymentID,
			&tx.CustomerID, &tx.CustomerPhone, &tx.CustomerName,
			&tx.EntryID, &tx.FamilyMemberID, &tx.ThockNumber, &tx.FamilyMemberName, &tx.PaymentScope,
			&tx.Amount, &tx.FeeAmount, &tx.TotalAmount,
			&tx.UTRNumber, &tx.PaymentMethod,
			&tx.Status, &tx.FailureReason,
			&tx.CreatedAt, &tx.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetAll returns all online transactions with optional filters
func (r *OnlineTransactionRepository) GetAll(ctx context.Context, filter *models.OnlineTransactionFilter) ([]*models.OnlineTransaction, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if filter.CustomerPhone != "" {
		whereClause += fmt.Sprintf(" AND customer_phone = $%d", argNum)
		args = append(args, filter.CustomerPhone)
		argNum++
	}

	if filter.CustomerID > 0 {
		whereClause += fmt.Sprintf(" AND customer_id = $%d", argNum)
		args = append(args, filter.CustomerID)
		argNum++
	}

	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}

	if filter.PaymentScope != "" {
		whereClause += fmt.Sprintf(" AND payment_scope = $%d", argNum)
		args = append(args, filter.PaymentScope)
		argNum++
	}

	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argNum)
		args = append(args, *filter.StartDate)
		argNum++
	}

	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argNum)
		args = append(args, *filter.EndDate)
		argNum++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM online_transactions %s", whereClause)
	var total int
	err := r.DB.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get data
	query := fmt.Sprintf(`
		SELECT id, razorpay_order_id, COALESCE(razorpay_payment_id, ''),
		       customer_id, customer_phone, customer_name,
		       entry_id, family_member_id, COALESCE(thock_number, ''), COALESCE(family_member_name, ''), payment_scope,
		       amount, fee_amount, total_amount,
		       COALESCE(utr_number, ''), COALESCE(payment_method, ''), COALESCE(bank, ''), COALESCE(vpa, ''),
		       status, COALESCE(failure_reason, ''),
		       created_at, completed_at
		FROM online_transactions
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*models.OnlineTransaction
	for rows.Next() {
		tx := &models.OnlineTransaction{}
		err := rows.Scan(
			&tx.ID, &tx.RazorpayOrderID, &tx.RazorpayPaymentID,
			&tx.CustomerID, &tx.CustomerPhone, &tx.CustomerName,
			&tx.EntryID, &tx.FamilyMemberID, &tx.ThockNumber, &tx.FamilyMemberName, &tx.PaymentScope,
			&tx.Amount, &tx.FeeAmount, &tx.TotalAmount,
			&tx.UTRNumber, &tx.PaymentMethod, &tx.Bank, &tx.VPA,
			&tx.Status, &tx.FailureReason,
			&tx.CreatedAt, &tx.CompletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// GetSummary returns summary statistics for online payments
func (r *OnlineTransactionRepository) GetSummary(ctx context.Context, startDate, endDate *time.Time) (*models.OnlinePaymentSummary, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argNum)
		args = append(args, *startDate)
		argNum++
	}

	if endDate != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argNum)
		args = append(args, *endDate)
		argNum++
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_transactions,
			COUNT(*) FILTER (WHERE status = 'success') as successful_payments,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_transactions,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_transactions,
			COALESCE(SUM(amount) FILTER (WHERE status = 'success'), 0) as total_amount,
			COALESCE(SUM(fee_amount) FILTER (WHERE status = 'success'), 0) as total_fees,
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'success'), 0) as total_collected
		FROM online_transactions
		%s
	`, whereClause)

	summary := &models.OnlinePaymentSummary{}
	var totalCollected float64

	err := r.DB.QueryRow(ctx, query, args...).Scan(
		&summary.TotalTransactions,
		&summary.SuccessfulPayments,
		&summary.FailedTransactions,
		&summary.PendingTransactions,
		&summary.TotalAmount,
		&summary.TotalFees,
		&totalCollected,
	)

	if err != nil {
		return nil, err
	}

	return summary, nil
}

// CheckOrderExists checks if an order already exists (for idempotency)
func (r *OnlineTransactionRepository) CheckOrderExists(ctx context.Context, orderID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM online_transactions WHERE razorpay_order_id = $1", orderID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsPaymentProcessed checks if a payment has already been processed
func (r *OnlineTransactionRepository) IsPaymentProcessed(ctx context.Context, orderID string) (bool, error) {
	var status string
	err := r.DB.QueryRow(ctx, "SELECT status FROM online_transactions WHERE razorpay_order_id = $1", orderID).Scan(&status)
	if err != nil {
		return false, err
	}
	return status == string(models.OnlineTxStatusSuccess), nil
}
