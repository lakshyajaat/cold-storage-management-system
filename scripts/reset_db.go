package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   Reset Database for Testing")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("⚠️  WARNING: This will DELETE ALL USER DATA!")
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("  - Delete all users (except admin)")
	fmt.Println("  - Delete all customers")
	fmt.Println("  - Delete all entries")
	fmt.Println("  - Delete all room entries")
	fmt.Println("  - Delete all payments")
	fmt.Println("  - Reset all ID sequences")
	fmt.Println()
	fmt.Print("Type 'yes' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("Reset cancelled.")
		return
	}

	// Load environment variables
	godotenv.Load()

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "cold_db")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println()
	fmt.Println("🔄 Resetting database...")

	ctx := context.Background()

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v\n", err)
	}
	defer tx.Rollback(ctx)

	// Disable foreign key checks
	_, err = tx.Exec(ctx, "SET session_replication_role = 'replica'")
	if err != nil {
		log.Fatalf("Failed to disable foreign key checks: %v\n", err)
	}

	// Truncate all tables
	tables := []string{
		"rent_payments",
		"entry_events",
		"room_entries",
		"entries",
		"customers",
		"users",
		"system_settings",
	}

	for _, table := range tables {
		_, err = tx.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Fatalf("Failed to truncate %s: %v\n", table, err)
		}
		fmt.Printf("  ✓ Cleared %s\n", table)
	}

	// Re-enable foreign key checks
	_, err = tx.Exec(ctx, "SET session_replication_role = 'origin'")
	if err != nil {
		log.Fatalf("Failed to enable foreign key checks: %v\n", err)
	}

	// Reset sequences
	sequences := []string{
		"users_id_seq",
		"customers_id_seq",
		"entries_id_seq",
		"entry_events_id_seq",
		"room_entries_id_seq",
		"rent_payments_id_seq",
		"system_settings_id_seq",
	}

	for _, seq := range sequences {
		_, err = tx.Exec(ctx, fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", seq))
		if err != nil {
			log.Printf("Warning: Failed to reset sequence %s: %v\n", seq, err)
		}
	}
	fmt.Println("  ✓ Reset ID sequences")

	// Create default admin user
	// Password: admin123
	_, err = tx.Exec(ctx, `
		INSERT INTO users (email, password_hash, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		"admin@cold.com",
		"$2a$10$N9qo8uLOickgx2ZMRZoMye7U4hWJQbFlLwt7xW.hQOKvH8QhPVN8S",
		"Administrator",
		"admin",
	)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v\n", err)
	}
	fmt.Println("  ✓ Created admin user")

	// Create default system settings
	settings := []struct {
		key   string
		value string
		desc  string
	}{
		{"rent_per_item", "10.00", "Rent price per item stored"},
		{"company_name", "Cold Storage Solutions", "Company name for receipts"},
		{"company_address", "123 Main Street, City, State", "Company address for receipts"},
		{"company_phone", "+91-1234567890", "Company phone number"},
	}

	for _, s := range settings {
		_, err = tx.Exec(ctx, `
			INSERT INTO system_settings (setting_key, setting_value, description, updated_at)
			VALUES ($1, $2, $3, NOW())`,
			s.key, s.value, s.desc,
		)
		if err != nil {
			log.Printf("Warning: Failed to create setting %s: %v\n", s.key, err)
		}
	}
	fmt.Println("  ✓ Created default settings")

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Database reset successful!")
	fmt.Println()
	fmt.Println("Default credentials:")
	fmt.Println("  Email:    admin@cold.com")
	fmt.Println("  Password: admin123")
	fmt.Println()
	fmt.Println("Database is now ready for testing!")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
