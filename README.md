# Learning Orchestrator

Learning Orchestrator connects what someone is learning with opportunities to use it in daily life. With explicit consent, calendar availability and chosen practice places can create timely learning moments. Assessed practice and self-reported real-world outcomes then change the learning track and future opportunities: context + current learning → apply → reflect → adapt.

The mobile client is Expo/React Native. The active backend is **four separately deployed Go HTTP services and one Temporal worker**, not one backend process. They share one Go module under [`backend/`](backend/) so contracts and builds stay simple; each has its own entrypoint, container, and Compose service. PostgreSQL uses separate domain schemas, although the current services still make some cross-schema queries. This is a practical service split, not yet strict database-per-service isolation.

## Where to look

| Area | What lives there |
|---|---|
| [`apps/mobile/`](apps/mobile/) | iOS-first app: onboarding, courses, opportunities, practice, settings, and native permissions |
| [`backend/cmd/`](backend/cmd/) | Five independent Go executables: `experience`, `learning`, `context`, `opportunity`, `opportunity-worker` |
| [`backend/internal/`](backend/internal/) | Each service's implementation; `platform` holds only shared HTTP/DB/Temporal plumbing |
| [`backend/migrations/`](backend/migrations/) | Active PostgreSQL migrations for the Go services |
| [`content/`](content/) | Fixed course curricula; legacy generated courses remain readable from PostgreSQL |
| [`packages/contracts/`](packages/contracts/) | Mobile-facing TypeScript contracts and OpenAPI artifact |
| [`compose.yaml`](compose.yaml) | Local PostgreSQL, Temporal, migration job, and five separate service containers |

The phone calls only `experience`. That API coordinates the private services. Context events enter Temporal; the worker asks Learning for a suitable activity and Opportunity stores the invitation. Finishing practice records evidence in Learning, which changes later course and opportunity choices.

## Local development

Requirements: Docker, Go 1.27, Node 24–26, pnpm 10.

```sh
docker compose up -d --build
curl http://localhost:3000/health/ready
cd backend && go test ./...
```

The app API is at `http://localhost:3000`. Temporal is at `localhost:7233`; PostgreSQL is at `localhost:54322`. Compose binds these development services to this Mac's loopback interface, not the LAN. Local auth uses the development token only when `APP_ENV=local`. The mobile app can run without an identity provider in development. Set `EXPO_PUBLIC_API_URL=http://localhost:3000` for the simulator, then run `pnpm dev:mobile`. A physical iPhone needs a separately secured, reachable test API; do not expose the shared local-development token on a LAN.

Practice uses OpenAI for assessed tutor turns and OpenAI Realtime over WebRTC for live voice. Put `OPENAI_API_KEY` and `OPENAI_MODEL` in a local `.env` before starting Compose (see `.env.example`). Keys stay on the Go learning service; the phone receives only a short-lived Realtime client secret for an active voice session. With no OpenAI key, courses can be browsed but practice is unavailable. The app asks for microphone access, streams the conversation, and sends completed learner transcripts to Learning for a separate understanding check. It closes its WebRTC peer when practice ends or the app backgrounds; server-side call termination and physical-iPhone audio quality remain unverified. A native iOS rebuild is required for `react-native-webrtc`. ElevenLabs speech/transcription adapters remain in the backend but are not the current iOS voice flow; their credentials are optional for this flow.

For signed-in builds, register a public native OIDC client with redirect URI `pausa://auth` and authorization-code + PKCE enabled. Set `EXPO_PUBLIC_OIDC_ISSUER` and `EXPO_PUBLIC_OIDC_CLIENT_ID` in the mobile build, and the same issuer plus the API access-token audience as `OIDC_ISSUER` and `OIDC_AUDIENCE` on Experience. The provider must issue short-lived JWT access tokens with `iss`, `sub`, `aud`, and `exp` claims and an `expires_in` response; refresh tokens support persistent sign-in. The hosted provider controls account creation. Local development needs none of these settings. A real-provider sign-in, refresh, revocation, and account-switch test remains required before release.

Create a learning plan in the Learn tab from the fixed course catalog. Spanish is the bundled, unreviewed draft. Its levels, activities, prompts, and prerequisites are fixed. OpenAI chooses the next eligible activity using completion and recent practice evidence; due reviews can reinforce completed skills. A validated choice is saved until the learner's state changes. Invalid or unavailable AI decisions fall back to catalog order. New arbitrary subjects are unavailable until fixed content is added. Previously generated courses remain readable on existing tracks. Multiple tracks can be active concurrently.

