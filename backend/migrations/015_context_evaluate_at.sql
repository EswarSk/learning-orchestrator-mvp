-- Future Calendar moments need a stable wake time even if outbox dispatch retries.
-- Existing location events keep their short delay when evaluate_at is null.
ALTER TABLE context.events ADD COLUMN evaluate_at timestamptz;
ALTER TABLE context.events ADD CONSTRAINT context_evaluate_before_expiry CHECK (evaluate_at IS NULL OR evaluate_at < valid_until);
