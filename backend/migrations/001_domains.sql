CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SCHEMA IF NOT EXISTS experience;
CREATE SCHEMA IF NOT EXISTS learning;
CREATE SCHEMA IF NOT EXISTS context;
CREATE SCHEMA IF NOT EXISTS opportunity;

CREATE TABLE IF NOT EXISTS experience.profiles (
 user_id text PRIMARY KEY,
 timezone text NOT NULL DEFAULT 'America/Los_Angeles',
 proactive_paused boolean NOT NULL DEFAULT false,
 quiet_start time NOT NULL DEFAULT '21:00',
 quiet_end time NOT NULL DEFAULT '09:00',
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS experience.devices (
 installation_id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES experience.profiles(user_id) ON DELETE CASCADE,
 push_token text,
 notification_permission text NOT NULL,
 location_permission text NOT NULL,
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS devices_user_idx ON experience.devices(user_id);
CREATE TABLE IF NOT EXISTS experience.consents (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 decision_seq bigserial UNIQUE,
 user_id text NOT NULL REFERENCES experience.profiles(user_id) ON DELETE CASCADE,
 purpose text NOT NULL,
 policy_version text NOT NULL,
 granted boolean NOT NULL,
 request_key text NOT NULL,
 UNIQUE(user_id,purpose,request_key),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS consents_latest_idx ON experience.consents(user_id,purpose,decision_seq DESC);

CREATE TABLE IF NOT EXISTS learning.tracks (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id text NOT NULL,
 subject_id text NOT NULL,
 goal text NOT NULL,
 status text NOT NULL DEFAULT 'active',
 UNIQUE(user_id,subject_id,goal),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS tracks_user_idx ON learning.tracks(user_id,status);
CREATE TABLE IF NOT EXISTS learning.progress (
 track_id uuid NOT NULL REFERENCES learning.tracks(id) ON DELETE CASCADE,
 node_id text NOT NULL,
 completed_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(track_id,node_id)
);
CREATE TABLE IF NOT EXISTS learning.sessions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id text NOT NULL,
 track_id uuid NOT NULL REFERENCES learning.tracks(id),
 node_id text NOT NULL,
 source_type text NOT NULL,
 source_id text NOT NULL,
 mode text NOT NULL,
 state text NOT NULL DEFAULT 'active',
 created_at timestamptz NOT NULL DEFAULT now(),
 completed_at timestamptz
);
CREATE INDEX IF NOT EXISTS sessions_user_idx ON learning.sessions(user_id,created_at DESC);
CREATE TABLE IF NOT EXISTS learning.idempotency (
 user_id text NOT NULL,
 operation text NOT NULL,
 request_key text NOT NULL,
 request_hash text NOT NULL,
 response jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(user_id,operation,request_key)
);
CREATE TABLE IF NOT EXISTS learning.turns (
 session_id uuid NOT NULL REFERENCES learning.sessions(id) ON DELETE CASCADE,
 message_id text NOT NULL,
 learner_text text NOT NULL,
 tutor_text text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(session_id,message_id)
);
CREATE TABLE IF NOT EXISTS learning.evidence (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id text NOT NULL,
 track_id uuid NOT NULL REFERENCES learning.tracks(id),
 skill_id text NOT NULL,
 session_id uuid NOT NULL UNIQUE REFERENCES learning.sessions(id),
 kind text NOT NULL,
 outcome text NOT NULL,
 assistance text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS evidence_skill_idx ON learning.evidence(track_id,skill_id,created_at DESC);
CREATE TABLE IF NOT EXISTS learning.application_templates (
 id text PRIMARY KEY,
 subject_id text NOT NULL,
 skill_id text NOT NULL,
 context_category text NOT NULL,
 title text NOT NULL,
 prompt text NOT NULL,
 reviewed boolean NOT NULL DEFAULT false,
 version integer NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS templates_match_idx ON learning.application_templates(context_category,subject_id) WHERE reviewed;

CREATE TABLE IF NOT EXISTS context.regions (
 id text PRIMARY KEY,
 area_id text NOT NULL,
 category text NOT NULL,
 version integer NOT NULL DEFAULT 1,
 latitude double precision NOT NULL,
 longitude double precision NOT NULL,
 radius_meters integer NOT NULL,
 enabled boolean NOT NULL DEFAULT false
);
CREATE TABLE IF NOT EXISTS context.events (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id text NOT NULL,
 source text NOT NULL,
 source_event_id text NOT NULL,
 source_resource_id text NOT NULL,
 category text NOT NULL,
 observed_at timestamptz NOT NULL,
 valid_until timestamptz NOT NULL,
 canceled_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(user_id,source,source_event_id)
);
CREATE INDEX IF NOT EXISTS events_active_idx ON context.events(user_id,valid_until) WHERE canceled_at IS NULL;
CREATE TABLE IF NOT EXISTS context.outbox (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 event_id uuid NOT NULL UNIQUE REFERENCES context.events(id) ON DELETE CASCADE,
 state text NOT NULL DEFAULT 'pending',
 attempts integer NOT NULL DEFAULT 0,
 next_attempt_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS opportunity.offers (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id text NOT NULL,
 track_id uuid NOT NULL,
 node_id text NOT NULL,
 context_event_id uuid NOT NULL UNIQUE,
 context_label text NOT NULL,
 reason_code text NOT NULL,
 status text NOT NULL DEFAULT 'ready',
 valid_until timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS offers_user_idx ON opportunity.offers(user_id,status,valid_until);
