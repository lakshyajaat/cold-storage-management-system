# Role-Based Access Control - Comprehensive Test Plan

## Test Environment
- Server: http://192.168.15.195:8080
- Date: 2025-12-13

---

## Prerequisites

### 1. Create Test Users (Run these SQL commands)

```sql
-- Connect to database
sudo docker exec -it cold-storage-postgres psql -U postgres -d cold_db

-- Create Employee Test User
INSERT INTO users (email, password_hash, role, created_at, updated_at)
VALUES (
    'employee@test.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye7U4hWJQbFlLwt7xW.hQOKvH8QhPVN8S',  -- password: test123
    'employee',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Create Accountant Test User
INSERT INTO users (email, password_hash, role, created_at, updated_at)
VALUES (
    'accountant@test.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye7U4hWJQbFlLwt7xW.hQOKvH8QhPVN8S',  -- password: test123
    'accountant',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Verify users created
SELECT id, email, role FROM users WHERE email LIKE '%test.com';

-- Exit
\q
```

### 2. Get Authentication Tokens

```bash
# Login as Employee
curl -X POST http://192.168.15.195:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"employee@test.com","password":"test123"}' | jq

# Save the token as EMPLOYEE_TOKEN

# Login as Accountant
curl -X POST http://192.168.15.195:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"accountant@test.com","password":"test123"}' | jq

# Save the token as ACCOUNTANT_TOKEN

# Login as Admin (existing)
curl -X POST http://192.168.15.195:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"lakshyajaat2002@gmail.com","password":"your_password"}' | jq

# Save the token as ADMIN_TOKEN
```

---

## Test Cases

### Test 1: Employee Role - Entry Creation (Should PASS ✅)

```bash
# Test: Employee creates an entry
curl -X POST http://192.168.15.195:8080/api/entries \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9999999991",
    "name": "Test Employee Entry",
    "village": "Test Village",
    "expected_quantity": 100,
    "truck_category": "seed"
  }'

# Expected: SUCCESS (200 OK)
# Result: ______________________________
```

### Test 2: Employee Role - Payment Creation (Should FAIL ❌)

```bash
# Test: Employee tries to create payment
curl -X POST http://192.168.15.195:8080/api/rent-payments \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "entry_id": 1,
    "customer_name": "Test",
    "customer_phone": "9999999999",
    "total_rent": 1000,
    "amount_paid": 500,
    "balance": -500,
    "notes": "Test payment"
  }'

# Expected: FORBIDDEN (403)
# Result: ______________________________
```

### Test 3: Employee Role - View Entries (Should PASS ✅)

```bash
# Test: Employee views entries
curl -X GET http://192.168.15.195:8080/api/entries \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"

# Expected: SUCCESS (200 OK) - Returns list of entries
# Result: ______________________________
```

### Test 4: Employee Role - View Payments (Should FAIL ❌)

```bash
# Test: Employee tries to view payments
curl -X GET http://192.168.15.195:8080/api/rent-payments \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"

# Expected: FORBIDDEN (403)
# Result: ______________________________
```

### Test 5: Accountant Role - Payment Creation (Should PASS ✅)

```bash
# Test: Accountant creates payment
curl -X POST http://192.168.15.195:8080/api/rent-payments \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "entry_id": 1,
    "customer_name": "Test Customer",
    "customer_phone": "9999999999",
    "total_rent": 1000,
    "amount_paid": 500,
    "balance": -500,
    "notes": "Test payment by accountant"
  }'

# Expected: SUCCESS (200 OK)
# Result: ______________________________
```

### Test 6: Accountant Role - Entry Creation (Should FAIL ❌)

```bash
# Test: Accountant tries to create entry
curl -X POST http://192.168.15.195:8080/api/entries \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9999999992",
    "name": "Test Accountant Entry",
    "village": "Test Village",
    "expected_quantity": 100,
    "truck_category": "seed"
  }'

# Expected: FORBIDDEN (403)
# Result: ______________________________
```

### Test 7: Accountant Role - View Entries (Should PASS ✅)

```bash
# Test: Accountant views entries (read-only)
curl -X GET http://192.168.15.195:8080/api/entries \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN"

# Expected: SUCCESS (200 OK)
# Result: ______________________________
```

### Test 8: Accountant Role - View Payments (Should PASS ✅)

