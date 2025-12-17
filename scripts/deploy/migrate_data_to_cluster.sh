#!/bin/bash

set -e

echo "=========================================="
echo "  Data Migration to K3s Cluster"
echo "=========================================="
echo ""

# Source database (local container)
SOURCE_HOST="localhost"
SOURCE_PORT="5432"
SOURCE_USER="postgres"
SOURCE_PASSWORD="postgres"
SOURCE_DB="cold_db"

# Target database (K3s cluster)
TARGET_HOST="localhost"
TARGET_PORT="15432"  # Will use port-forward
TARGET_USER="postgres"
TARGET_PASSWORD="SecurePostgresPassword123"
TARGET_DB="cold_db"

echo "Source: Local PostgreSQL container"
echo "Target: K3s PostgreSQL cluster"
echo ""

# Check if source database is running
echo "Step 1: Checking source database..."
if docker ps | grep -q postgres; then
    echo "✓ Local PostgreSQL container is running"
else
    echo "✗ Local PostgreSQL container not found"
    echo ""
    echo "Start your local PostgreSQL with:"
    echo "  docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres"
    exit 1
fi

# Test source connection
PGPASSWORD=$SOURCE_PASSWORD psql -h $SOURCE_HOST -p $SOURCE_PORT -U $SOURCE_USER -d $SOURCE_DB -c "SELECT 1;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Source database connection successful"
else
    echo "✗ Cannot connect to source database"
    exit 1
fi

# Setup port-forward to K3s cluster
echo ""
echo "Step 2: Setting up connection to K3s cluster..."
kubectl port-forward svc/cold-postgres-rw 15432:5432 > /dev/null 2>&1 &
PF_PID=$!
sleep 3

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test target connection
PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -c "SELECT 1;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ K3s cluster database connection successful"
else
    echo "✗ Cannot connect to K3s cluster database"
    exit 1
fi

# Export data from source
echo ""
echo "Step 3: Exporting data from local database..."
PGPASSWORD=$SOURCE_PASSWORD pg_dump -h $SOURCE_HOST -p $SOURCE_PORT -U $SOURCE_USER \
    --data-only \
    --inserts \
    --no-owner \
    --no-privileges \
    $SOURCE_DB > /tmp/cold_data_dump.sql

if [ -f /tmp/cold_data_dump.sql ]; then
    SIZE=$(du -h /tmp/cold_data_dump.sql | cut -f1)
    echo "✓ Data exported successfully ($SIZE)"
else
    echo "✗ Export failed"
    exit 1
fi

# Import data to target
echo ""
echo "Step 4: Importing data to K3s cluster..."
PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -f /tmp/cold_data_dump.sql

if [ $? -eq 0 ]; then
    echo "✓ Data imported successfully"
else
    echo "✗ Import failed"
    exit 1
fi

# Cleanup dump file
rm -f /tmp/cold_data_dump.sql

# Verify data
echo ""
echo "Step 5: Verifying data..."
echo ""
echo "Tables in K3s cluster database:"
PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -c "\dt"

echo ""
echo "Row counts:"
for table in users customers entries room_entries gate_passes invoices rent_payments; do
    COUNT=$(PGPASSWORD=$TARGET_PASSWORD psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs)
    if [ ! -z "$COUNT" ]; then
        printf "  %-20s %s rows\n" "$table:" "$COUNT"
    fi
done

echo ""
echo "=========================================="
echo "  ✓ Data Migration Completed"
echo "=========================================="

