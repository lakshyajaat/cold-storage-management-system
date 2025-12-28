package repositories

import (
	"context"
	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	if u.Role == "" {
		u.Role = "employee" // Default role
	}
	if !u.IsActive {
		u.IsActive = true // Default to active
	}
	return r.DB.QueryRow(ctx,
		`INSERT INTO users(name, email, password_hash, role, has_accountant_access, can_manage_entries, is_active)
         VALUES($1, $2, $3, $4, $5, $6, $7)
         RETURNING id, created_at, updated_at`,
		u.Name, u.Email, u.PasswordHash, u.Role, u.HasAccountantAccess, u.CanManageEntries, u.IsActive,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) Get(ctx context.Context, id int) (*models.User, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, has_accountant_access, can_manage_entries, is_active, created_at, updated_at
         FROM users WHERE id=$1`, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.HasAccountantAccess, &user.CanManageEntries, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role, has_accountant_access, can_manage_entries, is_active, created_at, updated_at
         FROM users WHERE email=$1`, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.HasAccountantAccess, &user.CanManageEntries, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

// List returns all users
func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, name, email, role, has_accountant_access, can_manage_entries, is_active, created_at, updated_at
         FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role,
			&user.HasAccountantAccess, &user.CanManageEntries, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	// If password is empty, don't update it (keep existing password)
	if u.PasswordHash != "" {
		_, err := r.DB.Exec(ctx,
			`UPDATE users SET name=$1, email=$2, password_hash=$3, role=$4, has_accountant_access=$5, can_manage_entries=$6, updated_at=CURRENT_TIMESTAMP
			 WHERE id=$7`,
			u.Name, u.Email, u.PasswordHash, u.Role, u.HasAccountantAccess, u.CanManageEntries, u.ID)
		return err
	}

	// Update without changing password
	_, err := r.DB.Exec(ctx,
		`UPDATE users SET name=$1, email=$2, role=$3, has_accountant_access=$4, can_manage_entries=$5, updated_at=CURRENT_TIMESTAMP
         WHERE id=$6`,
		u.Name, u.Email, u.Role, u.HasAccountantAccess, u.CanManageEntries, u.ID)
	return err
}

// ToggleActiveStatus toggles the is_active status of a user
func (r *UserRepository) ToggleActiveStatus(ctx context.Context, userID int, isActive bool) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE users SET is_active=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`,
		isActive, userID)
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}
