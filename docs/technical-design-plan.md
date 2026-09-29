# Technical design and implementation specification

> Historical TypeScript MVP plan. Paths such as `services/backend/` and `db/migrations/` now live under `legacy/typescript-backend/`. For the active Go services, start with the repository [README](../README.md) and [release evidence](./release-evidence.md).

Version 1.1 · September 19, 2026 · Spanish-learning MVP · iOS first

**Purpose:** Give implementation agents enough shared decisions and contracts to build one integrated application. Use with [Agent build plan](./agent-build-plan.md), which defines ownership, dependencies, prompts, and merge gates.

**Status:** Implementation specification, not implemented software. No application repository, deployed services, provisioned accounts, or live provider tests have been created by this planning task.

**Reuse review:** [Build versus reuse](./build-vs-buy-review.md) covers every implementation package. This revision adds React Native Paper, Expo EAS and RevenueCat Paywalls to the defaults. Nango is a connector candidate behind gate C1; the direct Google schema/routes below remain the concrete baseline until an approved integration decision coherently replaces them. Other reviewed platforms are alternatives, not additional dependencies.

**Precedence:** User instructions and the application repository's AGENTS.md come first. This specification supersedes earlier provider-neutral suggestions where it makes a concrete choice. The [architecture](./technical-architecture.md) remains the system map; the [research report](./connector-orchestration-research.md) explains connector constraints. Agents must propose contract changes to the integration owner before implementing conflicting versions.

## 1. Release contract

An adult English-speaking learner signs in, starts useful Spanish practice, and optionally connects Calendar and location. Authorized signals create opportunities without manual activity logging. Accepted sessions update skill evidence. A later review applies the skill in another context. Agents maintain a private opportunity pool and an eligible shared pattern catalog.

**Required end-to-end behaviors:**

| ID | Behavior | Acceptance evidence |
|---|---|---|
| R01 | Immediate ordinary learning | New user finishes a starter exercise before connecting Calendar/location |
| R02 | Autonomous Calendar opportunity | Real selected-calendar event produces a scheduled opportunity while the app is closed/backgrounded |
| R03 | Autonomous location opportunity | Physical iPhone geofence transition produces an opportunity without a check-in |
| R04 | Personalized live/text practice | Two different skill histories produce appropriate differences; text remains available when microphone is declined |
| R05 | Durable feedback loop | Session evidence changes a later scheduled recall/transfer activity |
| R06 | Automatic shared improvement | Eligible first-party session yields a reviewed generic pattern usable by a second account; restricted ancestry is rejected |
| R07 | Complete free/purchase path | 25 lifetime free sessions, bounded duration, five practical milestones, paid continuation, restore purchases |
| R08 | User control and isolation | Pause, quiet hours, reconnect/disconnect, export/delete; account B cannot read account A's records |
| R09 | Recovery and explainability | Worker restart, duplicate signal, canceled event, and provider failure have correct recorded outcomes |
| R10 | Polished mobile experience | Accessible, responsive onboarding, curriculum, opportunities, practice, progress, settings; loading/empty/error/offline states work |

**MVP boundary:** one target language, one launch area, a curated venue list, two external signal sources, one subscription. Music/food integrations, Android release, visual workflow authoring, unrestricted agents, public profiles, and user-authored activities are future work. Do not substitute simulated signals for R02/R03.

## 2. Fixed implementation defaults

These choices unblock implementation; they do not authorize purchasing or deploying paid infrastructure.

| Layer | Choice | Why / constraint |
|---|---|---|
| Language / workspace | TypeScript strict; Node 24 LTS; pnpm 10 workspace | Shared request/response contracts and one lockfile |
| Mobile | Expo SDK 56, matching React Native version, Expo Router; development builds | Pin the compatible native package set in F00; Expo Go is not the native acceptance environment |
| Standard mobile UI | React Native Paper, themed to the product | Reuse forms/buttons/dialogs; custom learning screens and app-level accessibility verification remain required |
| Mobile build / delivery | Expo EAS Build, Submit and compatible Update | Reuse build/signing/submission mechanics; native changes need a native build and store approval remains separate |
| Native capabilities | expo-location, expo-task-manager, expo-notifications, expo-secure-store, expo-auth-session; react-native-webrtc with its compatible Expo config plugin | Native background location, secure tokens, OAuth browser flow, real-time audio |
| Mobile state | React state/context for local UI; TanStack Query for server state | Server is authoritative; avoid duplicate global state stores |
| HTTP backend | Fastify 5; Zod 4 request/response validation; OpenAPI generated from shared contracts | One API process; authenticate before entering domain services |
| SQL access / migrations | pg; parameterized SQL; node-pg-migrate | Explicit transactions and SQL constraints; no second ORM |
| Jobs | pg-boss in the same Postgres instance | Delayed jobs, retries and worker recovery; application owns business idempotency |
| Database / identity | Supabase managed Postgres and Supabase Auth | App data in a private schema, accessed through the backend; mobile uses Supabase only for identity |
| Calendar integration | Direct Google API baseline; Nango auth/proxy evaluated at C1 before production A05 implementation | Adopt one path; Nango requires revised credential/routes/jobs/deletion ownership, not an SDK added alongside the direct path |
| Text reasoning | OpenAI Responses API; initial model gpt-4.1-mini-2025-04-14 | Pinned low-cost baseline for constrained tasks; promote only after evaluation |
| Live voice | OpenAI Realtime API; initial model gpt-realtime-2.1; WebRTC plus backend sideband | Pin resolved model/version in session records; actual account access and Spanish quality are F00 gates |
| Billing | RevenueCat React Native SDK and react-native-purchases-ui Paywalls over Apple StoreKit; server entitlement reconciliation | One entitlement named learning_plus; one monthly product; app retains the free-session and paid-duration ledger |
| Push | Expo Push Service through APNs on iOS | Use delivery tickets/receipts; receipt is not proof the user saw a notification |
| Pilot hosting | Render paid web service + background worker; Supabase in a nearby compatible region | Persistent processes for jobs and voice control; no idle-sleep free hosting for an autonomous pilot |
| Testing | Vitest + Fastify injection; real Postgres integration tests; React Native Testing Library; Maestro smoke flows | Test durable state against Postgres, not SQLite |
| Diagnostics | Structured redacted Fastify/Pino logs; Sentry with explicit scrubbing and replay disabled; restricted founder view | Reuse vendor dashboards; no raw calendar content, transcripts, secrets, or SDP in telemetry |

