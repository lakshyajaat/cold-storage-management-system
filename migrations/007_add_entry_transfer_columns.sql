-- Migration: Add transfer tracking columns to entries table
-- Required for entry transfer/reassign functionality

-- Add status column for tracking entry state
ALTER TABLE entries ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';
-- Values: 'active', 'transferred'

-- Add transfer tracking columns
ALTER TABLE entries ADD COLUMN IF NOT EXISTS transferred_to_customer_id INTEGER REFERENCES customers(id);
ALTER TABLE entries ADD COLUMN IF NOT EXISTS transferred_at TIMESTAMP WITH TIME ZONE;

-- Create index for filtering by status
CREATE INDEX IF NOT EXISTS idx_entries_status ON entries(status);
