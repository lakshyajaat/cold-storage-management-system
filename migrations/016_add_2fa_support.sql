-- Migration: 016_add_2fa_support.sql
-- Add Two-Factor Authentication (2FA) support for admin users

-- Add 2FA fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS backup_codes TEXT;

-- Index for quick lookup of 2FA-enabled users
CREATE INDEX IF NOT EXISTS idx_users_totp_enabled ON users(totp_enabled) WHERE totp_enabled = TRUE;

-- Rate limiting table for 2FA verification attempts
CREATE TABLE IF NOT EXISTS totp_verification_attempts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    success BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for rate limiting queries
CREATE INDEX IF NOT EXISTS idx_totp_attempts_user_time ON totp_verification_attempts(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_totp_attempts_ip_time ON totp_verification_attempts(ip_address, created_at);

-- System setting for mandatory 2FA for admins
INSERT INTO system_settings (setting_key, setting_value, description)
VALUES ('admin_2fa_required', 'false', 'When enabled, all admin users must set up 2FA')
ON CONFLICT (setting_key) DO NOTHING;

-- Add to protected settings (requires dual-admin approval to change)
INSERT INTO protected_settings (setting_key, description)
VALUES ('admin_2fa_required', 'Require 2FA for all admin users - requires dual admin approval')
ON CONFLICT (setting_key) DO NOTHING;
