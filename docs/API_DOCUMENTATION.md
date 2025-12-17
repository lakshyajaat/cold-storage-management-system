# API Documentation

Complete API reference for the Cold Storage Management System.

**Base URL:** `http://localhost:8080`

**Version:** 1.0.0

---

## Table of Contents

- [Authentication](#authentication)
- [Users API](#users-api)
- [Customers API](#customers-api)
- [Entries API](#entries-api)
- [Room Entries API](#room-entries-api)
- [Payments API](#payments-api)
- [System Settings API](#system-settings-api)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)

---

## Authentication

All protected endpoints require JWT authentication.

### Headers

```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Login

**Endpoint:** `POST /auth/login`

**Description:** Authenticate user and receive JWT token

**Request Body:**
```json
{
  "email": "string (required)",
  "password": "string (required)"
}
```

**Success Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "role": "employee",
    "phone": "",
    "village": "",
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  }
}
```

**Error Responses:**

`401 Unauthorized` - Invalid credentials
```json
{
  "error": "Invalid email or password"
}
```

`400 Bad Request` - Missing fields
```json
{
  "error": "Email and password are required"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@cold.com",
    "password": "admin123"
  }'
```

---

### Signup

**Endpoint:** `POST /auth/signup`

**Description:** Create a new user account

**Request Body:**
```json
{
  "email": "string (required)",
  "password": "string (required, min 6 chars)",
  "name": "string (required)",
  "role": "string (optional, default: employee)"
}
```

**Success Response:** `200 OK`
```json
{
  "message": "User created successfully",
  "user_id": 5
}
```

**Error Responses:**

`409 Conflict` - Email already exists
```json
{
  "error": "User with this email already exists"
}
```

`400 Bad Request` - Validation error
```json
{
  "error": "Password must be at least 6 characters"
}
```

---

## Users API

**Base Path:** `/api/users`

**Required Role:** Admin

### List All Users

**Endpoint:** `GET /api/users`

**Authorization:** Admin only

**Success Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Admin User",
    "email": "admin@cold.com",
    "phone": "",
    "village": "",
    "role": "admin",
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  },
  {
    "id": 2,
    "name": "Employee User",
    "email": "employee@cold.com",
    "phone": "9999999999",
    "village": "Village Name",
    "role": "employee",
    "created_at": "2025-12-13T10:30:00Z",
    "updated_at": "2025-12-13T10:30:00Z"
  }
]
```

**Example:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/users
```

---

### Get User by ID

**Endpoint:** `GET /api/users/{id}`

**Authorization:** Admin only

**URL Parameters:**
- `id` (integer) - User ID

**Success Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Admin User",
  "email": "admin@cold.com",
  "phone": "",
  "village": "",
  "role": "admin",
  "created_at": "2025-12-13T10:00:00Z",
  "updated_at": "2025-12-13T10:00:00Z"
}
```

**Error Response:** `404 Not Found`

---

### Create User

**Endpoint:** `POST /api/users`

**Authorization:** Admin only

**Request Body:**
```json
{
  "email": "string (required)",
  "password": "string (required)",
  "name": "string (required)",
  "role": "string (required: employee|accountant|admin)",
  "phone": "string (optional)",
  "village": "string (optional)"
}
```

**Success Response:** `200 OK`
```json
{
  "id": 3,
  "name": "New Employee",
  "email": "newemployee@cold.com",
  "phone": "9876543210",
  "village": "Test Village",
  "role": "employee",
  "created_at": "2025-12-13T11:00:00Z",
  "updated_at": "2025-12-13T11:00:00Z"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "accountant@example.com",
    "password": "secure123",
    "name": "New Accountant",
    "role": "accountant"
  }'
```

---

### Update User

**Endpoint:** `PUT /api/users/{id}`

**Authorization:** Admin only

**URL Parameters:**
- `id` (integer) - User ID

**Request Body:**
```json
{
  "name": "string (optional)",
  "role": "string (optional)",
  "phone": "string (optional)",
  "village": "string (optional)"
}
```

**Success Response:** `200 OK`
```json
{
  "id": 3,
  "name": "Updated Name",
  "email": "employee@cold.com",
  "phone": "9999999999",
  "village": "New Village",
  "role": "accountant",
  "created_at": "2025-12-13T10:00:00Z",
  "updated_at": "2025-12-13T12:00:00Z"
}
```

---

### Delete User

**Endpoint:** `DELETE /api/users/{id}`

**Authorization:** Admin only

**URL Parameters:**
- `id` (integer) - User ID

**Success Response:** `204 No Content`

**Error Response:** `404 Not Found`

---

## Customers API

**Base Path:** `/api/customers`

**Required Role:** All authenticated users

### List All Customers

**Endpoint:** `GET /api/customers`

**Success Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Customer Name",
    "phone": "9999999999",
    "village": "Village Name",
    "address": "Complete address",
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  }
]
```

---

### Search Customer by Phone

**Endpoint:** `GET /api/customers/search`

**Query Parameters:**
- `phone` (string, required) - Customer phone number

**Success Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Customer Name",
  "phone": "9999999999",
  "village": "Village Name",
  "address": "Complete address",
  "created_at": "2025-12-13T10:00:00Z",
  "updated_at": "2025-12-13T10:00:00Z"
}
```

**Error Response:** `404 Not Found`
```json
{
  "error": "Customer not found"
}
```

**Example:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/customers/search?phone=9999999999"
```

---

### Create Customer

**Endpoint:** `POST /api/customers`

**Request Body:**
```json
{
  "name": "string (required)",
  "phone": "string (required, unique)",
  "village": "string (required)",
  "address": "string (optional)"
}
```

**Success Response:** `200 OK`
```json
{
  "id": 2,
  "name": "New Customer",
  "phone": "9876543210",
  "village": "Test Village",
  "address": "123 Main Street",
  "created_at": "2025-12-13T11:00:00Z",
  "updated_at": "2025-12-13T11:00:00Z"
}
```

**Error Response:** `409 Conflict` - Phone already exists

---

### Update Customer

**Endpoint:** `PUT /api/customers/{id}`

**URL Parameters:**
- `id` (integer) - Customer ID

**Request Body:**
```json
{
  "name": "string (optional)",
  "village": "string (optional)",
  "address": "string (optional)"
}
```

**Note:** Phone number cannot be updated (unique identifier)

---

### Delete Customer

**Endpoint:** `DELETE /api/customers/{id}`

**URL Parameters:**
- `id` (integer) - Customer ID

**Success Response:** `204 No Content`

---

## Entries API

**Base Path:** `/api/entries`

### List All Entries

**Endpoint:** `GET /api/entries`

**Authorization:** All authenticated users

**Success Response:** `200 OK`
```json
[
  {
    "id": 1,
    "customer_id": 5,
    "phone": "9999999999",
    "name": "Customer Name",
    "village": "Village Name",
    "expected_quantity": 500,
    "truck_category": "seed",
    "truck_number": "SEED-001",
    "created_by_user_id": 2,
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  }
]
```

---

### Create Entry

**Endpoint:** `POST /api/entries`

**Authorization:** Employee or Admin only

**Request Body:**
```json
{
  "phone": "string (required)",
  "name": "string (required)",
  "village": "string (required)",
  "expected_quantity": "integer (required, 1-1500)",
  "truck_category": "string (required: seed|sell)"
}
```

**Business Rules:**
- Quantity 1-600: Automatically assigned to "seed" category
- Quantity 601-1500: Automatically assigned to "sell" category
- Truck number auto-generated: SEED-001, SEED-002, SELL-001, etc.

**Success Response:** `200 OK`
```json
{
  "id": 38,
  "customer_id": 14,
  "phone": "9999999991",
  "name": "Test Entry",
  "village": "Test Village",
  "expected_quantity": 100,
  "truck_category": "seed",
  "truck_number": "021/100",
  "created_by_user_id": 5,
  "created_at": "2025-12-13T10:30:00Z",
  "updated_at": "2025-12-13T10:30:00Z"
}
```

**Error Responses:**

`403 Forbidden` - Accountant trying to create entry
```json
{
  "error": "Forbidden: Insufficient permissions"
}
```

`400 Bad Request` - Validation error
```json
{
  "error": "Expected quantity must be between 1 and 1500"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/entries \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9999999999",
    "name": "Customer Name",
    "village": "Village Name",
    "expected_quantity": 450,
    "truck_category": "seed"
  }'
```

---

### Get Entry by ID

**Endpoint:** `GET /api/entries/{id}`

**URL Parameters:**
- `id` (integer) - Entry ID

**Success Response:** `200 OK`

---

### Get Entries by Customer

**Endpoint:** `GET /api/entries/customer/{customer_id}`

**URL Parameters:**
- `customer_id` (integer) - Customer ID

**Success Response:** `200 OK` - Array of entries

---

### Get Entry Count by Category

**Endpoint:** `GET /api/entries/count`

**Success Response:** `200 OK`
```json
{
  "seed": 150,
  "sell": 75
}
```

---

### Get Unassigned Entries

**Endpoint:** `GET /api/entries/unassigned`

**Description:** Get entries without room assignments

**Success Response:** `200 OK` - Array of entries without room_entries

---

## Room Entries API

**Base Path:** `/api/room-entries`

### List Room Entries

**Endpoint:** `GET /api/room-entries`

**Authorization:** All authenticated users

**Success Response:** `200 OK`
```json
[
  {
    "id": 1,
    "entry_id": 1,
    "room_no": "A-101",
    "floor": "1",
    "gate_no": "G1",
    "quantity": 500,
    "remark": "First floor storage",
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  }
]
```

---

### Create Room Entry

**Endpoint:** `POST /api/room-entries`

**Authorization:** Employee or Admin only

**Request Body:**
```json
{
  "entry_id": "integer (required)",
  "room_no": "string (required)",
  "floor": "string (required)",
  "gate_no": "string (required)",
  "quantity": "integer (required, > 0)",
  "remark": "string (optional)"
}
```

**Success Response:** `200 OK`
```json
{
  "id": 15,
  "entry_id": 1,
  "room_no": "A-102",
  "floor": "2",
  "gate_no": "G2",
  "quantity": 250,
  "remark": "Second floor allocation",
  "created_at": "2025-12-13T11:00:00Z",
  "updated_at": "2025-12-13T11:00:00Z"
}
```

**Error Response:** `403 Forbidden` - Accountant trying to create

---

### Get Room Entry by ID

**Endpoint:** `GET /api/room-entries/{id}`

**URL Parameters:**
- `id` (integer) - Room Entry ID

**Success Response:** `200 OK`

---

### Update Room Entry

**Endpoint:** `PUT /api/room-entries/{id}`

**Authorization:** Employee or Admin only

**URL Parameters:**
- `id` (integer) - Room Entry ID

**Request Body:**
```json
{
  "room_no": "string (required)",
  "floor": "string (required)",
  "gate_no": "string (required)",
  "quantity": "integer (required, > 0)",
  "remark": "string (optional)"
}
```

**Success Response:** `200 OK`
```json
{
  "id": 15,
  "entry_id": 1,
  "room_no": "A-103",
  "floor": "1",
  "gate_no": "G1",
  "quantity": 300,
  "remark": "Updated location",
  "created_at": "2025-12-13T11:00:00Z",
  "updated_at": "2025-12-13T12:00:00Z"
}
```

**Error Responses:**

`403 Forbidden` - Accountant trying to update
```json
{
  "error": "Forbidden: Insufficient permissions"
}
```

`404 Not Found` - Room entry not found

---

## Payments API

**Base Path:** `/api/rent-payments`

**Required Role:** Accountant or Admin

### Create Payment

**Endpoint:** `POST /api/rent-payments`

**Authorization:** Accountant or Admin only

**Request Body:**
```json
{
  "entry_id": "integer (required)",
  "customer_name": "string (required)",
  "customer_phone": "string (required)",
  "total_rent": "number (required, > 0)",
  "amount_paid": "number (required, > 0)",
  "balance": "number (required)",
  "notes": "string (optional)"
}
```

**Business Rules:**
- Balance is cumulative: (Previous Total Paid + Current Amount) - Total Rent
- Negative balance indicates amount owed
- Positive balance indicates overpayment/credit

**Success Response:** `200 OK`
```json
{
  "id": 14,
  "entry_id": 1,
  "customer_name": "Customer Name",
  "customer_phone": "9999999999",
  "total_rent": 10000.00,
  "amount_paid": 5000.00,
  "balance": -5000.00,
  "payment_date": "2025-12-13T11:00:00Z",
  "processed_by_user_id": 6,
  "notes": "Partial payment",
  "created_at": "2025-12-13T11:00:00Z"
}
```

**Error Response:** `403 Forbidden` - Employee trying to create payment

**Example:**
```bash
curl -X POST http://localhost:8080/api/rent-payments \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "entry_id": 1,
    "customer_name": "Customer Name",
    "customer_phone": "9999999999",
    "total_rent": 10000,
    "amount_paid": 5000,
    "balance": -5000,
    "notes": "First payment"
  }'
```

---

### List All Payments

**Endpoint:** `GET /api/rent-payments`

**Authorization:** Accountant or Admin only

**Success Response:** `200 OK` - Array of payments

**Error Response:** `403 Forbidden` - Employee trying to view

---

### Get Payments by Entry

**Endpoint:** `GET /api/rent-payments/entry/{entry_id}`

**Authorization:** Accountant or Admin only

**URL Parameters:**
- `entry_id` (integer) - Entry ID

**Success Response:** `200 OK` - Array of payments for the entry

---

### Get Payments by Phone

**Endpoint:** `GET /api/rent-payments/phone`

**Authorization:** Accountant or Admin only

**Query Parameters:**
- `phone` (string, required) - Customer phone number

**Success Response:** `200 OK` - Array of all payments for the customer

**Example:**
```bash
curl -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  "http://localhost:8080/api/rent-payments/phone?phone=9999999999"
```

---

## System Settings API

**Base Path:** `/api/settings`

**Required Role:** Admin (read), Admin (write)

### List All Settings

**Endpoint:** `GET /api/settings`

**Authorization:** All authenticated users

**Success Response:** `200 OK`
```json
[
  {
    "id": 1,
    "setting_key": "rent_per_item",
    "setting_value": "10.50",
    "description": "Rent price per item",
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  },
  {
    "id": 2,
    "setting_key": "company_name",
    "setting_value": "Cold Storage Solutions",
    "description": "Company name for receipts",
    "created_at": "2025-12-13T10:00:00Z",
    "updated_at": "2025-12-13T10:00:00Z"
  }
]
```

---

### Get Setting by Key

**Endpoint:** `GET /api/settings/{key}`

**Authorization:** All authenticated users

**URL Parameters:**
- `key` (string) - Setting key (e.g., "rent_per_item")

**Success Response:** `200 OK`
```json
{
  "id": 1,
  "setting_key": "rent_per_item",
  "setting_value": "10.50",
  "description": "Rent price per item",
  "created_at": "2025-12-13T10:00:00Z",
  "updated_at": "2025-12-13T10:00:00Z"
}
```

**Example:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/settings/rent_per_item
```

---

### Update Setting

**Endpoint:** `PUT /api/settings/{key}`

**Authorization:** Admin only

**URL Parameters:**
- `key` (string) - Setting key

**Request Body:**
```json
{
  "setting_value": "string (required)"
}
```

**Success Response:** `200 OK`
```json
{
  "id": 1,
  "setting_key": "rent_per_item",
  "setting_value": "12.00",
  "description": "Rent price per item",
  "created_at": "2025-12-13T10:00:00Z",
  "updated_at": "2025-12-13T13:00:00Z"
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/settings/rent_per_item \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "setting_value": "15.50"
  }'
```

---

## Error Handling

### Standard Error Response

```json
{
  "error": "Error message description"
}
```

### HTTP Status Codes

| Code | Description | When Used |
|------|-------------|-----------|
| 200 | OK | Successful GET, POST, PUT |
| 201 | Created | Resource created successfully |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Validation error, malformed request |
| 401 | Unauthorized | Missing or invalid token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate resource (email, phone) |
| 500 | Internal Server Error | Server-side error |

### Common Error Messages

**Authentication Errors:**
```json
{
  "error": "Authorization header required"
}
```
```json
{
  "error": "Invalid or expired token"
}
```

**Permission Errors:**
```json
{
  "error": "Forbidden: Insufficient permissions"
}
```

**Validation Errors:**
```json
{
  "error": "Invalid request body"
}
```
```json
{
  "error": "validation failed"
}
```

**Not Found Errors:**
```json
{
  "error": "User not found"
}
```
```json
{
  "error": "Customer not found"
}
```

---

## Rate Limiting

Currently, there is no rate limiting implemented. Consider adding rate limiting for production deployments.

**Recommended:** 100 requests per minute per IP address

---

## Pagination

Currently, list endpoints return all results. For large datasets, consider implementing pagination:

**Future Implementation:**
```
GET /api/entries?page=1&limit=50
```

---

## Versioning

Current API version: **v1**

API versioning is not currently implemented in the URL structure. All endpoints are considered v1.

**Future Implementation:**
```
/api/v1/entries
/api/v2/entries
```

---

## Testing Examples

### Complete Workflow Example

```bash
# 1. Login as employee
EMPLOYEE_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"employee@test.com","password":"test123"}' \
  | jq -r '.token')

# 2. Create a customer
curl -X POST http://localhost:8080/api/customers \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Customer",
    "phone": "9876543210",
    "village": "Test Village"
  }'

# 3. Create an entry
curl -X POST http://localhost:8080/api/entries \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9876543210",
    "name": "Test Customer",
    "village": "Test Village",
    "expected_quantity": 300,
    "truck_category": "seed"
  }'

# 4. Login as accountant
ACCOUNTANT_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"accountant@test.com","password":"test123"}' \
  | jq -r '.token')

# 5. Process payment
curl -X POST http://localhost:8080/api/rent-payments \
  -H "Authorization: Bearer $ACCOUNTANT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "entry_id": 1,
    "customer_name": "Test Customer",
    "customer_phone": "9876543210",
    "total_rent": 5000,
    "amount_paid": 2500,
    "balance": -2500,
    "notes": "First installment"
  }'
```

---

## Postman Collection

A Postman collection with all endpoints and example requests is available in:
`/docs/Cold_Storage_API.postman_collection.json`

Import this into Postman for easy API testing.

---

## WebSocket Support

WebSocket support is not currently implemented. Consider adding for real-time updates:

**Future Implementation:**
- Real-time entry notifications
- Live payment updates
- Dashboard statistics streaming

---

**Last Updated:** 2025-12-13

**Maintainer:** Cold Storage Development Team