The Render/Supabase default resolves the earlier open cloud choice for planning. Credits are not assumed to cover these vendors or OpenAI. If you secure a better-fitting grant before provisioning, replace hosting through a short decision record while retaining the API, worker and Postgres contracts. Do not create multi-cloud abstractions.

Freeze exact dependency patches in F00 after the native proof; no agent independently upgrades shared major versions. Model IDs above are starting defaults, not claims that a benchmark has already passed. If unavailable or inadequate, the integrator approves and records a replacement before dependent voice/AI work proceeds.

### Voice architecture refinement

The earlier diagram routed live media through the application backend. For the concrete build, **phone ↔ OpenAI WebRTC carries audio; API ↔ OpenAI sideband carries server control, tools and authoritative events**. The server brokers session creation and keeps permanent credentials private. This avoids implementing mobile PCM capture and a media relay.

OpenAI documents both Realtime and GPT-Live with different event contracts. This plan chooses **Realtime**. Agents must not mix /v1/live endpoints or Live session events into this implementation.

## 3. Repository and ownership boundaries

Create a new application repository outside the read-only synced sources directory. Copy these planning documents into its docs directory. Do not build inside sources/.

```text
apps/mobile/
  app/                    Expo Router routes
  src/features/           onboarding, learn, opportunities, practice, progress, settings
  src/native/             geofences, push, voice, OAuth browser, purchases
  src/ui/                 tokens, accessible components
  src/api/                typed client and query hooks
services/backend/
  src/api/                bootstrap, middleware, route registration
  src/modules/
    identity/             verified user context, profile, consent, device registration
    learning/             curriculum, sessions, evidence, progression
    connectors/           Google OAuth, sync, scoped reads
    signals/              event validation, normalization and source tracking
    workflows/            versioned definitions and run transitions
    ai/                   prompts, schemas, model calls and tools
    voice/                SDP setup, sideband control, call lifecycle
    notifications/        eligibility, outbox delivery and receipt processing
    billing/              RevenueCat events, reconciliation, usage reservations
    privacy/              source revocation, export, deletion and retention
    patterns/             eligibility, extraction and reviewed catalog
    operations/           restricted run views and review actions
  src/worker/             job registration and process lifecycle
  src/db/                 pools, owner-bound transactions, repositories
packages/contracts/       Zod schemas, inferred types, enum values, fixtures
content/spanish/           skills, milestones, curriculum nodes, templates, rubrics
db/migrations/            ordered SQL schema/data changes
tests/integration/        database and cross-module acceptance tests
tests/e2e/                Maestro flows and physical-device evidence instructions
infra/                    Render blueprint, Dockerfile, environment examples
docs/                     specification, ADRs, runbooks, integration evidence
```

Create directories when their package begins. Domain services are plain functions with explicit dependencies. Do not add an agent framework, event bus, generic plugin SDK, vector database, or separate microservices for this release.

**Only the integration owner edits shared contracts, root configuration/lockfile, migration ordering, and route/job registration.** Contributors submit proposed contract/migration changes with their work. The owner can delegate a specific shared file explicitly.

## 4. Mobile product specification

Bottom navigation: **Learn, Opportunities, Progress, Settings**. Practice is a focused full-screen route with a visible exit/pause control.

| Surface | Required behavior and states |
|---|---|
| Welcome / sign-in | Sign in with Apple plus email OTP through Supabase; authenticated stable UUID; no anonymous purchase identity |
| Onboarding | English support language, Spanish target, goal, rough level, interests, IANA time zone; consent explanations; start lesson before asking for all permissions |
| Learn | Five milestones with five nodes each; next activity, due review, remaining allowance; locked/upcoming/completed states; ordinary fallback clearly labeled |
| Connect apps | Connected account, selected calendars, granted capability, last sync, needs-reconnect, revoke; location and notification permissions have separate statuses |
| Opportunities | Available, preparing, snoozed, expired; concise why-this-appeared; start, dismiss, snooze; source-sensitive details shown only inside authenticated app |
| Practice | Text/listening/voice, captions, repeat/slower audio, microphone mute, interruption, reconnect, time remaining, finish; microphone starts after user action |
| Feedback | One useful correction, assistance recorded, uncertain assessment shown honestly, next review; flag incorrect feedback |
| Progress | Milestone completion, practiced skills, due reviews, XP/streak; distinguish completed content from demonstrated skill |
| Paywall | Actual localized store price, subscription period, renewal terms, remaining allowances, restore, terms/privacy links; no invented pricing |
| Settings | Quiet hours, proactive pause, source/category controls, export/delete, sign-out, support, subscription management |

Initial design tokens: 4-point spacing scale; minimum 44-point touch targets; system fonts supporting Dynamic Type; contrast meeting WCAG AA; VoiceOver labels; reduced-motion support. Use consistent cards, buttons, progress indicators, skeletons, and error recovery. A design agent may choose the brand/color palette within these constraints.

Compose and theme React Native Paper controls before adding custom primitives. Own the learning path/practice design, not a new component library. Use RevenueCat Paywalls for the purchase surface, with the app providing accurate allowance explanations and navigation. Test actual accessibility and localized disclosures after assembly; library support is not whole-app proof.

**Gamification:** 10 XP on a session's first completed result, transactionally unique by session. A learning day requires one completed useful session. Count streaks using the learner time zone captured at session start. Dismissed offers, missed days, low-confidence assessment and revoked permissions never remove earned skill evidence or XP. No leaderboards or virtual currency in v1.

Offline: previously saved results and bundled/reviewed content remain readable. New AI sessions require network; present explicit retry. Do not hold a microphone connection while app is backgrounded. Stop/pause media, preserve budget and state, and resume explicitly.

