package repositories

import (
	"context"

	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository struct {
	DB *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{DB: db}
}

func (r *CustomerRepository) Create(ctx context.Context, c *models.Customer) error {
	return r.DB.QueryRow(ctx,
		`INSERT INTO customers(name, phone, so, village, address)
         VALUES($1, $2, $3, $4, $5)
         RETURNING id, created_at, updated_at`,
		c.Name, c.Phone, c.SO, c.Village, c.Address,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *CustomerRepository) Get(ctx context.Context, id int) (*models.Customer, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, name, phone, COALESCE(so, '') as so, village, address,
		 COALESCE(status, 'active') as status, merged_into_customer_id, merged_at,
		 created_at, updated_at
         FROM customers WHERE id=$1`, id)

	var customer models.Customer
	err := row.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.SO, &customer.Village,
		&customer.Address, &customer.Status, &customer.MergedIntoCustomerID, &customer.MergedAt,
		&customer.CreatedAt, &customer.UpdatedAt)
	return &customer, err
}

func (r *CustomerRepository) GetByPhone(ctx context.Context, phone string) (*models.Customer, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, name, phone, COALESCE(so, '') as so, village, address,
		 COALESCE(status, 'active') as status, merged_into_customer_id, merged_at,
		 created_at, updated_at
         FROM customers WHERE phone=$1`, phone)

	var customer models.Customer
	err := row.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.SO, &customer.Village,
		&customer.Address, &customer.Status, &customer.MergedIntoCustomerID, &customer.MergedAt,
		&customer.CreatedAt, &customer.UpdatedAt)
	return &customer, err
}

func (r *CustomerRepository) List(ctx context.Context) ([]*models.Customer, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, name, phone, COALESCE(so, '') as so, village, address,
		 COALESCE(status, 'active') as status, merged_into_customer_id, merged_at,
		 created_at, updated_at
         FROM customers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.SO, &customer.Village,
			&customer.Address, &customer.Status, &customer.MergedIntoCustomerID, &customer.MergedAt,
			&customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		customers = append(customers, &customer)
	}
	return customers, nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *models.Customer) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE customers SET name=$1, phone=$2, so=$3, village=$4, address=$5, updated_at=CURRENT_TIMESTAMP
         WHERE id=$6`,
		c.Name, c.Phone, c.SO, c.Village, c.Address, c.ID)
	return err
}

func (r *CustomerRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM customers WHERE id=$1`, id)
	return err
}

// GetEntryCount returns the number of entries for a customer
func (r *CustomerRepository) GetEntryCount(ctx context.Context, customerID int) (int, error) {
	var count int
	err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE customer_id=$1`, customerID).Scan(&count)
	return count, err
}

// MergeCustomers moves all entries and payments from source customer to target customer
// Instead of deleting, marks source customer as 'merged' for audit trail
// Returns the number of entries moved
func (r *CustomerRepository) MergeCustomers(ctx context.Context, sourceID, targetID int, targetName, targetPhone, targetVillage, targetSO string, sourcePhone string) (int, error) {
	// Start transaction
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Count entries to be moved (by customer_id)
	var entriesMoved int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM entries WHERE customer_id=$1`, sourceID).Scan(&entriesMoved)
	if err != nil {
		return 0, err
	}

	// Also count orphaned entries with source phone (entries with invalid customer_id but matching phone)
	var orphanedCount int
	tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM entries
		WHERE phone=$1 AND customer_id NOT IN (SELECT id FROM customers)`,
		sourcePhone).Scan(&orphanedCount)
	entriesMoved += orphanedCount

	// Move all entries from source to target (update customer_id and denormalized fields)
	_, err = tx.Exec(ctx, `
		UPDATE entries
		SET customer_id=$1, name=$2, phone=$3, village=$4, so=$5, updated_at=NOW()
		WHERE customer_id=$6`,
		targetID, targetName, targetPhone, targetVillage, targetSO, sourceID)
	if err != nil {
		return 0, err
	}

	// Also move orphaned entries with source phone
	_, err = tx.Exec(ctx, `
		UPDATE entries
		SET customer_id=$1, name=$2, phone=$3, village=$4, so=$5, updated_at=NOW()
		WHERE phone=$6 AND customer_id NOT IN (SELECT id FROM customers)`,
		targetID, targetName, targetPhone, targetVillage, targetSO, sourcePhone)
	if err != nil {
		return 0, err
	}

	// Transfer all rent payments from source customer to target customer
	_, err = tx.Exec(ctx, `
		UPDATE rent_payments
		SET customer_name=$1, customer_phone=$2
		WHERE customer_phone=$3`,
		targetName, targetPhone, sourcePhone)
	if err != nil {
		return 0, err
	}

	// SOFT DELETE: Mark source customer as merged (don't delete)
	_, err = tx.Exec(ctx, `
		UPDATE customers
		SET status='merged', merged_into_customer_id=$1, merged_at=NOW(), updated_at=NOW()
		WHERE id=$2`,
		targetID, sourceID)
	if err != nil {
		return 0, err
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	return entriesMoved, nil
}

// GetMergedCustomers returns all customers that have been merged
func (r *CustomerRepository) GetMergedCustomers(ctx context.Context) ([]*models.Customer, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, name, phone, COALESCE(so, '') as so, village, address,
		 COALESCE(status, 'active') as status, merged_into_customer_id, merged_at,
		 created_at, updated_at
         FROM customers
         WHERE status = 'merged'
         ORDER BY merged_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.SO, &customer.Village,
			&customer.Address, &customer.Status, &customer.MergedIntoCustomerID, &customer.MergedAt,
			&customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		customers = append(customers, &customer)
	}
	return customers, nil
}

// UndoMerge reverses a merge by setting source customer status back to active
func (r *CustomerRepository) UndoMerge(ctx context.Context, sourceCustomerID int) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE customers
		 SET status = 'active', merged_into_customer_id = NULL, merged_at = NULL, updated_at = NOW()
		 WHERE id = $1 AND status = 'merged'`,
		sourceCustomerID)
	return err
}

// FuzzySearchByPhone searches customers by phone number (fuzzy match)
// Only returns active customers (not merged ones)
func (r *CustomerRepository) FuzzySearchByPhone(ctx context.Context, phone string) ([]*models.Customer, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, name, phone, COALESCE(so, '') as so, village, address,
		 COALESCE(status, 'active') as status, merged_into_customer_id, merged_at,
		 created_at, updated_at
         FROM customers
         WHERE phone LIKE $1 AND (status IS NULL OR status = 'active')
         ORDER BY created_at DESC
         LIMIT 10`,
		"%"+phone+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.SO, &customer.Village,
			&customer.Address, &customer.Status, &customer.MergedIntoCustomerID, &customer.MergedAt,
			&customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		customers = append(customers, &customer)
	}
	return customers, nil
}