```bash
# Test: Accountant views payments
curl -X GET http://192.168.15.195:8080/api/rent-payments \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN"

# Expected: SUCCESS (200 OK)
# Result: ______________________________
```

### Test 9: Accountant Role - Update Room Entry (Should FAIL ❌)

```bash
# Test: Accountant tries to update room entry
curl -X PUT http://192.168.15.195:8080/api/room-entries/1 \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "room_no": "A101",
    "floor": "1",
    "gate_no": "G1",
    "quantity": 50,
    "remark": "Updated by accountant"
  }'

# Expected: FORBIDDEN (403)
# Result: ______________________________
```

### Test 10: Admin Role - All Operations (Should PASS ✅)

```bash
# Test 10a: Admin creates entry
curl -X POST http://192.168.15.195:8080/api/entries \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9999999993",
    "name": "Test Admin Entry",
    "village": "Test Village",
    "expected_quantity": 100,
    "truck_category": "sell"
  }'

# Expected: SUCCESS (200 OK)
# Result: ______________________________

# Test 10b: Admin creates payment
curl -X POST http://192.168.15.195:8080/api/rent-payments \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "entry_id": 2,
    "customer_name": "Test Admin",
    "customer_phone": "9999999993",
    "total_rent": 2000,
    "amount_paid": 1000,
    "balance": -1000,
    "notes": "Test payment by admin"
  }'

# Expected: SUCCESS (200 OK)
# Result: ______________________________

# Test 10c: Admin views payments
curl -X GET http://192.168.15.195:8080/api/rent-payments \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Expected: SUCCESS (200 OK)
# Result: ______________________________
```

---

## Dashboard Access Tests (Browser)

### Test 11: Employee Dashboard Access

1. **Login as Employee:**
   - URL: http://192.168.15.195:8080/login
   - Email: employee@test.com
   - Password: test123

2. **Expected Redirect:**
   - Should redirect to: `/dashboard` (Employee Dashboard)
   - Result: ______________________________

3. **Try to Access Accountant Dashboard:**
   - URL: http://192.168.15.195:8080/accountant/dashboard
   - Expected: Access Denied or redirect to `/dashboard`
   - Result: ______________________________

4. **Try to Access Admin Dashboard:**
   - URL: http://192.168.15.195:8080/admin/dashboard
   - Expected: Access Denied or redirect to `/dashboard`
   - Result: ______________________________

5. **Try to Access Entry Room (Should WORK):**
   - URL: http://192.168.15.195:8080/entry-room
   - Expected: SUCCESS - Can create entries
   - Result: ______________________________

6. **Try to Access Rent Management (Should FAIL):**
   - URL: http://192.168.15.195:8080/rent-management
   - Expected: Cannot process payments (API will block)
   - Result: ______________________________

### Test 12: Accountant Dashboard Access

1. **Login as Accountant:**
   - URL: http://192.168.15.195:8080/login
   - Email: accountant@test.com
   - Password: test123

2. **Expected Redirect:**
   - Should redirect to: `/accountant/dashboard`
   - Result: ______________________________

3. **Verify Dashboard Features:**
   - Can see payment statistics: ______________________________
   - Can see recent payments: ______________________________
   - Quick actions work: ______________________________

4. **Try to Access Entry Room (Should FAIL):**
   - URL: http://192.168.15.195:8080/entry-room
   - Expected: Can view page but API calls will fail (403)
   - Result: ______________________________

5. **Try to Access Rent Management (Should WORK):**
   - URL: http://192.168.15.195:8080/rent-management
   - Expected: SUCCESS - Can process payments
   - Result: ______________________________

6. **Try to Create Entry (Should FAIL):**
   - Navigate to entry-room page
   - Try to submit an entry
   - Expected: API returns 403 Forbidden
   - Result: ______________________________

7. **Try to Process Payment (Should WORK):**
   - Navigate to rent-management
   - Search for customer
   - Process payment
   - Expected: SUCCESS
   - Result: ______________________________

### Test 13: Admin Dashboard Access

1. **Login as Admin:**
   - URL: http://192.168.15.195:8080/login
   - Email: lakshyajaat2002@gmail.com
   - Password: (your password)

2. **Expected Redirect:**
   - Should redirect to: `/admin/dashboard`
   - Result: ______________________________