## 5. Domain and database contract

All IDs are UUIDs unless called provider IDs. All instants are timestamptz/UTC; recurrence and quiet-hour logic also stores IANA time zones. JSONB holds versioned variable content, never substitutes for ownership/uniqueness columns.

Every private row includes user_id, created_at, updated_at. Use composite foreign keys (user_id, referenced_id) for links between private records. Index owner/time and owner/state access paths. Authoritative writes come from the server.

### Tables, essential fields, constraints

| Table | Fields beyond common columns | Required invariant |
|---|---|---|
| learners | id = auth subject; goal, level_hint, timezone, interests, proactive_paused, quiet_start/end, status | No client-controlled identity or authoritative proficiency |
| consents | purpose, policy_version, granted_at, revoked_at | Append history; current grant is queryable; scopes are not consent |
| devices | installation_id, platform, push_token, permission_state, last_seen_at | Unique active installation binding; detach on logout/account switch |
| connections | provider, stable_provider_account_id, status, granted_scopes, encrypted_token, encryption_key_version, last_sync_at, revoked_at | Owner/provider/account unique; credentials never returned to mobile |
| calendar_resources | connection_id, provider_calendar_id, selected, timezone, sync_token, query_config, watch_ids/expiry | Cursor and watch are per calendar, not per account |
| oauth_attempts | provider, state_hash, encrypted_pkce_verifier?, return_route, expires_at, consumed_at | Single-use, time-bounded, owner-bound; no secrets in mobile return URL |
| idempotency_requests | operation, key, body_hash, state, response_status/body, expires_at | Unique owner/operation/key; persist short-lived API replay separately from permanent domain deduplication |
| artifacts | id, kind, sharing_class, revoked_at, expires_at | Private source-bearing records use this ID; sharing class assigned by code |
| context_events | id = artifact ID; source, connection_id?, resource_id?, source_event_id, revision, observed_at, received_at, valid_until, facts, confidence, canceled_at | Unique owner/source/resource/event/revision; immutable observations |
| artifact_sources | artifact_id, source_context_event_id | Flattened union of all source ancestry, including context used indirectly through memory |
| opportunities | id = artifact ID; skill_id, template_id/version, context_label, valid_from/until, status, run_id, reason_code | Expired or revoked ancestry prevents contextual action |
| workflow_runs | workflow_key/version, trigger_key, opportunity_id?, parent_run_id?, state, step, wake_at, expires_at, revision, error_code | Unique workflow_key + trigger_key; compare-and-swap transitions |
| workflow_steps | run_id, step_key, attempt, status, input_hash, output_refs, model_version?, cost_units | One committed result per logical run/step; attempts retained separately |
| sessions | id = artifact ID; source_type/source_id, curriculum_node_id?, mode, state, started_at, finished_at, budget_seconds, used_seconds, useful_at, prompt/content_versions | Partial unique owner index for reserved/starting/active/paused states; resume same budget |
| session_messages | session_id, message_id, role, modality, text, assistance, provider_item_id?, media_epoch? | Unique logical message; interim transcripts are not final assessment input |
| voice_calls | session_id, epoch, provider_call_id?, status, call_deadline, controller_lease_until, started_at, ended_at | One live provider call per session; provider call ID bound server-side |
| skill_evidence | id = artifact ID; session_id, skill_id, task_id, outcome, assistance, confidence, rubric_version | Unique session/skill/task; uncertainty retained; no client scoring |
| skill_state | skill_id, stage, latest_evidence_id, next_review_at, successes, needs_help_count | Rebuildable projection of eligible evidence, not source of truth |
| curriculum_progress | node_id, content_version, completed_session_id?, completed_at | Unique owner/node; a retry cannot unlock or reward twice |
| reward_events | session_id, xp, local_learning_date, captured_timezone | Unique owner/session; progress is rebuildable from events |
| usage_accounts | free_total = 25, free_consumed, free_reserved, paid_period_id?, paid_seconds_used/reserved | Locked for allowance decisions; counters cannot become negative |
| usage_ledger | session_id, operation, seconds, period_id?, reason | Unique idempotent reservation/finalization/release entries |
| entitlements | provider, app_user_id, product_id, entitlement, environment, period_start/end, revoked_at, verified_at | Server-reconciled; sandbox cannot grant production entitlement |
| notification_outbox | opportunity_id, run_revision, local_day, due_at, status, provider_ticket/receipt, attempts | Unique invitation key; daily budget reservation is atomic |
| patterns | id, template_version, skill_tags, difficulty, status, eligibility_evidence, content, review_actor/time | Only approved generic structures can be retrieved across users |
| webhook_receipts | provider, environment, provider_event_id, payload_hash, processed_at | Duplicate/reordered delivery does not duplicate domain changes |
| privacy_jobs | kind, source_id?, state, checkpoint, requested_at, completed_at | Immediately block access; resume cleanup idempotently |

pg-boss owns its internal schema; do not duplicate its job tables or claim application side effects are exactly once merely because the queue claims a job once. Curriculum/skill/template versions are reviewed JSON in Git, imported into read-only skills, milestones, curriculum_nodes, and activity_templates tables by the content migration/seed process. Their stable IDs and version fields must match the content manifest and shared contracts.

**Data access:** app schema is not exposed through Supabase's public Data API. API uses a restricted database role with RLS for private tables and transaction-local verified user ID. Set it using a parameterized set_config call with local=true; never session-global across pooled requests. Worker uses a separate server-only role for scheduled maintenance, with explicit user-bound repository methods for domain reads/writes. Do not use the migration/admin credential at runtime. Shared reviewed content is read-only to user-scoped code.

**Lineage rule:** every AI input builder returns its source IDs along with content. On output, union them with any input session/template ancestry before writing artifact_sources. Model-generated claims of sharing eligibility are ignored. Revocation follows these links, cancels pending work, removes affected derivatives as required, and rebuilds skill_state. A normal practice session that used restricted historical evidence is also restricted. Shared extraction may use only independently eligible first-party sessions with no restricted ancestry; seed templates solve cold start.

## 6. API contracts

