# Build versus reuse — every package in the dependency map

September 19, 2026 · Solo founder · Spanish learning MVP · Public documentation review

**Recommendation:** reuse identity, database hosting, mobile primitives, background jobs, model inference, voice transport, purchases, push delivery, builds and monitoring. Build the learning policy and the controls that connect these services safely. The 12 packages in the [agent plan](./agent-build-plan.md) are responsibility boundaries; they are not 12 platforms to build from scratch.

This review covers F00 and A01–A11, including their underlying tasks. It changes the plan to explicitly reuse React Native Paper, Expo EAS, and RevenueCat Paywalls. Supabase Auth, pg-boss, OpenAI, RevenueCat purchases, Expo Push, Render and Supabase hosting were already selected. Nango is the strongest additional candidate to validate before writing Calendar credential infrastructure. Other products below are alternatives, not additional dependencies.

**Decision status:** no vendor accounts, purchases, installed dependencies, application code or live integrations were created. The [technical specification](./technical-design-plan.md) remains authoritative. Its direct Google Calendar implementation remains the fallback and current concrete contract until the Nango gate below passes and the integration owner updates that contract. Do not implement both connector paths.

## 1. Package-by-package decision

“Reuse” includes open-source libraries and native OS capabilities as well as paid services. “Build” means application-specific behavior on top of those components.

