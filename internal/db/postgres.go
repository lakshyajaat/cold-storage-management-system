package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"cold-backend/internal/config"
)

func Connect(cfg *config.Config) *pgxpool.Pool {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	// Run migrations
	if err := RunMigrations(pool); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	return pool
}

// RunMigrations creates necessary tables if they don't exist
func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// Create login_logs table
	loginLogsTable := `
		CREATE TABLE IF NOT EXISTS login_logs (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			login_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			logout_time TIMESTAMP,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_login_logs_user_id ON login_logs(user_id);
		CREATE INDEX IF NOT EXISTS idx_login_logs_login_time ON login_logs(login_time);
	`

	if _, err := pool.Exec(ctx, loginLogsTable); err != nil {
		return fmt.Errorf("failed to create login_logs table: %w", err)
	}

	// Create room_entry_edit_logs table
	editLogsTable := `
		CREATE TABLE IF NOT EXISTS room_entry_edit_logs (
			id SERIAL PRIMARY KEY,
			room_entry_id INTEGER NOT NULL REFERENCES room_entries(id) ON DELETE CASCADE,
			edited_by_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			old_room_no VARCHAR(10),
			new_room_no VARCHAR(10),
			old_floor VARCHAR(10),
			new_floor VARCHAR(10),
			old_gate_no VARCHAR(50),
			new_gate_no VARCHAR(50),
			old_quantity INTEGER,
			new_quantity INTEGER,
			old_remark TEXT,
			new_remark TEXT,
			edited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_room_entry_edit_logs_room_entry_id ON room_entry_edit_logs(room_entry_id);
		CREATE INDEX IF NOT EXISTS idx_room_entry_edit_logs_edited_by ON room_entry_edit_logs(edited_by_user_id);
	`

	if _, err := pool.Exec(ctx, editLogsTable); err != nil {
		return fmt.Errorf("failed to create room_entry_edit_logs table: %w", err)
	}

	// Create admin_action_logs table
	adminActionLogsTable := `
		CREATE TABLE IF NOT EXISTS admin_action_logs (
			id SERIAL PRIMARY KEY,
			admin_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			action_type VARCHAR(50) NOT NULL,
			target_type VARCHAR(50) NOT NULL,
			target_id INTEGER,
			description TEXT NOT NULL,
			old_value TEXT,
			new_value TEXT,
			ip_address VARCHAR(45),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_admin_action_logs_admin_user_id ON admin_action_logs(admin_user_id);
		CREATE INDEX IF NOT EXISTS idx_admin_action_logs_created_at ON admin_action_logs(created_at);
		CREATE INDEX IF NOT EXISTS idx_admin_action_logs_action_type ON admin_action_logs(action_type);
	`

	if _, err := pool.Exec(ctx, adminActionLogsTable); err != nil {
		return fmt.Errorf("failed to create admin_action_logs table: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
