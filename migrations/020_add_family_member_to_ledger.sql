-- Add family_member tracking to ledger_entries for per-family-member payment attribution
ALTER TABLE ledger_entries ADD COLUMN IF NOT EXISTS family_member_id INT;
ALTER TABLE ledger_entries ADD COLUMN IF NOT EXISTS family_member_name VARCHAR(100);

-- Add index for faster lookups by family member
CREATE INDEX IF NOT EXISTS idx_ledger_family_member ON ledger_entries(customer_phone, family_member_id);
