-- Create entry_events table for tracking entry status/state
CREATE TABLE IF NOT EXISTS entry_events (
    id SERIAL PRIMARY KEY,
    entry_id INTEGER REFERENCES entries(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    notes TEXT,
    created_by_user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX idx_entry_events_entry_id ON entry_events(entry_id);
CREATE INDEX idx_entry_events_created_at ON entry_events(created_at);
CREATE INDEX idx_entry_events_status ON entry_events(status);

-- Common event types:
-- CREATED - Entry first created
-- IN_STORAGE - Entry placed in storage
-- PROCESSING - Entry being processed
-- QUALITY_CHECK - Quality check performed
-- READY - Ready for pickup/delivery
-- COMPLETED - Entry completed/closed
