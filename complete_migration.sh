#!/bin/bash

set -e

echo "=========================================="
echo "  Complete Database Migration Workflow"
echo "=========================================="
echo ""
echo "This script will:"
echo "1. Run schema migrations (create tables)"
echo "2. Export data from local PostgreSQL"
echo "3. Import data to K3s cluster"
echo ""

# K3s cluster database credentials
TARGET_USER="postgres"
TARGET_PASSWORD="SecurePostgresPassword123"
TARGET_DB="cold_db"

# Setup port-forward
echo "Setting up connection to K3s cluster..."
kubectl port-forward svc/cold-postgres-rw 15432:5432 > /dev/null 2>&1 &
PF_PID=$!
sleep 3

cleanup() {
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Step 1: Run schema migrations
echo ""
echo "=========================================="
echo "Step 1: Running Schema Migrations"
echo "=========================================="
echo ""

for migration in $(ls -1 migrations/*.sql | sort); do
    filename=$(basename "$migration")
    
    # Skip reset migrations
    if [[ "$filename" == "000_reset.sql" ]] || [[ "$filename" == "000_complete_reset.sql" ]]; then
        echo "⊘ Skipping: $filename (reset script)"
        continue
    fi
    
    echo "→ Running: $filename"
    PGPASSWORD=$TARGET_PASSWORD psql -h localhost -p 15432 -U $TARGET_USER -d $TARGET_DB \
        -f "$migration" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Success"
    else
        echo "  ⚠ Warning: $filename (table may already exist)"
    fi
done

echo ""
echo "✓ Schema migrations completed"

# Step 2: Check if local database exists
echo ""
echo "=========================================="
echo "Step 2: Checking for Local Data"
echo "=========================================="
echo ""

if docker ps | grep -q postgres; then
    echo "✓ Local PostgreSQL container found"
    
    # Try to connect to local database
    if PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d cold_db -c "SELECT 1;" > /dev/null 2>&1; then
        echo "✓ Local database accessible"
        
        # Ask user if they want to migrate data
        echo ""
        echo "Do you want to migrate data from local database to K3s cluster? (yes/no)"
        read -p "> " MIGRATE_DATA
        
        if [[ "$MIGRATE_DATA" == "yes" || "$MIGRATE_DATA" == "y" ]]; then
            echo ""
            echo "Step 3: Migrating Data..."
            
            # Export data
            echo "  - Exporting data from local database..."
            PGPASSWORD=postgres pg_dump -h localhost -p 5432 -U postgres \
                --data-only --inserts --no-owner --no-privileges \
                cold_db > /tmp/cold_data_dump.sql
            
            # Import data
            echo "  - Importing data to K3s cluster..."
            PGPASSWORD=$TARGET_PASSWORD psql -h localhost -p 15432 -U $TARGET_USER -d $TARGET_DB \
                -f /tmp/cold_data_dump.sql
            
            rm -f /tmp/cold_data_dump.sql
            echo "  ✓ Data migration completed"
        else
            echo "Skipping data migration"
        fi
    else
        echo "⚠ Cannot connect to local database (no data to migrate)"
    fi
else
    echo "⚠ No local PostgreSQL container found (no data to migrate)"
fi

# Step 3: Verify
echo ""
echo "=========================================="
echo "Final Status"
echo "=========================================="
echo ""

echo "Tables in K3s cluster:"
PGPASSWORD=$TARGET_PASSWORD psql -h localhost -p 15432 -U $TARGET_USER -d $TARGET_DB -c "\dt"

echo ""
echo "Row counts:"
for table in users customers entries room_entries gate_passes invoices rent_payments; do
    COUNT=$(PGPASSWORD=$TARGET_PASSWORD psql -h localhost -p 15432 -U $TARGET_USER -d $TARGET_DB \
        -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs)
    if [ ! -z "$COUNT" ]; then
        printf "  %-20s %s rows\n" "$table:" "$COUNT"
    fi
done

echo ""
echo "=========================================="
echo "  ✓ Migration Workflow Completed"
echo "=========================================="
echo ""
echo "K3s cluster database is ready!"
echo "Connection: cold-postgres-rw.default.svc.cluster.local:5432"

