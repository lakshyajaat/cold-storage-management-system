# Database Migration - Quick Start Guide

## ✅ What's Been Completed

### 1. Automatic Migrations on App Startup

The application now **automatically creates all database tables** when it starts!

**How it works:**
- When the app starts, it runs all migration files in `/migrations/` directory
- Creates a `schema_migrations` table to track which migrations have been run
- Skips migrations that have already been applied
- Only runs new migrations

**No manual setup required** - just deploy the app and tables are created automatically!

### 2. Data Migration Tool

A tool to migrate your existing data from local PostgreSQL to the K3s cluster.

**Location:** `scripts/database/migrate_via_stdin.sh`

**Usage:**
```bash
cd /home/lakshya/jupyter-/cold/cold-backend
./scripts/database/migrate_via_stdin.sh
```

---

## 📋 Migration Status

✅ **Schema:** Automatically created on app startup
✅ **Data:** Successfully migrated to K3s cluster

**Current data in K3s cluster:**
- users: 1
- customers: 11
- entries: 40
- entry_events: 4

---

## 🚀 Deploying to K3s Cluster

### Option 1: Automatic (Recommended)

Just deploy the app - migrations run automatically:

```bash
# Build and deploy
sg docker -c "docker build -t cold-backend:v1.4 ."
sg docker -c "docker save cold-backend:v1.4 > cold-backend-v1.4.tar"

# Deploy to all K3s nodes
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    scp cold-backend-v1.4.tar root@$NODE_IP:/tmp/
    ssh root@$NODE_IP 'ctr -n k8s.io images import /tmp/cold-backend-v1.4.tar && rm /tmp/cold-backend-v1.4.tar'
done

# Update deployments
sed -i 's/v1.3.2/v1.4/g' k8s/03-deployment-employee.yaml k8s/04-deployment-customer.yaml
kubectl apply -f k8s/03-deployment-employee.yaml -f k8s/04-deployment-customer.yaml
```

**The app will:**
1. Connect to PostgreSQL
2. Run all pending migrations automatically
3. Create all required tables
4. Start serving requests

### Option 2: Manual Migration (for development)

If you need to migrate data from local PostgreSQL:

```bash
./scripts/database/migrate_via_stdin.sh
```

---

## 📂 File Structure

```
cold-backend/
│
├── migrations/                          # SQL migration files
│   ├── 001_create_users.sql
│   ├── 002_add_auth_to_users.sql
│   └── ...
│
├── internal/
│   └── database/
│       └── migrator.go                  # Automatic migration runner
│
├── cmd/server/main.go                   # App startup (runs migrations)
│
└── scripts/database/                    # Migration tools
    ├── README.md                         # Full documentation
    ├── QUICK_START.md                    # This file
    └── migrate_via_stdin.sh              # Data migration tool
```

---

## 🔍 Checking Migration Status

### Check which migrations have been applied

```bash
kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db -c "SELECT * FROM schema_migrations ORDER BY applied_at;"
```

### Check if tables exist

```bash
kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db -c "\dt"
```

### Check row counts

```bash
kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db -c "
SELECT
    tablename,
    n_live_tup as rows
FROM pg_stat_user_tables
ORDER BY tablename;
"
```

---

## 🛠 Creating New Migrations

1. Create a new SQL file in `migrations/` directory:
   ```bash
   touch migrations/020_add_new_feature.sql
   ```

2. Write your SQL (use idempotent statements):
   ```sql
   -- migrations/020_add_new_feature.sql
   CREATE TABLE IF NOT EXISTS new_table (
       id SERIAL PRIMARY KEY,
       name VARCHAR(255) NOT NULL
   );

   ALTER TABLE users
   ADD COLUMN IF NOT EXISTS new_column VARCHAR(100);
   ```

3. Deploy the app - migration runs automatically!

---

## 📝 Migration Best Practices

### ✅ DO:
- Use `CREATE TABLE IF NOT EXISTS`
- Use `ADD COLUMN IF NOT EXISTS` for ALTER TABLE
- Keep migrations small and focused
- Name files with sequential numbers (001_, 002_, etc.)
- Test migrations on development environment first

### ❌ DON'T:
- Modify existing migration files (create new ones instead)
- Delete migration files (they're tracked in schema_migrations table)
- Use `DROP TABLE` without backups
- Run `000_reset.sql` files in production (they're automatically skipped)

---

## 🎯 Summary

**Before (Manual):**
1. Connect to database manually
2. Run each migration file one by one
3. Track which migrations you've run yourself
4. Risk of missing migrations or running them twice

**Now (Automatic):**
1. Deploy app
2. Tables created automatically ✨
3. No manual intervention needed
4. Safe to redeploy - already-run migrations are skipped

---

## 📞 Troubleshooting

### App fails to start with "migration error"

**Check:**
1. Is PostgreSQL running? `kubectl get pods | grep postgres`
2. Can app connect to database? Check app logs: `kubectl logs -l app=cold-backend`
3. Are migration files present? `ls -l migrations/`

**Fix:**
- Check database connection string in config
- Ensure migrations folder is included in Docker image (check Dockerfile)

### Migration fails partway through

**The problem:**
- Migration tracking ensures already-run migrations are skipped
- Failed migration will be retried on next app start

**To fix:**
1. Fix the failing migration SQL file
2. Restart the app (it will retry the failed migration)

### Need to rollback a migration

**Option 1:** Create a new "down" migration:
```sql
-- migrations/021_rollback_feature.sql
DROP TABLE IF EXISTS new_table;
ALTER TABLE users DROP COLUMN IF EXISTS new_column;
```

**Option 2:** Manually remove from tracking table:
```bash
kubectl exec -i cold-postgres-1 -- psql -U postgres -d cold_db -c \
    "DELETE FROM schema_migrations WHERE filename = '020_add_new_feature.sql';"
```

Then modify the migration file and restart the app.

---

## 🎉 Success!

Your database migrations are now fully automated and documented!

- ✅ Tables created automatically on deployment
- ✅ Data migrated to K3s cluster
- ✅ Comprehensive documentation added
- ✅ Migration tools in `scripts/database/`
