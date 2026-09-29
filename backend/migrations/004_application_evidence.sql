ALTER TABLE learning.evidence ALTER COLUMN session_id DROP NOT NULL;
ALTER TABLE learning.evidence ADD COLUMN IF NOT EXISTS opportunity_id uuid UNIQUE;
ALTER TABLE learning.evidence ADD CONSTRAINT evidence_exactly_one_source CHECK ((session_id IS NULL) <> (opportunity_id IS NULL));
CREATE TABLE IF NOT EXISTS learning.skill_state (
 track_id uuid NOT NULL REFERENCES learning.tracks(id) ON DELETE CASCADE,
 skill_id text NOT NULL,
 stage text NOT NULL,
 next_review_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(track_id,skill_id)
);
CREATE INDEX IF NOT EXISTS skill_state_due_idx ON learning.skill_state(track_id,next_review_at);
