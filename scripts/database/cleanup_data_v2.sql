-- ========================================
-- Database Cleanup Script v2
-- Deletes all data EXCEPT customers & users
-- Keeps all table schemas intact
-- Handles missing sequences gracefully
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

-- Reset sequences to start from 1 (only if they exist)
DO $$
BEGIN
    -- Table ID sequences
    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'gate_passes_id_seq') THEN
        ALTER SEQUENCE gate_passes_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'gate_pass_pickups_id_seq') THEN
        ALTER SEQUENCE gate_pass_pickups_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'invoices_id_seq') THEN
        ALTER SEQUENCE invoices_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'invoice_items_id_seq') THEN
        ALTER SEQUENCE invoice_items_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'rent_payments_id_seq') THEN
        ALTER SEQUENCE rent_payments_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'room_entries_id_seq') THEN
        ALTER SEQUENCE room_entries_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'room_entry_edit_logs_id_seq') THEN
        ALTER SEQUENCE room_entry_edit_logs_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'entries_id_seq') THEN
        ALTER SEQUENCE entries_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'entry_events_id_seq') THEN
        ALTER SEQUENCE entry_events_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'login_logs_id_seq') THEN
        ALTER SEQUENCE login_logs_id_seq RESTART WITH 1;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'admin_action_logs_id_seq') THEN
        ALTER SEQUENCE admin_action_logs_id_seq RESTART WITH 1;
    END IF;

    -- Custom sequences for numbering (only if they exist)
    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'receipt_number_sequence') THEN
        PERFORM setval('receipt_number_sequence', 1, false);
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'invoice_number_sequence') THEN
        PERFORM setval('invoice_number_sequence', 1, false);
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'seed_entry_sequence') THEN
        PERFORM setval('seed_entry_sequence', 1, false);
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'cakes_entry_sequence') THEN
        PERFORM setval('cakes_entry_sequence', 1, false);
    END IF;

    IF EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'samogri_entry_sequence') THEN
        PERFORM setval('samogri_entry_sequence', 1, false);
    END IF;
END $$;

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
