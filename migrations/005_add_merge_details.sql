-- Migration: Add detailed merge tracking

-- Add details column to entry_management_logs for storing transferred thocks and payments
ALTER TABLE entry_management_logs ADD COLUMN IF NOT EXISTS merge_details JSONB;
-- Example structure:
-- {
--   "entries": [
--     {"id": 123, "thock_number": "1889/1", "expected_quantity": 100},
--     {"id": 124, "thock_number": "1890/2", "expected_quantity": 50}
--   ],
--   "payments": [
--     {"id": 1, "amount": 5000, "receipt_number": "R123", "payment_date": "2024-12-28"},
--     {"id": 2, "amount": 3000, "receipt_number": "R124", "payment_date": "2024-12-29"}
--   ]
-- }

-- Add source village and SO for better display
ALTER TABLE entry_management_logs ADD COLUMN IF NOT EXISTS source_customer_village VARCHAR(255);
ALTER TABLE entry_management_logs ADD COLUMN IF NOT EXISTS source_customer_so VARCHAR(255);
ALTER TABLE entry_management_logs ADD COLUMN IF NOT EXISTS target_customer_village VARCHAR(255);
ALTER TABLE entry_management_logs ADD COLUMN IF NOT EXISTS target_customer_so VARCHAR(255);

-- Add payments_moved count
ALTER TABLE entry_management_logs ADD COLUMN IF NOT EXISTS payments_moved INTEGER DEFAULT 0;
