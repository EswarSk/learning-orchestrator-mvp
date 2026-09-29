ALTER TABLE learning.turns ADD COLUMN IF NOT EXISTS assistance text NOT NULL DEFAULT 'none';
