#!/bin/bash

set -e

# Database connection details
DB_HOST="cold-postgres-rw"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="SecurePostgresPassword123"
DB_NAME="cold_db"

echo "=========================================="
echo "  Running Database Migrations"
echo "=========================================="
echo ""
echo "Target: K3s PostgreSQL Cluster"
echo "Service: $DB_HOST:$DB_PORT"
echo "Database: $DB_NAME"
echo ""

# Port-forward to access the database from local machine
echo "Setting up port-forward to PostgreSQL..."
kubectl port-forward svc/cold-postgres-rw 5432:5432 > /dev/null 2>&1 &
PF_PID=$!
sleep 3

# Function to cleanup port-forward on exit
cleanup() {
    echo "Cleaning up port-forward..."
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test connection
echo "Testing database connection..."
PGPASSWORD=$DB_PASSWORD psql -h localhost -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Database connection successful"
else
    echo "✗ Database connection failed"
    exit 1
fi

echo ""
echo "Running migrations..."
echo ""

# Run migrations in order
for migration in $(ls -1 migrations/*.sql | sort); do
    filename=$(basename "$migration")
    
    # Skip reset migrations in production
    if [[ "$filename" == "000_reset.sql" ]] || [[ "$filename" == "000_complete_reset.sql" ]]; then
        echo "⊘ Skipping: $filename (reset script)"
        continue
    fi
    
    echo "→ Running: $filename"
    PGPASSWORD=$DB_PASSWORD psql -h localhost -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Success"
    else
        echo "  ✗ Failed"
        echo "  Error running $filename"
        exit 1
    fi
done

echo ""
echo "=========================================="
echo "  ✓ All Migrations Completed Successfully"
echo "=========================================="
echo ""

# Show tables
echo "Database tables:"
PGPASSWORD=$DB_PASSWORD psql -h localhost -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt"