| Package | Decision and reusable component | Work removed from our scope | Work our application still owns |
|---|---|---|---|
| **F00 — Foundation and proofs** | **Mostly reuse:** Expo development builds and **EAS Build/Submit/Update**, existing TypeScript/Fastify/Zod/Postgres tooling, Supabase and Render. | Custom mobile build/signing infrastructure, database hosting, HTTP framework, schema validator, hand-built CI runner. | Repository setup, shared contracts, migrations, CI configuration, secrets configuration, native compatibility and physical-device/voice proofs. EAS does not approve the app or make incompatible native changes safe for over-the-air delivery. |
| **A01 — Identity, access and allowance** | **Reuse Supabase Auth** for Apple sign-in, email OTP, sessions and identity tokens; Postgres constraints/RLS for isolation. | Password storage, OTP delivery engine, sign-in token issuance/refresh and account recovery infrastructure. Do not add a second auth vendor. | Verify identity; enforce owner access, consent, account status and device binding. Build the atomic **25-session reservation/usage ledger** and one-active-session invariant. Authentication is not authorization or metered learning access. |
| **A02 — Curriculum and rubrics** | **Build the learning content; reuse content tools and licensed assets.** Reviewed JSON in Git is sufficient. **Learnosity** is an available authoring/assessment platform, deferred. | A custom CMS, quiz-editor platform, asset production where suitable licensed material exists. | Five milestones, 25 nodes, skill prerequisites, rubrics, helpful feedback, alternate contexts, rights and qualified Spanish review. An authoring platform does not automatically provide a validated curriculum or evidence of retention. |
| **A03 — Mobile shell and screens** | **Reuse Expo Router, React Native Paper and TanStack Query.** Theme standard components; build the learning experience. | A custom routing library, server-state cache, general button/input/dialog/menu library and generic component-state machinery. | Learn path, Connect apps, opportunities, practice, feedback and progress composition; brand, interaction details, loading/error/offline handling, VoiceOver/Dynamic Type verification. A component kit does not provide Duolingo-level product design automatically. |
| **A04 — Workflows and notifications** | **Reuse pg-boss and Expo Push/APNs.** Build versioned learning transitions around them. **Trigger.dev/Inngest** are managed alternatives, deferred. | A queue broker, retry scheduler, timer service, job claiming/recovery engine and push transport. | Opportunity selection, context expiry, waits linked to learner state, quiet hours, cancellation, source checks, delivery outbox, deduplication and business-state transitions. A job being retried does not make its external effects safe to repeat. |
| **A05 — Calendar connector** | **Hybrid; validate Nango first.** It offers OAuth/credential refresh, an authorized API proxy, read actions and sync templates. **Nylas** is the calendar-focused alternative. Direct Google integration remains the specified fallback. | With Nango adopted: our Google credential vault/refresh implementation and potentially some sync plumbing, depending on the validated integration mode. | Learner/account binding, read-only scopes, selected calendars, minimal fields, recurrence and rolling horizon, change/cancellation handling, watch lifecycle, lineage, disconnect and freshness checks. Template availability is not acceptance evidence. |
| **A06 — Sessions and learning intelligence** | **Build the core behavior using the existing OpenAI SDK/Responses API and Zod.** Use ordinary functions for the four AI roles. | Model training/hosting, inference infrastructure and a new generic agent framework. | Context interpretation rules, activity selection, tutor prompts, tool permissions, evidence-based assessment, assistance tracking, review timing, provenance and quality evaluations. A generic agent platform does not decide what counts as learning for this product. |
| **A07 — Live voice** | **Reuse OpenAI Realtime and native WebRTC.** **LiveKit Agents** is a credible replacement transport/agent layer if the voice proof shows a need. | Speech infrastructure and our own real-time media relay. LiveKit could also take over more connection/agent infrastructure if selected. | Native audio UX, trusted call/session binding, server budget enforcement, sideband tools, hangup/watchdog, reconnect accounting, transcript deduplication and real-device tests. Do not connect both direct Realtime and a second room/session system by default. |
| **A08 — Location and push** | **Reuse Expo Location/TaskManager/Notifications and iOS Core Location/APNs.** **Radar** is a later geofencing/place-intelligence option. | Background location framework, operating-system region monitoring and notification transport. | Permission UX, ten curated regions, owner-safe queued events, expiry/dedupe, logout cleanup, interruption policy and device validation. Neither Expo nor Radar bypasses iOS permissions or background limitations. |
| **A09 — Purchases and entitlements** | **Reuse RevenueCat SDK plus RevenueCat Paywalls** (`react-native-purchases-ui`) over StoreKit. | Subscription purchase/restore infrastructure and a custom general paywall renderer. | App-user mapping, server entitlement reconciliation, webhook verification/dedupe, refunds/expiry, free-session and paid-duration ledger, localized terms, sandbox tests. RevenueCat access status does not replace our AI-usage budget. |
| **A10 — Privacy and shared patterns** | **Build domain lineage and cleanup using provider deletion APIs and pg-boss.** **Transcend** is a later privacy-operations platform. | Provider-specific account systems and, if later purchased, some privacy-request intake/routing. | Immediate revocation, derived-record ancestry, cancellation, export, resumable deletion, backup handling, cross-user eligibility and reviewed generic patterns. A privacy platform cannot infer which learning summary came from a restricted Calendar event. Users still do not publish activities. |
| **A11 — Operations and release** | **Reuse Supabase/Render/RevenueCat dashboards, Sentry, EAS and CI tooling.** A small domain operations page is enough; **Retool** can replace its UI if useful. | A monitoring backend, crash reporter, subscription dashboard, mobile release service and general internal-tool builder. | Learning run/reason view, pattern review actions, restricted admin API, audits, kill switches, end-to-end validation, migration/rollback and backup-restore evidence. Retool must call scoped admin endpoints, not receive unrestricted production database access. |

Keep these packages in the dependency map: integration, configuration and acceptance work remain even when a vendor implements the underlying capability. Do not add a coding agent or microservice for each vendor.

## 2. The four choices that need judgment

### Workflow execution: retain pg-boss

The plan already reuses a durable Postgres job system. Implement the learning state machine; do not implement a general workflow platform.

| Option | What it supplies | Consequence for this app | Decision |
|---|---|---|---|
| **pg-boss** | Postgres jobs, delayed execution, retries and recovery | Fits one API, one worker and one database; same-database enqueue can be atomic with domain writes, subject to the pinned API proof in F00 | **Keep for MVP** |
| **Trigger.dev** | Hosted task execution, retries, queues, durable waits, execution visibility | Removes some worker hosting/operations; introduces remote execution/state and a database-to-service delivery boundary | Reconsider if worker operations become a material burden |
| **Inngest** | Managed durable orchestration, memoized steps, events, waits and flow controls | Execution still runs on your compute; remote step state adds another data processor and coordination boundary | Alternative if event orchestration needs justify it |

