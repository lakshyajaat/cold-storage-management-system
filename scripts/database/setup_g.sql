-- Private inventory database setup
-- Run this script manually on PostgreSQL server

-- Create database and user
-- CREATE DATABASE g_db;
-- CREATE USER g_user WITH PASSWORD 'your_secure_password_here';
-- GRANT ALL PRIVILEGES ON DATABASE g_db TO g_user;

-- Connect to g_db before running the rest
-- \c g_db

-- Items table
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(50),
    floor INT NOT NULL CHECK (floor >= 0 AND floor <= 4),
    current_qty INT DEFAULT 0,
    unit_cost DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Transactions table (audit log)
CREATE TABLE IF NOT EXISTS txns (
    id SERIAL PRIMARY KEY,
    item_id INT REFERENCES items(id),
    type VARCHAR(10) NOT NULL CHECK (type IN ('in', 'out')),
    qty INT NOT NULL,
    unit_price DECIMAL(10,2),
    total DECIMAL(10,2),
    reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Configuration table
CREATE TABLE IF NOT EXISTS cfg (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT NOT NULL
);

-- Access logs
CREATE TABLE IF NOT EXISTS access_log (
    id SERIAL PRIMARY KEY,
    device_hash VARCHAR(64),
    ip_address VARCHAR(45),
    success BOOLEAN,
    fail_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Active sessions
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    token VARCHAR(128) UNIQUE,
    device_hash VARCHAR(64),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_items_floor ON items(floor);
CREATE INDEX IF NOT EXISTS idx_txns_item_id ON txns(item_id);
CREATE INDEX IF NOT EXISTS idx_txns_created_at ON txns(created_at);
CREATE INDEX IF NOT EXISTS idx_access_log_created_at ON access_log(created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Grant permissions to user
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO g_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO g_user;
