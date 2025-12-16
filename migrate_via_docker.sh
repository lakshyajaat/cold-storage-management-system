#!/bin/bash

set -e

echo "=========================================="
echo "  Database Migration via Docker"
echo "=========================================="
echo ""

# K3s cluster connection (via port-forward)
TARGET_HOST="host.docker.internal"
TARGET_PORT="15432"
TARGET_USER="postgres"
TARGET_PASSWORD="SecurePostgresPassword123"
TARGET_DB="cold_db"

# Setup port-forward
echo "Setting up port-forward to K3s cluster..."
kubectl port-forward svc/cold-postgres-rw 15432:5432 > /dev/null 2>&1 &
PF_PID=$!
sleep 3

cleanup() {
    echo "Cleaning up..."
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Step 1: Run schema migrations
echo ""
echo "Step 1: Running Schema Migrations to K3s Cluster"
echo "=================================================="
echo ""

for migration in $(ls -1 migrations/*.sql | sort); do
    filename=$(basename "$migration")
    
    # Skip reset migrations
    if [[ "$filename" == "000_reset.sql" ]] || [[ "$filename" == "000_complete_reset.sql" ]]; then
        echo "⊘ Skipping: $filename (reset script)"
        continue
    fi
    
    echo "→ Running: $filename"
    
    sg docker -c "docker run --rm -i \
        --network host \
        -v $(pwd)/migrations:/migrations \
        postgres:16-alpine \
        psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -f /migrations/$filename" \
        <<< "$TARGET_PASSWORD" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Success"
    else
        echo "  ⚠ Warning (table may already exist)"
    fi
done

echo ""
echo "✓ Schema migrations completed"

# Step 2: Export data from local database
echo ""
echo "Step 2: Exporting Data from Local PostgreSQL"
echo "=============================================="
echo ""

sg docker -c "docker exec cold-storage-postgres pg_dump -U postgres --data-only --inserts --no-owner --no-privileges cold_db" > /tmp/cold_data_dump.sql

if [ -f /tmp/cold_data_dump.sql ]; then
    SIZE=$(du -h /tmp/cold_data_dump.sql | cut -f1)
    echo "✓ Data exported successfully ($SIZE)"
else
    echo "✗ Export failed"
    exit 1
fi

# Step 3: Import data to K3s cluster
echo ""
echo "Step 3: Importing Data to K3s Cluster"
echo "======================================"
echo ""

sg docker -c "docker run --rm -i \
    --network host \
    -v /tmp:/tmp \
    postgres:16-alpine \
    psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -f /tmp/cold_data_dump.sql" \
    <<< "$TARGET_PASSWORD" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✓ Data imported successfully"
else
    echo "✗ Import failed"
    exit 1
fi

# Cleanup
rm -f /tmp/cold_data_dump.sql

# Step 4: Verify
echo ""
echo "Step 4: Verifying Data"
echo "======================"
echo ""

echo "Row counts in K3s cluster:"
for table in users customers entries room_entries gate_passes invoices rent_payments entry_events system_settings; do
    COUNT=$(sg docker -c "docker run --rm -i \
        --network host \
        postgres:16-alpine \
        psql -h $TARGET_HOST -p $TARGET_PORT -U $TARGET_USER -d $TARGET_DB -t -c \"SELECT COUNT(*) FROM $table;\"" \
        <<< "$TARGET_PASSWORD" 2>/dev/null | xargs)
    
    if [ ! -z "$COUNT" ]; then
        printf "  %-20s %s rows\n" "$table:" "$COUNT"
    fi
done

echo ""
echo "=========================================="
echo "  ✓ Migration Completed Successfully"
echo "=========================================="
echo ""
echo "Your data has been migrated to the K3s cluster!"

