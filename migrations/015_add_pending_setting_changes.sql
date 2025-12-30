-- Migration: Add pending setting changes for dual admin approval
-- This enables secure management of sensitive settings like Razorpay credentials

-- Table to store pending setting change requests
CREATE TABLE IF NOT EXISTS pending_setting_changes (
    id SERIAL PRIMARY KEY,

    -- Setting being changed
    setting_key VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT NOT NULL,

    -- Request details
    requested_by INT NOT NULL REFERENCES users(id),
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reason TEXT,

    -- Approval details
    approved_by INT REFERENCES users(id),
    approved_at TIMESTAMP,

    -- Status: pending, approved, rejected, expired
    status VARCHAR(20) DEFAULT 'pending',
    rejection_reason TEXT,

    -- Expiry (pending requests expire after 24 hours)
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 hours')
);

-- Index for quick lookups
CREATE INDEX IF NOT EXISTS idx_pending_setting_status ON pending_setting_changes(status);
CREATE INDEX IF NOT EXISTS idx_pending_setting_key ON pending_setting_changes(setting_key);
CREATE INDEX IF NOT EXISTS idx_pending_setting_expires ON pending_setting_changes(expires_at) WHERE status = 'pending';

-- Add Razorpay credential settings (encrypted values)
INSERT INTO system_settings (setting_key, setting_value, description)
VALUES
    ('razorpay_key_id', '', 'Razorpay API Key ID'),
    ('razorpay_key_secret', '', 'Razorpay API Key Secret (encrypted)'),
    ('razorpay_webhook_secret', '', 'Razorpay Webhook Secret (encrypted)')
ON CONFLICT (setting_key) DO NOTHING;

-- Settings that require dual admin approval
CREATE TABLE IF NOT EXISTS protected_settings (
    setting_key VARCHAR(100) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert protected settings
INSERT INTO protected_settings (setting_key, description) VALUES
    ('razorpay_key_id', 'Razorpay API Key ID - requires dual admin approval'),
    ('razorpay_key_secret', 'Razorpay API Key Secret - requires dual admin approval'),
    ('razorpay_webhook_secret', 'Razorpay Webhook Secret - requires dual admin approval'),
    ('online_payment_enabled', 'Enable/disable online payments - requires dual admin approval'),
    ('online_payment_fee_percent', 'Transaction fee percentage - requires dual admin approval')
ON CONFLICT (setting_key) DO NOTHING;
