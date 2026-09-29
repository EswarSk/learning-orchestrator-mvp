CREATE TABLE IF NOT EXISTS learning.track_content (
 track_id uuid PRIMARY KEY REFERENCES learning.tracks(id) ON DELETE CASCADE,
 manifest jsonb NOT NULL,
 model text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
