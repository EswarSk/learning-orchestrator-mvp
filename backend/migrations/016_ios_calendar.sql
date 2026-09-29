ALTER TABLE experience.devices ADD COLUMN calendar_permission text NOT NULL DEFAULT 'not_requested'
 CHECK (calendar_permission IN ('not_requested','enabled','denied'));
CREATE INDEX calendar_events_source_idx ON context.events(user_id,source,source_resource_id)
 WHERE canceled_at IS NULL AND source='ios_calendar';
