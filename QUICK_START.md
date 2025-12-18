# Quick Start Guide

Get the Cold Storage Management System running in 5 minutes!

---

## Prerequisites

- Docker installed
- Go 1.23+ installed
- Git installed

---

## 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd cold-backend

# Copy environment file
cp .env.example .env
```

---

## 2. Start PostgreSQL Database

```bash
docker run --name cold-storage-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=cold_db \
  -p 5432:5432 \
  -d postgres:17
```

**Verify database is running:**
```bash
docker ps | grep cold-storage-postgres
```

---

## 3. Run Database Migrations

```bash
# Navigate to project directory
cd /path/to/cold-backend

# Run all migrations
for file in migrations/*.sql; do
  echo "Running $file..."
  docker exec -i cold-storage-postgres \
    psql -U postgres -d cold_db < "$file"
done
```

**Or run individually:**
```bash
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/001_create_users.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/002_add_auth_to_users.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/003_seed_admin_user.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/004_add_role_to_users.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/005_create_customers.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/006_create_entries.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/007_create_entry_events.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/008_create_room_entries.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/009_create_system_settings.sql
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/010_create_rent_payments.sql
```

---

## 4. Build and Run Application

```bash
# Install dependencies
go mod download

# Build the application
go build -o cold-backend cmd/server/main.go

# Run the application
./cold-backend
```

**Or run directly without building:**
```bash
go run cmd/server/main.go
```

---

## 5. Access the Application

Open your browser and navigate to:

```
http://localhost:8080
```

### Default Login Credentials

**Admin User:**
- Email: `admin@cold.com`
- Password: `admin123`

⚠️ **Change the admin password immediately after first login!**

---

## 6. Create Test Users (Optional)

```bash
# Get admin token first
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cold.com","password":"admin123"}' \
  | jq -r '.token')

# Create employee user
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "employee@test.com",
    "password": "test123",
    "name": "Test Employee",
    "role": "employee"
  }'

# Create accountant user
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "accountant@test.com",
    "password": "test123",
    "name": "Test Accountant",
    "role": "accountant"
  }'
```

---

## 7. Test the Application

### Test Employee Workflow

```bash
# Login as employee
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "employee@test.com",
    "password": "test123"
  }'

# Navigate to: http://localhost:8080/entry-room
# Create a new entry
```

### Test Accountant Workflow

```bash
# Login as accountant
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "accountant@test.com",
    "password": "test123"
  }'

# Navigate to: http://localhost:8080/accountant/dashboard
# Process a payment
```

---

## Quick Commands Reference

### Database Operations

```bash
# Connect to database
docker exec -it cold-storage-postgres psql -U postgres -d cold_db

# View all users
docker exec -it cold-storage-postgres psql -U postgres -d cold_db -c "SELECT id, email, role FROM users;"

# View all entries
docker exec -it cold-storage-postgres psql -U postgres -d cold_db -c "SELECT id, truck_number, name, expected_quantity FROM entries LIMIT 10;"

# Reset database (CAUTION: Deletes all data!)
docker exec -i cold-storage-postgres psql -U postgres -d cold_db < migrations/000_complete_reset.sql
```

### Server Operations

```bash
# Check if server is running
ps aux | grep cold-backend

# Stop server
pkill -f cold-backend

# Restart server
./cold-backend > /tmp/cold-backend.log 2>&1 &

# View logs
tail -f /tmp/cold-backend.log

# Check server port
lsof -i :8080
```

### Docker Operations

```bash
# Stop database
docker stop cold-storage-postgres

# Start database
docker start cold-storage-postgres

# Remove database (CAUTION: Deletes all data!)
docker rm -f cold-storage-postgres

# View database logs
docker logs cold-storage-postgres
```

---

## Troubleshooting

### Problem: Port 8080 already in use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Problem: Database connection refused

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check database logs
docker logs cold-storage-postgres

# Restart database
docker restart cold-storage-postgres
```

### Problem: Cannot run migrations

```bash
# Make sure you're in the project directory
pwd

# Check if migration files exist
ls -la migrations/

# Try running migrations one by one
docker exec -i cold-storage-postgres \
  psql -U postgres -d cold_db < migrations/001_create_users.sql
```

### Problem: Forgot admin password

```bash
# Connect to database
docker exec -it cold-storage-postgres psql -U postgres -d cold_db

# Reset admin password to 'admin123'
UPDATE users SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMye7U4hWJQbFlLwt7xW.hQOKvH8QhPVN8S' WHERE email = 'admin@cold.com';
```

### Problem: JWT token errors

```bash
# Check .env file has JWT_SECRET
cat .env | grep JWT_SECRET

# If missing, add it
echo "JWT_SECRET=your-super-secret-jwt-key" >> .env

# Restart the server
pkill -f cold-backend
./cold-backend
```

---

## Next Steps

1. **Read the full documentation:** See `README.md` for comprehensive documentation
2. **Explore the API:** Check `API_DOCUMENTATION.md` for API reference
3. **Run tests:** Review `TEST_RESULTS.md` for testing procedures
4. **Configure settings:** Access `/system-settings` as admin to configure rent rates
5. **Create real users:** Replace test users with actual employees and accountants

---

## Production Checklist

Before deploying to production:

- [ ] Change admin password
- [ ] Update JWT_SECRET in .env with a strong secret key
- [ ] Update database password
- [ ] Configure firewall rules
- [ ] Set up SSL/TLS (HTTPS)
- [ ] Configure backup strategy
- [ ] Set up monitoring and logging
- [ ] Review and test role-based permissions
- [ ] Test all critical workflows
- [ ] Document any custom configurations

---

## Support

For detailed documentation:
- **README.md** - Complete project documentation
- **API_DOCUMENTATION.md** - API reference
- **TEST_RESULTS.md** - Test results and examples
- **ROLE_BASED_ACCESS_TEST.md** - Role-based access testing guide

---

**Happy coding! 🚀**
