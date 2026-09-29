CREATE TABLE learning.path_decisions (
 track_id uuid NOT NULL REFERENCES learning.tracks(id) ON DELETE CASCADE,
 fingerprint text NOT NULL,
 node_id text NOT NULL,
 PRIMARY KEY(track_id,fingerprint)
);
