-- Migration: Add soft-delete columns for customers and entries
-- This enables full undo functionality for merges and transfers

-- Add soft-delete columns to customers table
ALTER TABLE customers ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';
-- Values: 'active', 'merged'
ALTER TABLE customers ADD COLUMN IF NOT EXISTS merged_into_customer_id INTEGER REFERENCES customers(id);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS merged_at TIMESTAMP WITH TIME ZONE;

-- Create index for filtering active customers
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);

-- Add original_customer_id to entries for tracking transfers/merges
ALTER TABLE entries ADD COLUMN IF NOT EXISTS original_customer_id INTEGER REFERENCES customers(id);

-- Update existing entries to set original_customer_id = customer_id
UPDATE entries SET original_customer_id = customer_id WHERE original_customer_id IS NULL;
