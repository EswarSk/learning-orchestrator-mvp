CREATE TABLE opportunity.notification_outbox (
 offer_id uuid PRIMARY KEY REFERENCES opportunity.offers(id) ON DELETE CASCADE,
 user_id text NOT NULL,
 state text NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','sent','suppressed','failed')),
 attempts integer NOT NULL DEFAULT 0,
 next_attempt_at timestamptz NOT NULL DEFAULT now(),
 provider_ticket text,
 sent_at timestamptz,
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notification_pending_idx ON opportunity.notification_outbox(next_attempt_at) WHERE state='pending';
WITH ranked AS (
 SELECT installation_id,row_number() OVER (PARTITION BY push_token ORDER BY updated_at DESC,installation_id) AS rank
 FROM experience.devices WHERE push_token IS NOT NULL
) UPDATE experience.devices d SET push_token=NULL,notification_permission='denied'
 FROM ranked r WHERE d.installation_id=r.installation_id AND r.rank>1;
CREATE UNIQUE INDEX devices_push_token_idx ON experience.devices(push_token) WHERE push_token IS NOT NULL;