On iOS, Connections can opt into one current place for a selected learning path. The app requests location access only after the learner taps “Use this place”; a precise fix, account-bound device, consent record, active owned path, and server-owned region are required before geofencing starts. Entry/exit events are queued and retried; removing the place revokes consent, disables the server region, and clears on-device monitoring. A physical-device background-delivery test remains required.

Connections also offers separate, optional iPhone and Google Calendar choices. Both send only a near-term free window to Context while the app is open; neither sends event titles or guests. When both sources are enabled on one phone, an invitation requires a fresh, overlapping 20-minute gap from both; stale or missing availability is not treated as free. To exercise Google, replace the placeholder iOS bundle identifier in `apps/mobile/app.json`, enable the Google Calendar API and an OAuth consent-screen test user in Google Cloud, create an **iOS** OAuth client for that bundle identifier, set `EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID` in the mobile build, and rebuild the native app so its redirect scheme is registered. No client secret belongs in the app. A real Google-account authorization, token refresh/revocation, selected-calendar availability, and physical-iPhone test are still required. Foreground checks do not provide dependable closed-app refresh.

Notifications are a separate opt-in. The Go worker records an outbox item with each opportunity and checks current consent, quiet hours, expiry, cancellation, device binding, and a one-per-local-day cap before submitting a generic invitation to Expo. It defers a push until quiet hours end only if the opportunity will still be valid. The worker checks Expo receipts after 15 minutes and records APNs/FCM acceptance, rejection, or a missing receipt; an invalid device token is removed. Set `PUSH_ENABLED=true` for both Experience and the opportunity worker and supply an EAS project ID in the iOS build (`EXPO_PUBLIC_EAS_PROJECT_ID`) only after APNs credentials and a real iPhone are configured. Local Compose leaves push disabled. Neither an Expo ticket nor an accepted receipt proves that the phone displayed a notification.

## Backend boundaries

| Process | Ownership | PostgreSQL schema |
|---|---|---|
| `experience` (HTTP :3000) | External API, identity, profile, consent, devices, mobile composition | `experience` |
| `learning` (HTTP :3001) | Tracks, curricula, sessions, evidence, progress, reviewed application templates | `learning` |
| `context` (HTTP :3002) | Approved regions, consent-checked signals, source expiry, durable outbox | `context` |
| `opportunity` (HTTP :3003) | Opportunity inbox and actions | `opportunity` |
| `opportunity-worker` (no HTTP port) | Temporal workflow, current-policy recheck, opportunity decision and push outbox | `opportunity` |

The context outbox starts a Temporal workflow with a stable workflow ID. A workflow waits, rechecks the event and source-specific consent, then requests an activity from Learning. A saved practice place selects the next activity or due review from its chosen path; other context categories still require reviewed application templates. Notification delivery is a deliberate shared-database exception: the opportunity worker locks and checks Experience and Context state at send time, and clears invalid device tokens. This needs its own database privileges and operational review before deployment. The mobile app calls Experience only.

Each process is an OCI container with configuration supplied by environment variables. This keeps deployment independent of AWS, Azure, or GCP; use managed PostgreSQL and either Temporal Cloud or a self-hosted Temporal cluster. Production must provide a real OIDC issuer/audience, TLS ingress, private service networking, unique internal secrets, dedicated database roles, monitoring, and backups. Docker Compose credentials are for local use only.

## Release status

This is a working local learning/context workflow, **not a live-user release**. The bundled Spanish draft was AI-authored and is labeled unreviewed; runtime course generation is disabled. A synthetic live-provider probe passed the OpenAI tutor reply and ElevenLabs speech/transcription adapters. A Realtime session connected in the iOS simulator, but physical-device audio quality, call termination, activity-selection quality, and real-user flows remain unverified. OIDC is untested against a real provider. Push registration, delivery, and receipts have database/mock-provider coverage, not real APNs or iPhone proof. Both Calendar sources are foreground-only; Google authorization is implemented but unverified with a real account, and dependable closed-app refresh is unfinished. Account deletion, content quality/safety evaluation, and production operations still need implementation and evidence. Paid subscriptions and background email access are deferred from the first launch. See [release evidence](docs/release-evidence.md) for the current gates.
