package repositories

import (
	"context"
	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RentPaymentRepository struct {
	DB *pgxpool.Pool
}

func NewRentPaymentRepository(db *pgxpool.Pool) *RentPaymentRepository {
	return &RentPaymentRepository{DB: db}
}

func (r *RentPaymentRepository) Create(ctx context.Context, payment *models.RentPayment) error {
	query := `
		INSERT INTO rent_payments (entry_id, customer_name, customer_phone, total_rent, amount_paid, balance, processed_by_user_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, payment_date, created_at
	`

	err := r.DB.QueryRow(ctx, query,
		payment.EntryID,
		payment.CustomerName,
		payment.CustomerPhone,
		payment.TotalRent,
		payment.AmountPaid,
		payment.Balance,
		payment.ProcessedByUserID,
		payment.Notes,
	).Scan(&payment.ID, &payment.PaymentDate, &payment.CreatedAt)

	return err
}

func (r *RentPaymentRepository) GetByEntryID(ctx context.Context, entryID int) ([]*models.RentPayment, error) {
	query := `
		SELECT id, entry_id, customer_name, customer_phone, total_rent, amount_paid, balance,
		       payment_date, COALESCE(processed_by_user_id, 0), COALESCE(notes, ''), created_at
		FROM rent_payments
		WHERE entry_id = $1
		ORDER BY payment_date DESC
	`

	rows, err := r.DB.Query(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.RentPayment
	for rows.Next() {
		payment := &models.RentPayment{}
		err := rows.Scan(
			&payment.ID,
			&payment.EntryID,
			&payment.CustomerName,
			&payment.CustomerPhone,
			&payment.TotalRent,
			&payment.AmountPaid,
			&payment.Balance,
			&payment.PaymentDate,
			&payment.ProcessedByUserID,
			&payment.Notes,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *RentPaymentRepository) GetByPhone(ctx context.Context, phone string) ([]*models.RentPayment, error) {
	query := `
		SELECT id, entry_id, customer_name, customer_phone, total_rent, amount_paid, balance,
		       payment_date, COALESCE(processed_by_user_id, 0), COALESCE(notes, ''), created_at
		FROM rent_payments
		WHERE customer_phone = $1
		ORDER BY payment_date DESC
	`

	rows, err := r.DB.Query(ctx, query, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.RentPayment
	for rows.Next() {
		payment := &models.RentPayment{}
		err := rows.Scan(
			&payment.ID,
			&payment.EntryID,
			&payment.CustomerName,
			&payment.CustomerPhone,
			&payment.TotalRent,
			&payment.AmountPaid,
			&payment.Balance,
			&payment.PaymentDate,
			&payment.ProcessedByUserID,
			&payment.Notes,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *RentPaymentRepository) List(ctx context.Context) ([]*models.RentPayment, error) {
	query := `
		SELECT id, entry_id, customer_name, customer_phone, total_rent, amount_paid, balance,
		       payment_date, COALESCE(processed_by_user_id, 0), COALESCE(notes, ''), created_at
		FROM rent_payments
		ORDER BY payment_date DESC
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.RentPayment
	for rows.Next() {
		payment := &models.RentPayment{}
		err := rows.Scan(
			&payment.ID,
			&payment.EntryID,
			&payment.CustomerName,
			&payment.CustomerPhone,
			&payment.TotalRent,
			&payment.AmountPaid,
			&payment.Balance,
			&payment.PaymentDate,
			&payment.ProcessedByUserID,
			&payment.Notes,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}
