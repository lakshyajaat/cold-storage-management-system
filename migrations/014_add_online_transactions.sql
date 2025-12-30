-- Migration: Add online_transactions table for Razorpay payment integration
-- Date: 2025-12-31

-- Create online_transactions table
CREATE TABLE IF NOT EXISTS online_transactions (
    id SERIAL PRIMARY KEY,

    -- Razorpay identifiers
    razorpay_order_id VARCHAR(50) UNIQUE NOT NULL,
    razorpay_payment_id VARCHAR(50),
    razorpay_signature VARCHAR(255),

    -- Customer information
    customer_id INT NOT NULL REFERENCES customers(id),
    customer_phone VARCHAR(15) NOT NULL,
    customer_name VARCHAR(100) NOT NULL,

    -- Payment scope (determines what this payment is for)
    entry_id INT REFERENCES entries(id),              -- Per truck payment
    family_member_id INT REFERENCES family_members(id), -- Per family member payment
    thock_number VARCHAR(50),
    family_member_name VARCHAR(100),
    payment_scope VARCHAR(20) DEFAULT 'account',      -- 'truck', 'family_member', 'account'

    -- Amount details
    amount DECIMAL(12,2) NOT NULL,                    -- Original payment amount
    fee_amount DECIMAL(12,2) NOT NULL DEFAULT 0,     -- Transaction fee (2.5%)
    total_amount DECIMAL(12,2) NOT NULL,             -- amount + fee_amount (what customer pays)

    -- Payment details from Razorpay webhook/callback
    utr_number VARCHAR(100),                          -- Bank UTR/Reference number
    payment_method VARCHAR(30),                       -- upi, card, netbanking, wallet
    bank VARCHAR(100),                                -- Bank name if available
    vpa VARCHAR(100),                                 -- UPI ID if UPI payment
    card_last4 VARCHAR(4),                           -- Last 4 digits of card if card payment
    card_network VARCHAR(20),                        -- visa, mastercard, etc.

    -- Status tracking
    status VARCHAR(20) DEFAULT 'pending',             -- pending, success, failed, refunded
    failure_reason TEXT,

    -- Linked records (populated after successful verification)
    rent_payment_id INT REFERENCES rent_payments(id),
    ledger_entry_id INT,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,

    -- Constraints
    CONSTRAINT chk_online_tx_status CHECK (status IN ('pending', 'success', 'failed', 'refunded')),
    CONSTRAINT chk_online_tx_scope CHECK (payment_scope IN ('truck', 'family_member', 'account')),
    CONSTRAINT chk_online_tx_amounts CHECK (amount > 0 AND fee_amount >= 0 AND total_amount >= amount)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_online_tx_order_id ON online_transactions(razorpay_order_id);
CREATE INDEX IF NOT EXISTS idx_online_tx_payment_id ON online_transactions(razorpay_payment_id);
CREATE INDEX IF NOT EXISTS idx_online_tx_customer_phone ON online_transactions(customer_phone);
CREATE INDEX IF NOT EXISTS idx_online_tx_customer_id ON online_transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_online_tx_status ON online_transactions(status);
CREATE INDEX IF NOT EXISTS idx_online_tx_utr ON online_transactions(utr_number);
CREATE INDEX IF NOT EXISTS idx_online_tx_created_at ON online_transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_online_tx_entry_id ON online_transactions(entry_id);
CREATE INDEX IF NOT EXISTS idx_online_tx_family_member_id ON online_transactions(family_member_id);

-- Add comments for documentation
COMMENT ON TABLE online_transactions IS 'Tracks all Razorpay payment transactions from customer portal';
COMMENT ON COLUMN online_transactions.utr_number IS 'Bank transaction reference from Razorpay payment details';
COMMENT ON COLUMN online_transactions.fee_amount IS 'Platform fee charged to customer (configurable, default 2.5%)';
COMMENT ON COLUMN online_transactions.payment_scope IS 'What the payment is for: truck (single entry), family_member (all entries of a family member), account (full balance)';

-- Add system settings for online payments (if not exists)
INSERT INTO system_settings (setting_key, setting_value, description)
VALUES
    ('online_payment_enabled', 'false', 'Enable/disable online payments from customer portal'),
    ('online_payment_fee_percent', '2.5', 'Transaction fee percentage charged to customers for online payments')
ON CONFLICT (setting_key) DO NOTHING;

-- Update ledger entry type check constraint to include ONLINE_PAYMENT
-- First check if constraint exists and drop it
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_entry_type'
        AND conrelid = 'ledger_entries'::regclass
    ) THEN
        ALTER TABLE ledger_entries DROP CONSTRAINT chk_entry_type;
    END IF;
END $$;

-- Add updated constraint with ONLINE_PAYMENT
ALTER TABLE ledger_entries ADD CONSTRAINT chk_entry_type
    CHECK (entry_type IN ('CHARGE', 'PAYMENT', 'CREDIT', 'REFUND', 'DEBT_APPROVAL', 'ONLINE_PAYMENT'));