Moving to either managed service requires a transactional outbox/reconciler for committed database work, idempotent external effects, data minimization and cancellation tests. Their idempotency features do not replace the app's evidence/XP/usage constraints. Trigger.dev documents that waits shorter than 60 seconds do not checkpoint; do not estimate all waiting as free. No Temporal, Kafka, Redis cluster, n8n instance or visual workflow builder is required for the current scope.

### Connectors: Nango is the first candidate to prove

This is the largest remaining opportunity to remove infrastructure work. Nango fits the longer-term Connect apps direction. Nylas deserves preference if the roadmap becomes primarily Google/Microsoft/iCloud calendar support rather than diverse sources. Select one integration service, never both for the same calendar account.

However, the reviewed Nango `calendar-events` template:

- Supports `calendarsToSync` and defaults to the primary calendar.
- Saves full event objects, including fields our application does not need. Minimize upstream responses, provider-side stored records and logs; trimming only our own database copy is insufficient.
- Uses an `updated_after` timestamp checkpoint and `updatedMin`, rather than the specification's Google `syncToken` algorithm. Do not paste it in and claim the current sync contract is implemented unchanged.
- Does not eliminate watch registration/renewal/stop, rolling-horizon refresh or resource selection. Nango's webhook guide explicitly requires channel renewal and a further API read after a notification. For non-primary calendars, its forwarding configuration also needs exact resource-URI mappings; unmatched notifications can lack a connection ID.

**Gate C1, owned by F00/A05:** with a consenting test account, demonstrate native authorization return, server-bound owner mapping, selected secondary calendar, create/edit/cancel, recurring-instance change, advance of the seven-day horizon, missed-notification recovery, revoked credentials and disconnect cleanup. Inspect exactly which fields are fetched, stored and logged. Validate incoming Nango webhook authenticity and bind it to a known active connection/resource; an unbound event never authorizes a user operation. Test duplicated, unmatched and out-of-order notifications. Record Google production verification requirements and Nango cost/data-processing terms.

Use a short, bounded prototype; this is not a request to maintain two production integrations. If credentials are unavailable, prepare the tests and record the gate as unrun while independent packages continue.

**Gate output:** a single decision record before production A05 implementation:

1. **Nango auth/proxy only:** Nango owns credentials; our existing Calendar sync/watch jobs own synchronization. This is the smallest adoption boundary and the first mode to try.
2. **Nango auth plus customized managed sync:** adopt only if it reduces total work and passes the same behavior/data tests. Name exactly which sync/watch jobs move to Nango and remove their local duplicates.
3. **Direct Google fallback:** keep the current specification if Nango does not justify its cost, data exposure or integration complexity. Reuse the official client/auth library; do not implement OAuth cryptography or token transport from scratch.

If Nango passes, revise the connection schema, callback/webhook routes, credentials configuration, deletion behavior, jobs and diagram together. Remove app-stored Google refresh tokens when Nango owns them. Specify the one-time mobile connection flow and backend verification; a browser success callback alone cannot attach an account. Keep provider credentials inaccessible to the model and mobile client. Production Google approval and access policies still apply.

MCP is an interface option for supported tools. It is not a substitute for OAuth ownership, event delivery or scheduled workflows. No connector platform unlocks arbitrary listening activity from other iOS apps or guarantees Spotify/food-app access.

### Voice: keep the existing direct Realtime boundary

LiveKit supplies WebRTC infrastructure, agent runtimes, turn handling and managed deployment. It is worth evaluating if F00 exposes an unacceptable integration burden or a real requirement for multiple speech/model providers. It also introduces rooms/tokens, an agent runtime, another processor and another usage bill.

The existing design already removes our media relay: phone-to-provider WebRTC carries audio; backend sideband owns tools and control. Retain it if the physical-device proof succeeds. An alternative must demonstrate interruption, audio routing, app backgrounding, reconnect, server cutoff and cost before replacing it. This review did not validate a LiveKit React Native build or compare voice quality in live sessions.

### Location: native geofencing fits the committed scope

Radar supplies geofence enter/exit/dwell events, SDKs and webhooks, but its documented geofence events are generated server-side from location data. Adopting it changes the data flow; it is not simply a nicer local SDK. Ten curated native regions do not currently justify that additional service. Revisit for broader venue discovery, richer place intelligence or measured operational needs. Preserve separate consent and actual OS delivery limits whichever platform is selected.

