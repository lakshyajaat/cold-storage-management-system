#!/bin/bash
# Database Restore Script with Replica Resync
# This script restores the database and reinitializes all replicas
# Usage: ./restore_db.sh /path/to/backup/folder

set -e

BACKUP_DIR="${1:-}"
PRIMARY_HOST="192.168.15.120"
REPLICA_HOSTS=("192.168.15.121" "192.168.15.122")
REPLICATOR_PASSWORD="replicator123"
K8S_HOST="192.168.15.110"

if [ -z "$BACKUP_DIR" ]; then
    echo "Usage: $0 /path/to/backup/folder"
    echo "Example: $0 /home/lakshya/backup/postgresql/base/backup_20251228_220001"
    exit 1
fi

if [ ! -f "$BACKUP_DIR/schema.sql" ] || [ ! -f "$BACKUP_DIR/data.sql" ]; then
    echo "Error: Backup folder must contain schema.sql and data.sql"
    exit 1
fi

echo "=== Starting Database Restore ==="
echo "Backup: $BACKUP_DIR"
echo "Primary: $PRIMARY_HOST"
echo "Replicas: ${REPLICA_HOSTS[*]}"
echo ""

# Step 1: Backup current state
echo "[1/6] Creating backup of current state..."
ssh root@$PRIMARY_HOST "sudo -u postgres pg_dump cold_db > /tmp/cold_db_before_restore_\$(date +%Y%m%d_%H%M%S).sql"

# Step 2: Copy backup files to primary
echo "[2/6] Copying backup files to primary..."
scp "$BACKUP_DIR/schema.sql" root@$PRIMARY_HOST:/tmp/restore_schema.sql
scp "$BACKUP_DIR/data.sql" root@$PRIMARY_HOST:/tmp/restore_data.sql

# Step 3: Restore database on primary
echo "[3/6] Restoring database on primary..."
ssh root@$PRIMARY_HOST "sudo -u postgres psql -c \"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'cold_db' AND pid <> pg_backend_pid();\" && \
    sleep 1 && \
    sudo -u postgres dropdb --force cold_db && \
    sudo -u postgres createdb cold_db && \
    sudo -u postgres psql cold_db < /tmp/restore_schema.sql > /dev/null && \
    sudo -u postgres psql cold_db < /tmp/restore_data.sql > /dev/null && \
    sudo -u postgres psql -c \"ALTER USER replicator WITH PASSWORD '$REPLICATOR_PASSWORD';\" && \
    echo 'Primary restored'"

# Step 4: Reinitialize replicas
echo "[4/6] Reinitializing replicas..."
for host in "${REPLICA_HOSTS[@]}"; do
    echo "  Reinitializing $host..."
    ssh root@$host "echo '$PRIMARY_HOST:5432:*:replicator:$REPLICATOR_PASSWORD' > /var/lib/postgresql/.pgpass && \
        chown postgres:postgres /var/lib/postgresql/.pgpass && \
        chmod 600 /var/lib/postgresql/.pgpass && \
        systemctl stop postgresql@17-main && \
        rm -rf /var/lib/postgresql/17/main/* && \
        sudo -u postgres pg_basebackup -h $PRIMARY_HOST -D /var/lib/postgresql/17/main -U replicator -P -R -X stream && \
        systemctl start postgresql@17-main && \
        echo '  $host done'"
done

# Step 5: Clear Redis cache
echo "[5/6] Clearing Redis cache..."
ssh root@$PRIMARY_HOST "redis-cli FLUSHALL"

# Step 6: Restart K8s deployments
echo "[6/6] Restarting application pods..."
ssh root@$K8S_HOST "kubectl rollout restart deployment cold-backend-customer cold-backend-employee"

echo ""
echo "=== Restore Complete ==="
echo "Verify with: ssh root@$PRIMARY_HOST 'sudo -u postgres psql cold_db -c \"SELECT COUNT(*) FROM customers;\"'"
