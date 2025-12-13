#!/bin/bash

# Load environment variables
export $(grep -v '^#' .env | xargs)

# PostgreSQL connection details
PGHOST=${DB_HOST:-localhost}
PGPORT=${DB_PORT:-5432}
PGUSER=${DB_USER:-postgres}
PGPASSWORD=${DB_PASSWORD:-postgres}
PGDATABASE=${DB_NAME:-cold_db}

export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

echo "Running migrations on $PGDATABASE..."

# Run migrations in order
for migration in migrations/*.sql; do
    echo "Running $migration..."
    psql -f "$migration"
    if [ $? -ne 0 ]; then
        echo "Migration $migration failed!"
        exit 1
    fi
done

echo "All migrations completed successfully!"
