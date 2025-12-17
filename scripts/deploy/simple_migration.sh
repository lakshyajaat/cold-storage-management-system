#!/bin/bash

set -e

echo "=========================================="
echo "  Simple Database Migration"
echo "=========================================="
echo ""

POD_NAME="cold-postgres-1"
DB_NAME="cold_db"

# Step 1: Export data from local PostgreSQL
echo "Step 1: Exporting data from local PostgreSQL..."
sg docker -c "docker exec cold-storage-postgres pg_dump -U postgres \
    --data-only \
    --inserts \
    --no-owner \
    --no-privileges \
    $DB_NAME" > /tmp/data_dump.sql

if [ -f /tmp/data_dump.sql ]; then
    SIZE=$(du -h /tmp/data_dump.sql | cut -f1)
    echo "✓ Data exported ($SIZE)"
else
    echo "✗ Export failed"
    exit 1
fi

# Step 2: Copy migrations to pod
echo ""
echo "Step 2: Copying files to K3s pod..."
kubectl cp migrations/ $POD_NAME:/tmp/migrations/
kubectl cp /tmp/data_dump.sql $POD_NAME:/tmp/data_dump.sql
echo "✓ Files copied to pod"

# Step 3: Run schema migrations
echo ""
echo "Step 3: Running schema migrations..."
for migration in $(ls -1 migrations/*.sql | sort); do
    filename=$(basename "$migration")
    
    # Skip reset migrations
    if [[ "$filename" == "000_reset.sql" ]] || [[ "$filename" == "000_complete_reset.sql" ]]; then
        continue
    fi
    
    echo "  → $filename"
    kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -f /tmp/migrations/$filename > /dev/null 2>&1 || echo "    (may already exist)"
done
echo "✓ Schema migrations completed"

# Step 4: Import data
echo ""
echo "Step 4: Importing data..."
kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -f /tmp/data_dump.sql > /dev/null 2>&1
echo "✓ Data imported"

# Step 5: Cleanup
echo ""
echo "Step 5: Cleaning up..."
kubectl exec $POD_NAME -- rm -rf /tmp/migrations /tmp/data_dump.sql
rm -f /tmp/data_dump.sql
echo "✓ Cleanup complete"

# Step 6: Verify
echo ""
echo "Step 6: Verification"
echo "===================="
echo ""

echo "Tables:"
kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -c "\dt"

echo ""
echo "Row counts:"
for table in users customers entries room_entries gate_passes invoices rent_payments entry_events; do
    COUNT=$(kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs)
    if [ ! -z "$COUNT" ]; then
        printf "  %-20s %s rows\n" "$table:" "$COUNT"
    fi
done

echo ""
echo "=========================================="
echo "  ✓ Migration Complete!"
echo "=========================================="

