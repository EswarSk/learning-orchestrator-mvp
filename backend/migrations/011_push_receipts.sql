ALTER TABLE opportunity.notification_outbox
 ADD COLUMN device_installation_id text,
 ADD COLUMN push_token_hash text,
 ADD COLUMN receipt_status text CHECK (receipt_status IN ('pending','accepted','rejected','missing','failed')),
 ADD COLUMN receipt_error text,
 ADD COLUMN receipt_attempts integer NOT NULL DEFAULT 0,
 ADD COLUMN receipt_next_check_at timestamptz,
 ADD COLUMN receipt_checked_at timestamptz;
CREATE INDEX notification_receipt_due_idx ON opportunity.notification_outbox(receipt_next_check_at)
 WHERE state='sent' AND receipt_status='pending';
