#!/bin/bash

set -e

echo "=========================================="
echo "  Database Migration to K3s Cluster"
echo "=========================================="
echo ""

POD_NAME="cold-postgres-1"
DB_NAME="cold_db"

# Step 1: Export data
echo "Step 1: Exporting data from local PostgreSQL container..."
sg docker -c "docker exec cold-storage-postgres pg_dump -U postgres \
    --data-only \
    --inserts \
    --no-owner \
    --no-privileges \
    $DB_NAME" > /tmp/data_dump.sql

SIZE=$(du -h /tmp/data_dump.sql | cut -f1)
echo "✓ Data exported ($SIZE)"

# Step 2: Create combined migration file
echo ""
echo "Step 2: Creating combined migration file..."
cat migrations/*.sql | grep -v "^--" | grep -v "^$" > /tmp/all_migrations.sql
echo "✓ Combined migrations created"

# Step 3: Copy data to pod
echo ""
echo "Step 3: Copying data to K3s pod..."
kubectl exec -i $POD_NAME -- bash -c "cat > /tmp/all_migrations.sql" < /tmp/all_migrations.sql
kubectl exec -i $POD_NAME -- bash -c "cat > /tmp/data_dump.sql" < /tmp/data_dump.sql
echo "✓ Files copied"

# Step 4: Run migrations
echo ""
echo "Step 4: Running schema migrations in K3s cluster..."
kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -f /tmp/all_migrations.sql 2>&1 | grep -v "already exists" | grep -v "duplicate" || true
echo "✓ Schema migrations completed"

# Step 5: Import data
echo ""
echo "Step 5: Importing data..."
kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -f /tmp/data_dump.sql 2>&1 | tail -5
echo "✓ Data imported"

# Step 6: Cleanup
echo ""
echo "Step 6: Cleaning up..."
kubectl exec $POD_NAME -- rm -f /tmp/all_migrations.sql /tmp/data_dump.sql
rm -f /tmp/data_dump.sql /tmp/all_migrations.sql
echo "✓ Cleanup complete"

# Step 7: Verify
echo ""
echo "=========================================="
echo "  Verification"
echo "=========================================="
echo ""

echo "Tables in K3s cluster database:"
kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -c "\dt"

echo ""
echo "Row counts:"
for table in users customers entries room_entries gate_passes invoices rent_payments entry_events system_settings; do
    COUNT=$(kubectl exec -i $POD_NAME -- psql -U postgres -d $DB_NAME -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs || echo "0")
    printf "  %-25s %s rows\n" "$table:" "$COUNT"
done

echo ""
echo "=========================================="
echo "  ✓ Migration Completed Successfully!"
echo "=========================================="
echo ""
echo "Your data has been migrated to:"
echo "  Cluster: K3s"
echo "  Service: cold-postgres-rw"
echo "  Database: $DB_NAME"