F00 implements these as Zod schemas and generated OpenAPI. All routes below are under /v1 except provider callbacks and health endpoints. JSON keys use camelCase. Mobile calls never carry authoritative userId; derive it from the verified JWT.

**Common rules:**

- Bearer JWT verification includes signature, issuer, audience, expiry; authorization includes learner status.
- IDs belonging to another learner return 404. Missing/invalid credentials return 401; denied capability 403.
- Mutating session, device-event, opportunity-action and privacy requests require Idempotency-Key. Scope by owner + operation + key; store body hash and result. Same key/different payload returns 409. Financial/session deduplication also uses persistent domain keys.
- Validation errors: 422; state conflict: 409; rate limit: 429 with Retry-After; provider unavailable: 503; exhausted allowance: 402 with code allowance_exhausted.
- Error envelope: { error: { code, message, retryable, requestId } }. Never send provider secrets or raw diagnostic errors.
- List response: { items, nextCursor }. Opaque cursor, maximum 50 records.
- Store callback/webhook events durably before acknowledging; unavailable persistence returns retryable failure, not success.

| Method / path | Request | Response / behavior |
|---|---|---|
| GET /bootstrap | — | learner, consentSummary, connectionSummary, allowance, nextNode, dueReviewCount |
| PATCH /me | goal?, levelHint?, interests?, timezone?, quietHours?, proactivePaused? | Updated learner; reject fields outside allowlist |
| POST /consents | purpose, policyVersion, granted | Append grant/revocation; revocation starts required cancellation |
| PUT /devices/:installationId | pushToken?, notificationPermission, locationPermission | Upsert device for current owner; bounded fields |
| GET /curriculum | — | Version, five milestones, 25 node summaries and completion state |
| GET /progress | — | Skill stages, reviews, completed nodes, XP, streak; no CEFR claim |
| GET /connections | — | Provider/account label, resources, capabilities, health, lastSyncAt |
| POST /connections/google/start | returnRoute enum | authorizationUrl, attemptId, expiresAt; state bound to user/attempt |
| GET /connections/:id/resources | — | Authorized calendars and selection; no tokens |
| PUT /connections/:id/resources | resourceIds[] | Validate accessible resources; reconcile changes and cancel removed-source work |
| DELETE /connections/:id | — | 202 cleanup job ID; immediately revoke application access |
| GET /location/regions | areaId | Version and up to 10 curated public place regions for the selected launch area |
| POST /signals/location | schema in section 7 | Accepted event ID / duplicate flag; client cannot submit Calendar or assessment events |
| GET /opportunities | status? cursor? | Current private opportunity summaries and explainable reason |
| POST /opportunities/:id/actions | action: dismiss or snooze; snoozeUntil? | New state; snooze cannot exceed expiry |
| POST /sessions | source union; mode: text or voice | 201 session ID/state/budget/allowance; atomically reserve |
| GET /sessions/:id | — | Current session, remaining time, finalized messages, result if ready |
| POST /sessions/:id/turns | messageId, text, assistanceRequested? | Accepted turn ID; stream result through session events |
| POST /sessions/:id/voice/offer | mediaEpoch, sdpOffer | sdpAnswer, mediaEpoch, expiresAt; server creates/binds provider call |
| GET /sessions/:id/events | afterSequence? | Authenticated SSE: status, tutor delta/final, assessment-ready, budget, error |
| POST /sessions/:id/pause | — | Close media, checkpoint time/state, expire pause after 10 minutes |
| POST /sessions/:id/resume | — | Resume same session/budget; new media epoch only after old call ended |
| POST /sessions/:id/finish | reason: completed or userEnded | 202 assessing state; repeats return same operation |
| POST /sessions/:id/feedback | messageId?, reason, comment? | Store feedback flag; no direct proficiency mutation |
| GET /billing/status | — | Verified entitlement, period, free/paid usage and freshness |
| POST /billing/reconcile | — | Rate-limited refresh from RevenueCat for authenticated app-user ID |
| POST /privacy/export | — | 202 privacy job ID; recent authentication required |
| DELETE /me | — | 202 deletion job ID; immediately disable learner and close calls |
| GET /privacy/jobs/:id | — | State; short-lived export URL only when authorized and ready |

Provider routes: GET /oauth/google/callback, POST /webhooks/google/calendar, POST /webhooks/revenuecat. Each has its own authentication and replay rules; none accepts ordinary app-user identity as proof of provider origin. GET /health/live and /health/ready expose no secrets.

Admin routes: GET /ops/runs, GET /ops/runs/:id, GET /ops/patterns, POST /ops/patterns/:id/review, POST /ops/workflows/pause. Require a server-assigned admin role and audit record. A small server-rendered interface is sufficient; no separate frontend framework.

Session source union:

```json
{"source":{"type":"opportunity","id":"<uuid>"},"mode":"voice"}
```

Other variants: {type:"curriculum",nodeId:"m1-01"} and {type:"review",skillId:"repair.repeat"}. Server verifies eligibility and chooses template/difficulty. A review ID never authorizes a different learner's evidence.

SSE is a display channel. Reconnect obtains authoritative session state; do not claim a byte-perfect replay of transient audio/text deltas. Events include increasing per-session sequence, sessionId, type, payload; deduplicate finalized messages by messageId. Native voice media uses WebRTC, not SSE.

## 7. Signals and connector implementation

### Normalized internal event

```typescript
type ContextEvent = {
  schemaVersion: 1;
  id: string;                  // generated by server
  userId: string;              // verified owner, never supplied by model
  source: "google_calendar" | "ios_geofence" | "learning";
  connectionId: string | null;
  sourceResourceId: string | null;
  sourceEventId: string;
  sourceRevision: string;
  observedAt: string;
  receivedAt: string;          // server clock
  validUntil: string;
  facts: CalendarFacts | GeofenceFacts | LearningFacts;
  provenance: {
    sharingClass: "private_only" | "eligible_first_party";
    sourceArtifactIds: string[];
  };
};
```

