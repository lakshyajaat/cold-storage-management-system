-- Migration: Add entry management features (reassign entries, merge customers)

-- Add can_manage_entries column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS can_manage_entries BOOLEAN DEFAULT FALSE;

-- Create entry_management_logs table for tracking reassignments and merges
CREATE TABLE IF NOT EXISTS entry_management_logs (
    id SERIAL PRIMARY KEY,
    action_type VARCHAR(20) NOT NULL, -- 'reassign' or 'merge'
    performed_by_id INTEGER NOT NULL REFERENCES users(id),

    -- For reassignment: entry details
    entry_id INTEGER,
    thock_number VARCHAR(20),
    old_customer_id INTEGER,
    old_customer_name VARCHAR(255),
    old_customer_phone VARCHAR(20),
    new_customer_id INTEGER,
    new_customer_name VARCHAR(255),
    new_customer_phone VARCHAR(20),

    -- For merge: customer details
    source_customer_id INTEGER,
    source_customer_name VARCHAR(255),
    source_customer_phone VARCHAR(20),
    target_customer_id INTEGER,
    target_customer_name VARCHAR(255),
    target_customer_phone VARCHAR(20),
    entries_moved INTEGER,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_entry_management_logs_action_type ON entry_management_logs(action_type);
CREATE INDEX IF NOT EXISTS idx_entry_management_logs_created_at ON entry_management_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_entry_management_logs_performed_by ON entry_management_logs(performed_by_id);
