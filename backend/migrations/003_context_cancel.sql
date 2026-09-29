ALTER TABLE context.outbox ADD COLUMN IF NOT EXISTS kind text NOT NULL DEFAULT 'observe';
ALTER TABLE context.outbox DROP CONSTRAINT IF EXISTS outbox_event_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS outbox_event_kind_idx ON context.outbox(event_id,kind);