CalendarFacts: kind, startAt, endAt, timezone, minimalTitle, categoryHint?, status, allDay, recurringInstanceKey?. Exclude descriptions/attendees by default and skip sensitive/private-marked events in the pilot. Titles remain untrusted input.

GeofenceFacts: kind, regionId, regionVersion, transition enter/exit, category. Client payload contains installationId, clientEventId, regionId/version, transition, observedAt. Server resolves category/geometry from its curated region catalog; client cannot invent a known venue or trusted category. No continuous routes.

LearningFacts: kind session_finished/review_due, sessionId?, skillIds. Only backend produces this source. Provenance inherits the session's ancestry.

### Google Calendar

Before production implementation, complete or explicitly resolve connector gate C1 in the build-versus-reuse review. The following is the direct Google fallback contract. A successful Nango auth/proxy proof replaces credential storage/lifecycle; adopting its managed sync additionally requires equivalent source-behavior tests and removal of duplicated local sync work. The integrator updates this section plus affected schema, API, configuration and privacy contracts together before agents use the new path.

Use provider OAuth with state binding, PKCE where supported, exact registered redirects, offline refresh, and selected read-only resources. Request calendar.events.readonly and calendar.calendarlist.readonly; identity scopes only as required for stable account binding. Tokens are encrypted with AES-256-GCM and key-version metadata; the wrapping secret is server environment secret storage. Reconnect must preserve the same account binding or create a separate connection.

Persist single-use OAuth attempts (10-minute expiry), bind them to the authenticated learner, and send only a one-time completion indicator in the mobile deep link. Never put tokens in the redirect. After callback, mobile fetches connection status through authenticated API.

Sync algorithm:
1. Register selected calendars; serialize sync per calendar resource.
2. Run initial paginated sync using a recorded, documented query configuration. Commit cursor only after all pages reconcile successfully.
3. Watch notifications authenticate stored channel ID, token and resource mapping. They enqueue sync; they carry no event body.
4. Incremental sync uses the stored token and compatible query parameters; process tombstones/cancellations. Handle recurrence exceptions and all-day/time-zone values explicitly.
5. On 410, replace only that calendar's source cache/cursor using a new full sync. Supersede obsolete opportunities; do not wipe unrelated learner or billing data.
6. Maintain a rolling seven-day opportunity horizon. A daily horizon refresh must discover upcoming recurring instances even without a change webhook. Do not combine unsupported time filters with syncToken.
7. Renew watches before returned expiry; tolerate overlapping old/new watches; periodic reconciliation recovers lost notifications.
8. Before a Calendar-derived invitation, re-fetch/check current event status. Provider failure defers or suppresses the contextual invitation.
9. Revocation/invalid refresh marks reconnect-required, blocks reads and stops pending source work. Handle source 403/404 as unavailable resources.

Contract tests cover pagination, recurrence modification, cancellation, 410, expired watches, resource deselection, out-of-order changes and DST. Exact Google query options must be pinned against its official list/sync contract by A05; they are not guessed from an SDK example.

### Location

Use a top-level registered Expo task and up to 10 curated regions, leaving room under the documented iOS limit. Select a launch area during onboarding; no continuous coordinate upload is needed. User must grant relevant OS permissions. Permission denial keeps other learning paths usable.

Process enter as a possible context, not proof of visiting a business. Delay eligibility by two minutes, cancel if an exit arrived, expire after 20 minutes, and do not claim confirmed dwell. Deduplicate repeats; suppress the same region for six hours. Accept late offline reports for diagnostic reconciliation but never notify from an expired event. Validate future timestamps with a five-minute skew tolerance. On sign-out revoke device binding, clear queued account events and unregister regions.

The launch-area venue dataset, its permitted use and coverage are a founder input. F00 can use labeled developer fixtures, but R03 requires a real approved region/device test.

## 8. Orchestration and agent contracts

### Durable workflow runtime

Use pg-boss for scheduling/retries; workflow_runs and workflow_steps remain business state. Named queues: calendar.sync, calendar.renew, workflow.resume, session.assess, notification.send, notification.receipt, billing.reconcile, patterns.extract, privacy.cleanup, voice.watchdog.

Do not reimplement pg-boss job claiming, scheduling or retries. Application runs/steps represent learning progress and cancellation; they are not a second general workflow engine. Trigger.dev or Inngest would be replacements requiring a new database-to-service delivery design, not extra services in this stack.

A domain transaction commits the state change and next job together using pg-boss send with its transaction-bound db option. F00 implements the small adapter for the existing pg transaction client required by the pinned release and proves atomic rollback/commit. Do not hold an unrelated queue connection and assume it shares the transaction. notification_outbox separately records external sends and delivery attempts; it is not a second general job queue.

No database transaction remains open across an LLM or external API call. Reserve a step attempt briefly, release the transaction, perform bounded work, then compare the run revision/step/permission state before committing. Discard stale results. Queue lease expiry is not permission to commit a canceled attempt.

Run states: pending, running, waiting, completed, canceled, expired, superseded, failed. Waiting records wakeAt or an expected response condition. Retriable failures back off with jitter, maximum three attempts for model generation and five for transient connector failures; permanent authorization/schema errors do not retry forever. Record reason codes.

Workflow definitions:
- context_practice_v1: validate source → interpret → choose skill/template → store opportunity → wait → recheck gates → prepare → invite → await accept/expire → link session result.
- review_v1: due evidence → choose alternate context or ordinary practice → same gating/delivery → assess → update next review.
- ordinary_practice_v1: user starts → authorize/reserve → reviewed task/tutor → assessment; no external-context requirement.

Opportunity states: candidate, ready, scheduled, offered, started, completed, dismissed, suppressed, canceled, expired, failed. Snooze is scheduled with a new due time. Expiry/cancellation cannot transition back to ready; generate a new explicitly linked opportunity if context changes.

### AI role contracts

