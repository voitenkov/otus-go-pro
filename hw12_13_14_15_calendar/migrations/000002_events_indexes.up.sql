CREATE INDEX IF NOT EXISTS list_events_idx
ON events (user_id, start_time, finish_time);