## 3. What we should actually build

The distinctive system is the loop **authorized context → useful opportunity → accepted practice → demonstrated evidence → adapted review**. Our code must own these decisions and the shared learning patterns that emerge from eligible experiences.

Custom work falls into two categories:

- **Product logic:** selecting useful moments, matching curriculum and ability, helpful practice, evidence and review, personal opportunity pools, and eligible shared patterns.
- **Application controls:** owner isolation, consent/resource checks, transactional allowance, cancellation, lineage and deletion. These are required integrations around reused infrastructure; they are not a justification for building general platforms.

Use one backend codebase and one database. Keep the four AI roles as bounded functions, not four autonomous deployments. Use reviewed content files instead of a CMS, SQL for the pilot's product metrics, and vendor dashboards before building custom dashboards. Configure Supabase's production sign-in/email delivery and rate limits; a hosted auth SDK still requires deployment configuration. Sentry replay is off by default; scrub private facts and transcripts from error reports. HappyRobot remains a design reference; this review has not established that purchasing its enterprise platform would replace our consumer learning stack.

## 4. Cost implications checked in public pricing

These are listed prices retrieved September 19, 2026, not quotes, credits or a full operating budget. Recheck before provisioning. No unsupported percentage of engineering effort saved is assumed.

| Product | Observed pricing / cost driver | Implication |
|---|---|---|
| **Nango** | Free: 10 connections, 10 compute hours/month, 10 GB/month, hard-capped. Pay-as-you-go: **$50/month with $50 credits**; **$0.29/connection/month**, **$0.72/compute hour**, **$0.50/GB** at listed rates | Fifty learners with one connection each represent $14.50 in connection usage but exceed the free connection cap; the paid plan minimum still applies. 1,000 learners with two connections each represent $580 in connection usage before compute/data or negotiated discounts. Do not buy the $450/month Growth add-on for this pilot. |
| **RevenueCat** | Free up to **$2,500 monthly tracked revenue**; then advertised **1% of tracked revenue** | This is not 1% only of revenue above the threshold, and it does not include or replace store fees. Our usage accounting remains necessary. |
| **Trigger.dev**, if substituted | Free includes $5 monthly credits; Hobby **$10/month** with $10 credits; Pro **$50/month** with $50 credits; execution/run charges apply | Compare its actual task pattern and residual API hosting with the existing worker cost. Free credits do not make the rest of the application free. |
| **Existing stack** | Supabase, Render, EAS, OpenAI and Sentry have separate usage/plan costs | This review did not refresh every existing vendor's pricing. Keep the previous budget as provisional; measure voice/model cost and concurrency before setting the paid allowance. |

Do not increase vendor count because credits are available. A small recurring bill can be worthwhile if it removes real maintenance; a free tier can be unsuitable if it stops autonomous work at its cap. Store permanent vendor IDs separately from our learner/session IDs so basic account data remains portable, without building an unused multi-provider abstraction.

## 5. Changes to the build assignments

| When | Change |
|---|---|
| **F00** | Pin EAS/native build setup and Paper compatibility. Record C1 as passed, failed or unrun; publish the Calendar decision before its production implementation. Keep the pg-boss transaction proof. |
| **A01** | Explicitly integrate Supabase Auth; do not create sign-in infrastructure. Retain authorization, consent and allowance tests. |
| **A02/A03** | Keep content in reviewed versioned files; compose Paper components into the learning UX. Avoid a CMS or generic UI library project. |
| **A04** | Use pg-boss scheduling/retries. Build only application states/policies and external-delivery bookkeeping. |
| **A05** | Follow the chosen connector ownership boundary; retire superseded credential/sync work if Nango is adopted. Preserve all source-behavior acceptance tests. |
| **A06/A07/A08** | Reuse the selected model/voice/native SDKs; retain learning, server-control and real-device evidence. |
| **A09** | Add RevenueCat Paywalls to the existing RevenueCat purchase integration; test restore, accessibility, localized terms and free-allowance explanations. |
| **A10** | Use provider cleanup APIs; keep lineage and cross-user eligibility local and testable. |
| **A11** | Use existing service dashboards and Sentry; implement only the missing learning-specific views/actions. EAS handles build/submission mechanics when authorized. |

