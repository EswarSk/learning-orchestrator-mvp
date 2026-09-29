ALTER TABLE learning.sessions ADD COLUMN IF NOT EXISTS voice_requests integer NOT NULL DEFAULT 0 CHECK (voice_requests >= 0);
