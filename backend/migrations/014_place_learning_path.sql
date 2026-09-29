ALTER TABLE learning.tracks ADD CONSTRAINT tracks_owner_id_unique UNIQUE(user_id,id);
ALTER TABLE context.regions ADD COLUMN track_id uuid;
ALTER TABLE context.regions ADD CONSTRAINT region_track_owner_fk FOREIGN KEY (user_id,track_id) REFERENCES learning.tracks(user_id,id);
ALTER TABLE context.events ADD COLUMN track_id uuid;

-- An old place did not say which path it served. Keep its coordinates but
-- stop monitoring it until the learner explicitly chooses a path.
WITH retired AS (
 UPDATE context.regions SET enabled=false WHERE user_id IS NOT NULL AND track_id IS NULL RETURNING id,user_id
), canceled AS (
 UPDATE context.events e SET canceled_at=now() FROM retired r
 WHERE e.user_id=r.user_id AND e.source='ios_geofence' AND e.source_resource_id=r.id
 AND e.canceled_at IS NULL RETURNING e.id
)
INSERT INTO context.outbox(event_id,kind) SELECT id,'cancel' FROM canceled ON CONFLICT DO NOTHING;