| Role | Allowed input | Structured result | Tools |
|---|---|---|---|
| Interpret | Minimal allowed facts, known categories, expiry | situation enum, confidence 0–1, supporting source IDs, uncertainty | None by default |
| Plan | Skill state, due reviews, approved templates, eligible situation | skillId, templateId/version, format, difficulty 1–5, objective, supportLevel, source IDs | get_current_context, get_skill_state, find_reviewed_templates |
| Tutor | Validated plan, rubric, bounded private context, recent session turns | Spoken/text turns; bounded tool requests | get_reviewed_example, repeat_or_simplify, request_finish |
| Assess/reflect | Finalized response evidence, assistance, rubric | per-skill observed/needs_help/uncertain result, confidence, cited message IDs, correction, review proposal | None; server commits evidence |

Tool arguments use Zod schemas, known identifiers, small bounded result sets, current owner/grant checks, timeouts and per-session counters. Tool dispatcher is shared by workflows and voice; only backend executes tools. No arbitrary SQL, URL fetching, arbitrary tool names, calendar writes or external messaging.

Model output must be valid schema and reference supplied IDs. One repair attempt is permitted. If still invalid, use an eligible reviewed fallback or fail quietly with a reason. No fabricated context. Suppression and dismissal do not reduce skill.

### Shared planning and assessment shapes

F00 owns the executable Zod equivalent of these interfaces. Additional fields require an integration-owner contract change. All source/message/template IDs must be validated against the provided input set; length limits belong in the Zod schemas.

```typescript
type ActivityPlan = {
  schemaVersion: 1;
  skillId: string;
  templateId: string;
  templateVersion: number;
  format: "listen_respond" | "dialogue" | "recall";
  difficulty: 1 | 2 | 3 | 4 | 5;
  objective: string;           // max 160 characters
  supportLevel: "guided" | "hint_available" | "independent";
  situation: "dining" | "social" | "travel" | "daily_routine" | "ordinary_practice";
  situationConfidence: number; // 0..1, inference rather than fact
  sourceArtifactIds: string[];
  taskIds: string[];           // known tasks in the selected template
};

type AssessmentProposal = {
  schemaVersion: 1;
  sessionId: string;
  observations: Array<{
    taskId: string;
    skillId: string;
    outcome: "observed" | "needs_help" | "uncertain";
    assistance: "none" | "replay" | "hint" | "model_answer";
    confidence: number;        // 0..1; not a calibrated probability
    citedMessageIds: string[];
    correction: string | null; // max 400 characters
  }>;
  learnerFeedback: string;     // max 600 characters
};
```

The server derives assistance from runtime events and uses the more assisted/uncertain interpretation if the model disagrees. It assigns the authenticated session ID, authoritative rubric version, lineage, final evidence IDs and review time itself. A confidence threshold of 0.8 is an initial conservative promotion rule to evaluate, not a validated probability of mastery. Model confidence alone never overrides missing response evidence or recorded assistance.

The template payload consumed by mobile has schemaVersion, templateId/version, title, objective, format and tasks[]. Each task has a stable taskId, targetSkillIds, promptText, optional reviewedAudioAssetId, acceptedResponseModes (text/voice), maxTurns, and support options. Rubrics/model instructions stay server-side. Provider URLs/HTML are never rendered as arbitrary executable content. The authenticated API assembles the activity from reviewed content plus validated plan fields.

**Initial limits:** maximum four model calls in preparation (including repair), maximum four tutor tool calls/session, 20 finalized conversational turns, 30-second model timeout, three-minute session budget. Version these constants and measure them; do not let an agent change limits via prompt text.

**Planning priority:** due skill needing repair → current milestone prerequisite gap → current milestone practice → optional interest match. Filter context expiry and eligibility before model input. Keep at most 10 open opportunities per learner; coalesce overlapping context by owner + practical skill + 30-minute window. User-selected practice remains available.

**Notification policy:** at most two proactive invitations per learner local day; at least three hours apart; default quiet hours 21:00–09:00 local. Reserve the notification allowance transactionally before sending and release on confirmed pre-send failure. Recheck at dispatch. Generic lock-screen text; private explanation appears after opening the app. Count receipt, open and completion separately.

## 9. Live voice, learning evidence and curriculum

Session lifecycle: reserved → starting → active → assessing → completed. Active may transition to paused and resume to starting using the same ID and remaining budget. Terminal alternatives are abandoned, failed, or expired. Initial reservation expires after two minutes if setup never starts; pause expires after 10 minutes. A session leaves the active constraint only after its live call is confirmed closed or explicitly marked unknown and blocked from replacement pending reconciliation.

User-ended partial work can be assessed and retained while ending as abandoned; it does not complete a curriculum node or earn completion XP. Useful partial work still consumes its free slot. A template defines the minimum required tasks for completion. The server checks that coverage before marking completed; the model cannot unlock a node by merely saying it finished. Assessment retries do not reopen audio or reserve another session.

### Server-controlled voice flow

1. POST /sessions reserves allowance; microphone is still off.
2. User starts voice. Mobile creates WebRTC offer; authenticated voice/offer endpoint checks current session, quota and mediaEpoch.
3. Backend creates the call through POST /v1/realtime/calls with server-owned model/instructions/tools. Store the provider call ID from the provider response; never trust a client-supplied call ID.
4. Attach backend sideband wss://api.openai.com/v1/realtime?call_id=... and register server event handling. Return SDP answer only after control is attached. On failure hang up the created call and release setup-only reservations as appropriate.
5. Mobile sends/receives audio through WebRTC. Backend is sole executor for tool calls and authoritative transcript/evidence persistence; mobile data-channel captions are UI-only.
6. Compute budget from server timestamps and admitted connected intervals, including listening/silence. Pausing closes the provider call and checkpoints used time. No persistent background microphone.
7. At budget expiry, consent/account revocation or explicit finish, call POST /v1/realtime/calls/{call_id}/hangup server-side. Merely asking the client to stop is insufficient.
8. API control lease is renewed while healthy. Worker watchdog checks open calls at least every 10 seconds and hangs up overdue/orphaned calls. Record unknown provider finalization for reconciliation. Provider outages may delay remote hangup; cap per-project spend and alert.
9. Reconnect closes/reconciles the previous call before admitting a new media epoch. Load bounded finalized conversation context and remaining allowance. Unknown active-call status blocks creation of a duplicate.
10. On completion, deduplicate finalized transcript items, collect provider usage, and enqueue assessment once. Late provider events cannot recreate a deleted session.

