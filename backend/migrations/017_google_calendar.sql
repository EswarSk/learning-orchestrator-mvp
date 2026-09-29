ALTER TABLE experience.devices ADD COLUMN google_calendar_permission text NOT NULL DEFAULT 'not_requested'
 CHECK (google_calendar_permission IN ('not_requested','enabled','denied'));
CREATE INDEX google_calendar_events_source_idx ON context.events(user_id,source,source_resource_id)
 WHERE canceled_at IS NULL AND source='google_calendar';
