# Database Migration Tools

This directory contains tools for managing database migrations and data transfers for the Cold Storage Backend.

## Overview

The Cold Storage Backend uses PostgreSQL databases in multiple environments:
- **Local Development**: PostgreSQL container (`cold-storage-postgres`)
- **K3s Cluster**: CloudNativePG cluster with 5 replicas

These tools help migrate schema and data between environments.

## Tools

### 1. migrate_via_stdin.sh

**Purpose**: Migrate database schema and data from local PostgreSQL to K3s cluster

**Usage**:
```bash
cd /path/to/cold-backend
./scripts/database/migrate_via_stdin.sh
```

**What it does**:
1. Exports data from local PostgreSQL container
2. Runs all schema migrations in K3s cluster
3. Imports data to K3s cluster
4. Verifies the migration

**Requirements**:
- Local PostgreSQL container named `cold-storage-postgres` must be running
- kubectl configured to access K3s cluster
- Docker access (via `sg docker`)

**Output**:
- Displays progress for each migration file
- Shows final row counts for verification
- Reports any errors encountered

---

## Migration Files

Migration SQL files are stored in `/migrations/` directory and are run in alphabetical order.

### Migration Naming Convention

```
[sequence]_[description].sql
```

Example:
- `001_create_users.sql` - Creates users table
- `002_add_auth_to_users.sql` - Adds authentication columns
- `003_seed_admin_user.sql` - Inserts admin user

### Reset Migrations

Files starting with `000_` are reset scripts and are **skipped** during normal migrations:
- `000_reset.sql` - Simple reset
- `000_complete_reset.sql` - Complete database reset

---

## K3s Cluster PostgreSQL Details

### Connection Information

- **Service (Read-Write)**: `cold-postgres-rw` (connects to primary)
- **Service (Read-Only)**: `cold-postgres-ro` (connects to replicas)
- **Port**: 5432
- **Database**: `cold_db`
- **User**: `postgres`
- **Password**: Stored in secret `cold-postgres-superuser`

### PostgreSQL Pods

The K3s cluster runs 5 PostgreSQL pods:
- `cold-postgres-1` - Primary (read-write)
- `cold-postgres-2` - Replica (read-only)
- `cold-postgres-3` - Replica (read-only)
- `cold-postgres-4` - Replica (read-only)
- `cold-postgres-5` - Replica (read-only)

### Get Database Password

```bash
kubectl get secret cold-postgres-superuser -o jsonpath='{.data.password}' | base64 -d
```

---

## Common Tasks

### Connect to K3s PostgreSQL

```bash
# Via kubectl exec (direct to pod)
kubectl exec -it cold-postgres-1 -- psql -U postgres -d cold_db

# Via port-forward (from local machine)
kubectl port-forward svc/cold-postgres-rw 15432:5432
psql -h localhost -p 15432 -U postgres -d cold_db
```

### Run a Single Migration

```bash
kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db < migrations/001_create_users.sql
```

### Export Data from K3s Cluster

```bash
kubectl exec -i cold-postgres-1 -- pg_dump -U postgres cold_db > k3s_backup.sql
```

### Check Row Counts

```bash
kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db -c "
SELECT 
    schemaname,
    tablename, 
    n_live_tup as row_count
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;
"
```

---

## Troubleshooting

### Migration Script Fails

**Problem**: Script exits with error during migration

**Solution**:
1. Check which migration failed (look at last output)
2. Run migrations manually to see full error:
   ```bash
   kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db < migrations/XXX_failing_migration.sql
   ```
3. Fix the migration file and re-run

### "Read-only file system" Error

**Problem**: Cannot write files to PostgreSQL pod

**Solution**: The migration script uses stdin piping (not file writing) to avoid this issue. Ensure you're using `migrate_via_stdin.sh` and not older scripts.

### Local PostgreSQL Container Not Found

**Problem**: Script can't find `cold-storage-postgres`

**Solution**:
1. Check if container is running: `sg docker -c "docker ps | grep postgres"`
2. Start it if needed: `docker start cold-storage-postgres`
3. Or create new one:
   ```bash
   docker run --name cold-storage-postgres \
     -e POSTGRES_PASSWORD=postgres \
     -p 5432:5432 \
     -d postgres:16-alpine
   ```

### Duplicate Key Errors During Import

**Problem**: Data already exists in target database

**Solution**:
1. Drop and recreate database:
   ```bash
   kubectl exec -i cold-postgres-1 -- psql -U postgres -c "DROP DATABASE cold_db;"
   kubectl exec -i cold-postgres-1 -- psql -U postgres -c "CREATE DATABASE cold_db;"
   ```
2. Re-run migration

---

## Best Practices

1. **Always backup before migrating**:
   ```bash
   kubectl exec -i cold-postgres-1 -- pg_dump -U postgres cold_db > backup_$(date +%Y%m%d).sql
   ```

2. **Test migrations on dev environment first** before running on production

3. **Use transactions** in migration files to allow rollback on error:
   ```sql
   BEGIN;
   -- your changes here
   COMMIT;
   ```

4. **Keep migrations idempotent** - use `IF NOT EXISTS` when possible:
   ```sql
   CREATE TABLE IF NOT EXISTS users (...);
   ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);
   ```

5. **Never modify existing migrations** - create new ones to fix issues

---

## Architecture

### Local Development

```
┌─────────────────────────────┐
│  Local Development Machine  │
│                             │
│  ┌───────────────────────┐ │
│  │  Docker Container     │ │
│  │  cold-storage-postgres│ │
│  │  Port: 5432           │ │
│  │  Database: cold_db    │ │
│  └───────────────────────┘ │
└─────────────────────────────┘
```

### K3s Cluster

```
┌────────────────────────────────────────────────┐
│  K3s Cluster                                   │
│                                                │
│  ┌──────────────────────────────────────────┐ │
│  │  CloudNativePG Cluster                   │ │
│  │                                          │ │
│  │  ┌────────────┐  ┌────────────┐        │ │
│  │  │ Primary    │  │ Replica 1  │        │ │
│  │  │ postgres-1 │──│ postgres-2 │        │ │
│  │  └────────────┘  └────────────┘        │ │
│  │        │                                │ │
│  │        ├─── Replica 2 (postgres-3)     │ │
│  │        ├─── Replica 4 (postgres-4)     │ │
│  │        └─── Replica 5 (postgres-5)     │ │
│  │                                          │ │
│  │  Service: cold-postgres-rw (→ primary)  │ │
│  │  Service: cold-postgres-ro (→ replicas) │ │
│  └──────────────────────────────────────────┘ │
└────────────────────────────────────────────────┘
```

### Migration Flow

```
Local PostgreSQL  ─[export]→  SQL Dump  ─[stdin]→  K3s PostgreSQL
    (source)                    (file)              (destination)
```

---

## File Structure

```
cold-backend/
├── migrations/              # SQL migration files
│   ├── 001_create_users.sql
│   ├── 002_add_auth_to_users.sql
│   └── ...
│
├── scripts/
│   └── database/           # Database management tools
│       ├── README.md       # This file
│       └── migrate_via_stdin.sh  # Main migration tool
│
└── k8s/                    # Kubernetes manifests
    ├── cold-postgres-cluster.yaml
    └── ...
```

---

## Support

For issues or questions:
1. Check this README first
2. Review migration file comments
3. Check K3s pod logs: `kubectl logs cold-postgres-1`
4. Check application logs: `kubectl logs -l app=cold-backend`

---

## Version History

- **v1.0** - Initial migration tools
- **v1.1** - Added stdin-based migration (fixes read-only filesystem)
- **v1.2** - Added comprehensive documentation
