package repositories

import (
	"context"
	"fmt"

	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryRepository struct {
	DB *pgxpool.Pool
}

func NewEntryRepository(db *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{DB: db}
}

func (r *EntryRepository) Create(ctx context.Context, e *models.Entry) error {
	// Use COUNT-based logic for thock numbers
	// This ensures counters auto-reset when entries are deleted (season reset)

	var nextNumber int
	var err error

	if e.ThockCategory == "seed" {
		// SEED: starts at 1, format 0001/quantity
		err = r.DB.QueryRow(ctx, "SELECT COALESCE(COUNT(*), 0) + 1 FROM entries WHERE thock_category = 'seed'").Scan(&nextNumber)
	} else if e.ThockCategory == "sell" {
		// SELL: starts at 1501, format 1501/quantity
		err = r.DB.QueryRow(ctx, "SELECT COALESCE(COUNT(*), 0) + 1501 FROM entries WHERE thock_category = 'sell'").Scan(&nextNumber)
	} else {
		return fmt.Errorf("invalid thock category: %s", e.ThockCategory)
	}

	if err != nil {
		return fmt.Errorf("failed to get next thock number: %w", err)
	}

	// Generate thock number: NO/QUANTITY format
	// SEED: 0001/quantity format (4-digit padded)
	// SELL: 1501+/quantity format
	var thockNumber string
	if e.ThockCategory == "seed" {
		thockNumber = fmt.Sprintf("%04d/%d", nextNumber, e.ExpectedQuantity)
	} else {
		thockNumber = fmt.Sprintf("%d/%d", nextNumber, e.ExpectedQuantity)
	}
	e.ThockNumber = thockNumber

	return r.DB.QueryRow(ctx,
		`INSERT INTO entries(customer_id, phone, name, village, so, expected_quantity, thock_category, thock_number, created_by_user_id)
         VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
         RETURNING id, created_at, updated_at`,
		e.CustomerID, e.Phone, e.Name, e.Village, e.SO, e.ExpectedQuantity, e.ThockCategory, e.ThockNumber, e.CreatedByUserID,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *EntryRepository) Get(ctx context.Context, id int) (*models.Entry, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, customer_id, phone, name, village, so, expected_quantity, thock_category, thock_number, created_by_user_id, created_at, updated_at
         FROM entries WHERE id=$1`, id)

	var entry models.Entry
	err := row.Scan(&entry.ID, &entry.CustomerID, &entry.Phone, &entry.Name, &entry.Village, &entry.SO,
		&entry.ExpectedQuantity, &entry.ThockCategory, &entry.ThockNumber, &entry.CreatedByUserID,
		&entry.CreatedAt, &entry.UpdatedAt)
	return &entry, err
}

func (r *EntryRepository) List(ctx context.Context) ([]*models.Entry, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, customer_id, phone, name, village, so, expected_quantity, thock_category, thock_number, created_by_user_id, created_at, updated_at
         FROM entries ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		var entry models.Entry
		err := rows.Scan(&entry.ID, &entry.CustomerID, &entry.Phone, &entry.Name, &entry.Village, &entry.SO,
			&entry.ExpectedQuantity, &entry.ThockCategory, &entry.ThockNumber, &entry.CreatedByUserID,
			&entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	return entries, nil
}

func (r *EntryRepository) ListByCustomer(ctx context.Context, customerID int) ([]*models.Entry, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, customer_id, phone, name, village, so, expected_quantity, thock_category, thock_number, created_by_user_id, created_at, updated_at
         FROM entries WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		var entry models.Entry
		err := rows.Scan(&entry.ID, &entry.CustomerID, &entry.Phone, &entry.Name, &entry.Village, &entry.SO,
			&entry.ExpectedQuantity, &entry.ThockCategory, &entry.ThockNumber, &entry.CreatedByUserID,
			&entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	return entries, nil
}

func (r *EntryRepository) GetCountByCategory(ctx context.Context, category string) (int, error) {
	// Return actual COUNT of entries for this category
	// This matches the thock number generation logic
	if category != "seed" && category != "sell" {
		return 0, fmt.Errorf("invalid category: %s", category)
	}

	var count int
	err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM entries WHERE thock_category = $1", category).Scan(&count)
	return count, err
}

func (r *EntryRepository) ListUnassigned(ctx context.Context) ([]*models.Entry, error) {
	// Get entries that don't have a room entry yet
	rows, err := r.DB.Query(ctx,
		`SELECT e.id, e.customer_id, e.phone, e.name, e.village, e.so, e.expected_quantity,
		        e.thock_category, e.thock_number, e.created_by_user_id, e.created_at, e.updated_at
         FROM entries e
         LEFT JOIN room_entries re ON e.id = re.entry_id
         WHERE re.id IS NULL
         ORDER BY e.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		var entry models.Entry
		err := rows.Scan(&entry.ID, &entry.CustomerID, &entry.Phone, &entry.Name, &entry.Village, &entry.SO,
			&entry.ExpectedQuantity, &entry.ThockCategory, &entry.ThockNumber, &entry.CreatedByUserID,
			&entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}
	return entries, nil
}

// GetByThockNumber retrieves an entry by thock number
func (r *EntryRepository) GetByThockNumber(ctx context.Context, thockNumber string) (*models.Entry, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, customer_id, phone, name, village, so, expected_quantity, thock_category, thock_number, created_by_user_id, created_at, updated_at
         FROM entries WHERE thock_number=$1`, thockNumber)

	var entry models.Entry
	err := row.Scan(&entry.ID, &entry.CustomerID, &entry.Phone, &entry.Name, &entry.Village, &entry.SO,
		&entry.ExpectedQuantity, &entry.ThockCategory, &entry.ThockNumber, &entry.CreatedByUserID,
		&entry.CreatedAt, &entry.UpdatedAt)
	return &entry, err
}