A public release requires a physical-iPhone proof of audio routing, barge-in, captions, network loss, app backgrounding and server-initiated termination. Text-only fallback is useful, but does not satisfy the live-voice release requirement.

### Skill evidence

Start with 25 functional skills and five milestones:
1. Start/repair a conversation.
2. Get what you need: order, prices, preferences.
3. Find your way: direction, time, transport.
4. Connect with people: interests, plans, follow-ups.
5. Describe experiences and resolve misunderstandings.

Each milestone contains two guided nodes, two context-adapted nodes and one transfer-check node. A node is a reviewed activity structure, not a fixed conversation script. Structured fallback replaces missing context with an honest ordinary scenario.

Every template specifies objective, prerequisite IDs, accepted answer variation, common errors, support steps, listening material, rubric, and transfer variant. Teacher review is a release gate; AI-generated drafts are marked unreviewed and cannot be presented as validated curriculum.

Skill stages: unseen, practicing, demonstrated. Require two sufficiently confident unassisted successes on separate days, including one changed context, for demonstrated; this is a product heuristic, not a validated fluency claim. Low-confidence speech is uncertain. Do not claim pronunciation accuracy from transcript text alone.

Review scheduling heuristic: needs_help/assisted → next day; unassisted success → three days; another spaced success → seven days. Due dates are opportunities, not automatic notifications. Missed reviews do not erase evidence.

Milestone completion means its five nodes were completed; it does not certify mastery. Extra practice consumes a free session when it starts a new AI interaction; show that clearly. Resume uses the same session and allowance. A learner can use the free quota before completing every node if they choose extra practice.

## 10. Free allowance, subscriptions and usage transactions

Free access is 25 useful sessions per account, lifetime. Each session has at most 180 seconds of admitted AI interaction; total free interaction is bounded by 4,500 seconds. Pending invitations and previews consume nothing. Reading old feedback remains free.

Refresh a stale entitlement before beginning reservation. Start transaction: lock usage_accounts → recheck saved entitlement version/expiry → reject concurrent active session → verify source/permissions → reserve one free slot or paid duration budget → create session + ledger entry → commit. Perform AI/media provider calls afterward.

Entitlement network refresh happens before opening the reservation transaction; inside the transaction recheck the saved entitlement revision/expiry. Never hold the account row lock during a provider request. Usage entries have a stable operation key; duration entries additionally identify their media/text interval. Repeated close/finalize events cannot add the same interval twice. Paid reservation initially holds up to 180 remaining period-seconds, then releases unused seconds at finish.

Useful delivery milestone: finalized tutor prompt delivered by the service plus a nonempty accepted learner response. This is an operational delivery definition, not proof the learner heard or learned it. Finalize the free slot once at that milestone. A verified setup/technical failure before it releases the slot. A deliberate exit after useful interaction consumes it.

Before that milestone, still track provider cost and enforce session time, maximum three abandoned starts/day and a per-user setup cost cap; free reservations must not become an unlimited-cost loophole. Paused sessions resume within 10 minutes; expiration runs reconciliation before another session is admitted.

Paid baseline for development: learning_plus entitlement, 100 AI minutes per verified subscription billing period, maximum 180 seconds/session. Price is read from StoreKit offerings; the prior $14.99 suggestion remains a pricing hypothesis until measured costs are acceptable. Make paid limits explicit in UI and terms before selling.

RevenueCat app_user_id is the authenticated stable UUID. No anonymous purchase flow. Server verifies configured bearer authorization and HMAC over raw body, deduplicates event ID, and fetches current subscriber state. This avoids trusting webhook arrival order. Sandbox/production remain separate. Reconcile on purchase/restore, provider events and periodic stale-record checks.

Cancellation normally preserves access through verified period end; refund/revocation removes access according to verified provider state. Default restore policy preserves the original app-account owner: restore works after signing into the same stable UUID on another device; a different account receives a sign-in/recovery explanation, not an automatic entitlement transfer. Configure the matching RevenueCat restore policy and test it before release. Any later account-transfer feature requires a separate reviewed design.

## 11. Privacy, shared learning and operational controls

**Retention defaults:** transient external context up to seven days and no longer than its permitted purpose; raw audio is not durably stored by our app; finalized transcripts 30 days unless earlier deletion/restrictions apply; skill evidence remains until applicable revocation/deletion. Provider-side retention must be verified separately. Persist minimal billing records required for reconciliation/legal duties, without lesson content.

Sharing consent is separate from Calendar/location access. patterns.extract first checks artifact ancestry and consent in code, then strips learner-specific entities before the model receives eligible content. Automatically propose candidates; operator/teacher approves novel patterns in the pilot. Other learners see only approved generic structures. Removing names never waives provider restrictions.

Disconnect: mark connection unusable immediately → stop watch/refresh work → invalidate descendants and queued dispatch → delete applicable derivatives → rebuild affected skill summaries. Completed cleanup is observable. If disconnected in flight, late results are discarded.

Delete account: mark deleting and stop sessions/jobs immediately; revoke source access; remove app records, exports, tokens, and applicable provider records; delete auth identity after cleanup can finish. Use resumable privacy_jobs and a documented backup-aging/tombstone procedure. Never promise instant deletion from immutable backups. Export URLs expire after 15 minutes; export files after 24 hours.

Ops view: list runs by state/definition/reason; inspect redacted step outputs and provider timings; show connection health, usage, notification outcomes, false-context reports and assessment flags. Pause global proactive dispatch or a workflow version independently of ordinary practice. Pattern approvals and admin actions are audited. No arbitrary SQL console.

Reuse Supabase, Render, RevenueCat and Sentry dashboards for their existing diagnostics. The custom view covers only missing learning-specific state/actions. Retool may replace that view over the same restricted admin API if it demonstrably saves work; do not grant a builder broad production database access.

Pilot metrics: first useful session, first real autonomous activity, accepted/ignored/wrong-context offers, next-day/seven-day return, delayed unassisted recall, cost/session and cost/active learner, connector failures, hangup/recovery failures. Marketing events do not receive source text or transcripts.