Every implementation handoff should name **what was reused, what was configured, and what application logic was added**, followed by actual validation evidence. Do not mark a task complete merely because an SDK was installed or a provider advertises the feature.

## 6. Sources and verification limits

Public first-party documentation and a public integration template were reviewed; no vendor demo, commercial negotiation, security audit or live product comparison was performed. Statements about suitability are engineering judgments against this app's scope. Versions, plan limits and permissions remain implementation gates.

| Source | Evidence used |
|---|---|
| [Supabase Auth](https://supabase.com/docs/guides/auth) · [snapshot](./research/B01-supabase-auth.md) | Hosted identity methods, SDKs, JWT/RLS integration |
| [Expo EAS](https://docs.expo.dev/eas/) · [snapshot](./research/B02-expo-eas.md) | Build, Submit, Update, Workflows |
| [React Native Paper](https://callstack.github.io/react-native-paper/) · [snapshot](./research/B03-paper.md) | Components, themes and accessibility support |
| [Trigger.dev execution](https://trigger.dev/docs/how-it-works) · [snapshot](./research/B04-trigger.md); [pricing](https://trigger.dev/pricing) · [snapshot](./research/B05-trigger-price.md) | Hosted execution, retry/checkpoint behavior, pricing |
| [Inngest execution](https://www.inngest.com/docs/learn/how-functions-are-executed) · [snapshot](./research/B06-inngest.md) | Steps, managed orchestration and execution boundary |
| [Nango Calendar](https://nango.dev/docs/api-integrations/google-calendar) · [snapshot](./research/B07-nango-calendar.md) | OAuth, proxy, actions, templates and production credentials |
| [Nango Calendar template](https://github.com/NangoHQ/integration-templates/blob/main/integrations/google-calendar/syncs/calendar-events.ts) · [snapshot](./research/B08-nango-sync.md) | Actual fields, selected calendars and checkpoint algorithm; moving main-branch snapshot, pin before adoption |
| [Nango pricing](https://nango.dev/pricing) · [snapshot](./research/B09-nango-price.md); [Calendar webhooks](https://nango.dev/docs/api-integrations/google-calendar/webhooks) · [snapshot](./research/B10-nango-webhook.md) | Usage costs, channel renewal/stop and forwarding behavior |
| [Nylas Calendar](https://developer.nylas.com/docs/v3/calendar/) · [snapshot](./research/B11-nylas.md) | Calendar abstraction and notification scope; all-calendar notifications need configuration |
| [LiveKit Agents](https://docs.livekit.io/agents/) · [snapshot](./research/B12-livekit.md) | Agent runtime, media and cloud deployment |
| [Radar geofences](https://docs.radar.com/geofencing/geofences) · [snapshot](./research/B13-radar.md) | Server-generated geofence events and location handling |
| [Learnosity authoring](https://learnosity.com/build/author/) · [snapshot](./research/B14-learnosity.md) | Authoring/review toolkit; no independent outcome validation |
| [Transcend account deletion](https://docs.transcend.io/docs/articles/privacy-requirements/apple-account-deletion) · [snapshot](./research/B15-transcend.md) | Privacy request workflow and configured integrations |
| [Retool REST integration](https://docs.retool.com/data-sources/guides/connect/rest) · [snapshot](./research/B16-retool.md) | UI over REST/OpenAPI resources |
| [Sentry React Native](https://docs.sentry.io/platforms/react-native/) · [snapshot](./research/B17-sentry.md) | Error reporting, traces and optional replay |
| [RevenueCat Paywalls](https://www.revenuecat.com/docs/tools/paywalls/displaying-paywalls) · [snapshot](./research/B18-revenuecat-ui.md); [pricing](https://www.revenuecat.com/pricing/) · [snapshot](./research/B19-revenuecat-price.md) | React Native paywall integration and current listed fee |

Existing evidence for pg-boss, Realtime, Google sync and iOS limitations remains linked in the [technical specification](./technical-design-plan.md) and [connector research](./connector-orchestration-research.md).
