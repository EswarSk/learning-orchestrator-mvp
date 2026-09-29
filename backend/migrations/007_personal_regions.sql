ALTER TABLE context.regions ADD COLUMN IF NOT EXISTS user_id text;
CREATE UNIQUE INDEX IF NOT EXISTS regions_one_per_user_idx ON context.regions(user_id) WHERE user_id IS NOT NULL;