## 12. Environments, delivery and proof

Local: Supabase CLI or matching local Postgres/Auth environment, backend API + worker, Expo development build. Synthetic signals/provider doubles are labeled development-only; they never count as external integration proof.

Staging: separate Supabase project, Render API/worker, OAuth client, Expo build channel, OpenAI project, RevenueCat sandbox and webhook secrets. Use selected consenting test accounts. Production must not reuse staging databases/keys.

Environment configuration: DATABASE_URL (runtime), MIGRATION_DATABASE_URL, Supabase issuer/JWKS/public client configuration, GOOGLE_CLIENT_ID/SECRET/REDIRECT_URI, TOKEN_ENCRYPTION_KEY and key version, OPENAI_API_KEY/TEXT_MODEL/VOICE_MODEL, RevenueCat public mobile key plus server/webhook secrets, EXPO_ACCESS_TOKEN, APP_BASE_URL, allowed redirects, admin subjects, PROACTIVE_ENABLED, provider budget limits. Only explicitly public values enter EXPO_PUBLIC_*.

Use a verified direct or session-pool database connection for persistent backend/pg-boss processes; copy the actual provider connection URL. Confirm network/IP compatibility and migration support. Bound API/worker connection pools together below the database connection allowance.

CI gates per PR: frozen-lockfile install, lint/typecheck, affected unit tests, contract generation drift check, real-Postgres integration tests for changed invariants, build API/worker. Native changes require a development build and named device evidence; fake/mocked native tests do not replace it.

Use EAS Build for reproducible signed development/distribution builds and EAS Submit for authorized store uploads. Compatible JavaScript/assets may use EAS Update with a matching runtime policy; native package/permission changes require a new binary. Record the exact native/runtime versions rather than building a custom release platform.

Deploy: migrate once in a controlled release job; expand-compatible schema first; deploy API/worker; run smoke checks; distribute mobile through TestFlight. Breaking schema removal waits for old clients to age out. Rollback uses earlier container/prompt/workflow version plus compatible migrations, not destructive down-migrations over user data.

Graceful shutdown: stop accepting new sessions/jobs, checkpoint/release work safely, close/hang up live media under a deadline, and let durable tasks recover. Backup restore test before a paid release.

### Required executable acceptance suite

| Test | Must prove |
|---|---|
| T01 | User B cannot read/update A's connection, session, opportunity, export, or tool scope |
| T02 | Concurrent starts cannot exceed free allowance; session 25 succeeds and 26 gates; retries/resume do not double-charge |
| T03 | Duplicate/changed/canceled Calendar events and invalid sync token yield one current opportunity and correct cleanup |
| T04 | Worker crash before/after a side effect cannot double-commit evidence, XP or usage; late canceled results are rejected |
| T05 | Quiet hours including DST, cooldown, daily cap, stale signals, and revocation prevent inappropriate dispatch |
| T06 | Event-title prompt injection cannot expand tools, ownership, permissions or shared-data eligibility |
| T07 | Voice cutoff is server-enforced; API crash/watchdog, reconnect and duplicate epoch do not create unbounded parallel calls |
| T08 | Assessment uncertainty and assistance are retained; next review changes; ignored notifications do not reduce skill |
| T09 | Restricted ancestry never reaches shared extraction; eligible approved template works for a second learner |
| T10 | Purchase/restore/renewal/expiry/refund and duplicate/out-of-order webhooks produce correct entitlement |
| T11 | Delete/revoke during a queued/in-flight operation stops future work and removes applicable descendants |
| T12 | Real iPhone background/locked/reboot/network/permission/audio-routing tests document actual supported behavior |

Targets for measurement: prepared activity opens around two seconds; typical spoken response around three seconds; log p50/p95. Source-to-notification latency includes provider/OS delivery and is measured separately. These targets are not guarantees.

**Ready for an invite-only pilot:** R01–R06, R08–R10, T01–T09/T11–T12, reviewed initial content, privacy controls, real connector proofs and enforced spend ceiling. Paid/public release additionally requires R07/T10, full 25-node reviewed content, backup restore, production billing/identity setup and store review readiness.

## 13. Provider facts verified for this design

Retrieved September 18, 2026. This is targeted implementation verification, not a successful integration test.

Additional reuse/provider research retrieved September 19 is recorded with source snapshots in the [build-versus-reuse review](./build-vs-buy-review.md).

- [OpenAI WebRTC](https://developers.openai.com/api/docs/guides/voice-webrtc?api=realtime): Realtime unified session setup uses /v1/realtime/calls; current example uses gpt-realtime-2.1. The page also contains separate GPT-Live examples; keep the API family consistent.
- [OpenAI server controls](https://developers.openai.com/api/docs/guides/voice-server-controls): backend sideband can observe/control the same Realtime call and execute tools server-side.
- [OpenAI hangup](https://developers.openai.com/api/reference/resources/realtime/subresources/calls/methods/hangup): server hangup supports WebRTC as well as SIP.
- [GPT-4.1 mini](https://developers.openai.com/api/docs/models/gpt-4.1-mini): structured output/tool support and the pinned snapshot used as the text baseline.
- [Expo WebRTC example](https://github.com/expo/examples/tree/master/with-webrtc) and [config plugin](https://github.com/expo/config-plugins/tree/main/packages/react-native-webrtc): native development build and compatible plugin required; inspected compatibility table includes Expo 56.
- [pg-boss](https://pgboss.io/introduction): Postgres-backed scheduling and retries. External side-effect idempotency still belongs to the application.
- [Supabase connections](https://supabase.com/docs/guides/database/connecting-to-postgres): persistent servers use direct or session-pool connections according to network compatibility.
- [RevenueCat webhooks](https://www.revenuecat.com/docs/integrations/webhooks): authentication/signature verification, deduplication, retry behavior and current-subscriber reconciliation.
- [Prior connector research](./connector-orchestration-research.md): Calendar sync/watch semantics, native location limits, provider data restrictions, and unproven music/food access.