3. **Try All Features (Should ALL WORK):**
   - Access Entry Room: ______________________________
   - Create Entry: ______________________________
   - Access Rent Management: ______________________________
   - Process Payment: ______________________________
   - Access System Settings: ______________________________
   - Access Employee Management: ______________________________

---

## Permission Matrix Verification

| Action | Employee | Accountant | Admin |
|--------|----------|------------|-------|
| View Entries | ✅ Should PASS | ✅ Should PASS | ✅ Should PASS |
| Create Entry | ✅ Should PASS | ❌ Should FAIL | ✅ Should PASS |
| Update Room Entry | ✅ Should PASS | ❌ Should FAIL | ✅ Should PASS |
| View Payments | ❌ Should FAIL | ✅ Should PASS | ✅ Should PASS |
| Create Payment | ❌ Should FAIL | ✅ Should PASS | ✅ Should PASS |
| View Dashboard | ✅ Employee | ✅ Accountant | ✅ Admin |
| System Settings | ❌ Should FAIL | ❌ Should FAIL | ✅ Should PASS |

Mark each cell with actual results:
- ✅ = Passed as expected
- ❌ = Failed as expected
- ⚠️ = Unexpected result

---

## Test Summary Report

### Test Results Overview

| Test # | Description | Expected | Actual | Status |
|--------|-------------|----------|--------|--------|
| 1 | Employee creates entry | PASS | ______ | ______ |
| 2 | Employee creates payment | FAIL (403) | ______ | ______ |
| 3 | Employee views entries | PASS | ______ | ______ |
| 4 | Employee views payments | FAIL (403) | ______ | ______ |
| 5 | Accountant creates payment | PASS | ______ | ______ |
| 6 | Accountant creates entry | FAIL (403) | ______ | ______ |
| 7 | Accountant views entries | PASS | ______ | ______ |
| 8 | Accountant views payments | PASS | ______ | ______ |
| 9 | Accountant updates room | FAIL (403) | ______ | ______ |
| 10a | Admin creates entry | PASS | ______ | ______ |
| 10b | Admin creates payment | PASS | ______ | ______ |
| 10c | Admin views payments | PASS | ______ | ______ |
| 11 | Employee dashboard access | Correct | ______ | ______ |
| 12 | Accountant dashboard access | Correct | ______ | ______ |
| 13 | Admin dashboard access | Full Access | ______ | ______ |

### Issues Found

1. _______________________________________________________________
2. _______________________________________________________________
3. _______________________________________________________________

### Notes

_______________________________________________________________
_______________________________________________________________
_______________________________________________________________

---

## Quick Test Script

Save this as `test_roles.sh` and run it:

```bash
#!/bin/bash

echo "🧪 Role-Based Access Control Test Suite"
echo "========================================"
echo ""

# Set your tokens here
EMPLOYEE_TOKEN="paste_employee_token_here"
ACCOUNTANT_TOKEN="paste_accountant_token_here"
ADMIN_TOKEN="paste_admin_token_here"

BASE_URL="http://192.168.15.195:8080"

echo "Test 1: Employee creates entry (Should PASS)"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -X POST $BASE_URL/api/entries \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone":"9999999991","name":"Test","village":"Test","expected_quantity":100,"truck_category":"seed"}'

echo ""
echo "Test 2: Employee creates payment (Should FAIL - 403)"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -X POST $BASE_URL/api/rent-payments \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"entry_id":1,"customer_name":"Test","customer_phone":"999","total_rent":1000,"amount_paid":500,"balance":-500}'

echo ""
echo "Test 3: Accountant creates payment (Should PASS)"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -X POST $BASE_URL/api/rent-payments \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"entry_id":1,"customer_name":"Test","customer_phone":"999","total_rent":1000,"amount_paid":500,"balance":-500}'

echo ""
echo "Test 4: Accountant creates entry (Should FAIL - 403)"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -X POST $BASE_URL/api/entries \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone":"9999999992","name":"Test","village":"Test","expected_quantity":100,"truck_category":"seed"}'

echo ""
echo "Test 5: Accountant views payments (Should PASS)"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -X GET $BASE_URL/api/rent-payments \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN"

echo ""
echo "Test 6: Employee views payments (Should FAIL - 403)"
curl -s -o /dev/null -w "Status: %{http_code}\n" \
  -X GET $BASE_URL/api/rent-payments \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"

echo ""
echo "✅ Test suite complete!"
```

---

## End of Test Plan
