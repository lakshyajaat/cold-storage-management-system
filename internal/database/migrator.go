package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Note: We read migrations from the filesystem at runtime
// instead of embedding them, to allow easier updates without recompiling

// Migrator handles database schema migrations
type Migrator struct {
	pool *pgxpool.Pool
}

// NewMigrator creates a new migration runner
//
// Parameters:
//   - pool: PostgreSQL connection pool
//
// Returns:
//   - *Migrator: New migrator instance
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{
		pool: pool,
	}
}

// RunMigrations executes all pending database migrations
//
// This function:
//   1. Creates a migrations tracking table if it doesn't exist
//   2. Reads all migration files from the embedded filesystem
//   3. Skips migrations that have already been run
//   4. Executes new migrations in alphabetical order
//   5. Records successful migrations in the tracking table
//
// Migrations are skipped if:
//   - Filename contains "reset" (destructive operations)
//   - Migration has already been run (tracked in migrations table)
//
// Returns:
//   - error: If any migration fails
func (m *Migrator) RunMigrations(ctx context.Context) error {
	log.Println("Starting database migrations...")

	// Create migrations tracking table if it doesn't exist
	// This table keeps track of which migrations have been run
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migrations that have already been run
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files from filesystem
	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migrations alphabetically to ensure correct execution order
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute each migration
	migrationsRun := 0
	for _, filename := range migrationFiles {
		// Skip reset migrations (destructive operations)
		if strings.Contains(filename, "reset") {
			log.Printf("  ⊘ Skipping: %s (reset script)", filename)
			continue
		}

		// Skip if migration has already been applied
		if appliedMigrations[filename] {
			log.Printf("  ✓ Already applied: %s", filename)
			continue
		}

		// Read migration file content
		filepath := filepath.Join("migrations", filename)
		content, err := os.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute the migration SQL
		log.Printf("  → Running: %s", filename)
		if _, err := m.pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", filename, err)
		}

		// Record successful migration
		if err := m.recordMigration(ctx, filename); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		migrationsRun++
	}

	if migrationsRun > 0 {
		log.Printf("✓ Successfully ran %d new migration(s)", migrationsRun)
	} else {
		log.Println("✓ All migrations already applied - database is up to date")
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
//
// This table tracks which migrations have been applied to prevent re-running them
//
// Schema:
//   - id: Auto-incrementing primary key
//   - filename: Migration filename (unique)
//   - applied_at: Timestamp when migration was applied
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := m.pool.Exec(ctx, query)
	return err
}

// getAppliedMigrations returns a map of all migrations that have been applied
//
// Returns:
//   - map[string]bool: Map where key is filename and value is true if applied
//   - error: If database query fails
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.pool.Query(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

// recordMigration records a successful migration in the tracking table
//
// Parameters:
//   - ctx: Context for database operation
//   - filename: Name of the migration file that was applied
//
// Returns:
//   - error: If database insert fails
func (m *Migrator) recordMigration(ctx context.Context, filename string) error {
	query := `
		INSERT INTO schema_migrations (filename)
		VALUES ($1)
		ON CONFLICT (filename) DO NOTHING
	`

	_, err := m.pool.Exec(ctx, query, filename)
	return err
}
