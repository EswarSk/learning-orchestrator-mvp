-- A learner who enabled both sources on one installation must have a fresh,
-- shared 20-minute gap; one source alone must not claim that time is free.
-- ponytail: Aggregate across devices only if multi-device Calendar support becomes a requirement.
CREATE OR REPLACE FUNCTION context.calendar_sources_agree(p_event uuid, p_user text) RETURNS boolean
LANGUAGE sql STABLE AS $$
 SELECT COALESCE((SELECT CASE WHEN e.source NOT IN ('ios_calendar','google_calendar') THEN true ELSE
  e.canceled_at IS NULL AND e.valid_until>now()
  AND e.observed_at>now()-interval '5 minutes' AND e.observed_at<=now()+interval '1 minute'
  AND
  NOT EXISTS (
   SELECT 1 FROM experience.devices d
   WHERE d.user_id=p_user
   AND d.installation_id=CASE WHEN e.source='ios_calendar' THEN substring(e.source_resource_id FROM 5) ELSE substring(e.source_resource_id FROM 8) END
   AND CASE WHEN e.source='ios_calendar' THEN d.google_calendar_permission ELSE d.calendar_permission END='enabled'
   AND COALESCE((SELECT c.granted FROM experience.consents c WHERE c.user_id=p_user
    AND c.purpose=CASE WHEN e.source='ios_calendar' THEN 'google_calendar_context' ELSE 'ios_calendar_context' END
    ORDER BY c.decision_seq DESC LIMIT 1),false)
   AND NOT EXISTS (
    SELECT 1 FROM context.events other
    WHERE other.user_id=p_user
    AND other.source=CASE WHEN e.source='ios_calendar' THEN 'google_calendar' ELSE 'ios_calendar' END
    AND other.source_resource_id=CASE WHEN e.source='ios_calendar'
     THEN 'google:'||substring(e.source_resource_id FROM 5) ELSE 'ios:'||substring(e.source_resource_id FROM 8) END
    AND other.canceled_at IS NULL
    AND other.observed_at>now()-interval '5 minutes' AND other.observed_at<=now()+interval '1 minute'
    AND GREATEST(e.evaluate_at,other.evaluate_at,now())+interval '20 minutes'<=LEAST(e.valid_until,other.valid_until)
   )
  ) END FROM context.events e WHERE e.id=p_event AND e.user_id=p_user),false);
$$;
