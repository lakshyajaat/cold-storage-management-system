-- Migration 008: Add customer activity logs table
-- This table stores customer portal activity for auditing and monitoring

CREATE TABLE IF NOT EXISTS customer_activity_logs (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER DEFAULT 0,
    phone VARCHAR(15) NOT NULL,
    action VARCHAR(50) NOT NULL,
    details TEXT,
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast querying
CREATE INDEX IF NOT EXISTS idx_customer_activity_logs_customer_id ON customer_activity_logs(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_activity_logs_phone ON customer_activity_logs(phone);
CREATE INDEX IF NOT EXISTS idx_customer_activity_logs_action ON customer_activity_logs(action);
CREATE INDEX IF NOT EXISTS idx_customer_activity_logs_created_at ON customer_activity_logs(created_at DESC);

-- Insert default SMS rate limiter settings if they don't exist
-- All limits set to 0 = disabled by default
INSERT INTO system_settings (setting_key, setting_value, description, updated_at)
VALUES
    ('sms_otp_cooldown_minutes', '0', 'Cooldown between OTP requests in minutes (0 = disabled)', CURRENT_TIMESTAMP),
    ('sms_max_otp_per_window', '0', 'Max OTP requests per phone in window (0 = unlimited)', CURRENT_TIMESTAMP),
    ('sms_otp_window_minutes', '60', 'Time window for max OTP limit in minutes', CURRENT_TIMESTAMP),
    ('sms_max_otp_per_day', '0', 'Max OTP requests per phone per day (0 = unlimited)', CURRENT_TIMESTAMP),
    ('sms_max_otp_per_ip_window', '0', 'Max OTP requests per IP in window (0 = unlimited)', CURRENT_TIMESTAMP),
    ('sms_otp_ip_window_minutes', '60', 'Time window for IP limit in minutes', CURRENT_TIMESTAMP),
    ('sms_max_otp_per_ip_day', '0', 'Max OTP requests per IP per day (0 = unlimited)', CURRENT_TIMESTAMP),
    ('sms_max_daily_total', '0', 'Total SMS budget per day (0 = unlimited)', CURRENT_TIMESTAMP)
ON CONFLICT (setting_key) DO NOTHING;
