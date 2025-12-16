-- ========================================
-- Database Cleanup Script
-- Deletes all data EXCEPT customers & users
-- Keeps all table schemas intact
-- ========================================

BEGIN;

-- Disable triggers to avoid constraint issues during cleanup
SET session_replication_role = 'replica';

-- Delete all data from tables (ORDER MATTERS - respect foreign keys)
-- Start with tables that reference others

-- 1. Gate pass related
DELETE FROM gate_pass_pickups;
DELETE FROM gate_passes;

-- 2. Invoice related
DELETE FROM invoice_items;
DELETE FROM invoices;

-- 3. Payment related
DELETE FROM rent_payments;

-- 4. Room related
DELETE FROM room_entry_edit_logs;
DELETE FROM room_entries;

-- 5. Entry related
DELETE FROM entry_events;
DELETE FROM entries;

-- 6. Logs
DELETE FROM login_logs;
DELETE FROM admin_action_logs;

-- 7. System settings (if you want to reset these)
-- DELETE FROM system_settings;  -- Uncomment if you want to delete settings

-- Re-enable triggers
SET session_replication_role = 'origin';

-- Reset sequences to start from 1
ALTER SEQUENCE IF EXISTS gate_passes_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS gate_pass_pickups_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS invoices_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS invoice_items_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS rent_payments_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS room_entries_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS room_entry_edit_logs_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS entries_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS entry_events_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS login_logs_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS admin_action_logs_id_seq RESTART WITH 1;

-- Reset custom sequences for numbering
SELECT setval('receipt_number_sequence', 1, false);
SELECT setval('invoice_number_sequence', 1, false);
SELECT setval('seed_entry_sequence', 1, false);
SELECT setval('cakes_entry_sequence', 1, false);
SELECT setval('samogri_entry_sequence', 1, false);

COMMIT;

-- Verify what's left
SELECT 'customers' as table_name, COUNT(*) as record_count FROM customers
UNION ALL
SELECT 'users', COUNT(*) FROM users
UNION ALL
SELECT 'entries', COUNT(*) FROM entries
UNION ALL
SELECT 'gate_passes', COUNT(*) FROM gate_passes
UNION ALL
SELECT 'rent_payments', COUNT(*) FROM rent_payments
UNION ALL
SELECT 'invoices', COUNT(*) FROM invoices;
