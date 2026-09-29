# Release evidence — 2026-09-21

## Simulator iPhone Calendar foreground reconciliation — 2026-09-29

- The already-installed simulator app showed three selected iPhone calendars; the local database held iPhone-Calendar context events and offers with dispatched observe/cancel work. After one controlled app reopen, the previously active availability event was canceled while device permission and consent stayed enabled, consistent with an empty replacement free-window snapshot. The app returned to Today. This is local simulator/Compose observation, not proof of a specific real event, a physical-iPhone background sync, Google OAuth, or live notification delivery. The development client required Metro to relaunch; it was restored and left running. **Do not publish to live users.**

## Live-voice assessment before session finish — 2026-09-29

- Ending a live-voice session now stops the microphone and waits for already-transcribed learner turns to finish server assessment instead of aborting them. A failed pending assessment blocks “Continue learning”; explicitly ending an unfinished session can still abandon it. Requests have a 35-second client bound. The focused ordering test, all 31 mobile tests, mobile typechecking, and iOS bundle export passed. This does not prove physical-iPhone audio quality, provider interruption handling, or server-side Realtime call termination. **Do not publish to live users.**

## Google Calendar refresh/sign-out ordering — 2026-09-29

- Calendar status reads now share the connector's operation queue, so a token refresh finishes before sign-out clears and revokes the resulting credential. A regression test held refresh pending while sign-out began and verified no connection or revocation token remained. All 30 mobile tests and mobile typechecking passed. This is mocked concurrency proof; live Google OAuth/revocation and physical-iPhone Calendar behavior remain unverified. **Do not publish to live users.**

## Account-switch response isolation — 2026-09-28

- The shared mobile API now rejects a response if the signed-in token changed while the request was in flight. This applies to account-specific reads and writes, including Calendar calls, so an old account's response cannot be returned to the new account's screen. A regression test switches accounts before resolving a pending response; all 29 mobile tests and mobile TypeScript checking passed. This is local test evidence, not real-provider identity or device proof. **Do not publish to live users.**

## Live AI provider control probe — 2026-09-28

- Newly configured OpenAI and ElevenLabs credentials passed read-only provider checks. An explicitly gated test then used synthetic Spanish content through the app's existing adapters: OpenAI returned a valid, moderated structured tutor assessment; ElevenLabs generated a short audio clip and transcribed it back to “Hola.” No learner data or recording was sent, and no audio artifact was retained.
- `RUN_LIVE_AI_TEST=1 go test ./internal/learning -run '^TestLiveProviders$' -count=1 -v` passed, as did the ordinary Go suite and vet. This is real-provider API proof, **not** physical-iPhone microphone/speaker quality, interruption handling, multi-subject educational evaluation, or production readiness. The probe may incur provider charges and is skipped by default. **Do not publish to live users.**

## Live voice practice wiring — 2026-09-28

- Voice mode now uses OpenAI Realtime over WebRTC with server VAD: it opens the microphone, starts the coach, and streams responses without a record/review/Send loop. Learning issues a short-lived client secret only for an authenticated owner's active voice session; the long-lived OpenAI key stays server-side. Completed learner transcriptions still go through the existing server assessment before practice can advance.
- Go build and mobile TypeScript checking passed; local Learning and Experience containers were rebuilt. The iPhone 18 Pro simulator connected to OpenAI Realtime and displayed the coach's opening response, proving the secret endpoint and session negotiation. I did not acoustically verify the speaker. The simulator microphone immediately captured ambient speech, so I ended the session; use a quiet room and headphones for the demo. Dropped-network recovery, physical-device routing, server-side call termination, and authoritative transcript reconciliation remain unverified. The mobile currently relies on its authenticated client to close the WebRTC peer; this is a hackathon demo path, not a production Realtime control plane. **Do not publish to live users.**

## Google authorization abort cleanup — 2026-09-28

- A Google grant obtained while the learner switches accounts, before local connection state fails to save, or without a renewable token is now queued for revocation before it can be treated as connected. Failed revocations remain in secure storage for retry. This closes local credential-orphan paths; it does not prove Google's live revocation response.
- Eleven focused Google tests, all 23 mobile tests, mobile typechecking, and iOS bundle export passed. OAuth and revocation were mocked. A real Google test account and physical iPhone are still required. **Do not publish to live users.**

## Google Calendar reconnection after failed revocation — 2026-09-28

- Disconnect still clears the local connection and securely queues a failed Google token revocation. A new Google connection now waits for that revocation to succeed, so a reconnect cannot silently discard the old grant.
- The focused eight-test Google Calendar suite and mobile TypeScript check passed. Revocation used a mocked Google endpoint; no real account or physical iPhone was tested. **Do not publish to live users.**

## Concurrent free-cap proof — 2026-09-27

- The active Go Learning service's account-level transaction lock was tested at the 25-session boundary. With 24 existing sessions, 30 simultaneous, distinct starts accepted exactly one request, rejected 29 with the free-cap response, and left 25 counted sessions. No model call or purchase was made.
- The focused test and full Go suite passed against freshly migrated disposable PostgreSQL databases; `go vet ./...` passed and both test databases were removed. This proves a local cost-control invariant, not live provider spend or paid entitlement behavior. **Do not publish to live users.**

## Simulator Calendar-screen inspection — 2026-09-27

- Xcode Device Hub exposed the booted iPhone 18 Pro app to accessibility inspection. The optional Connections screen showed separate iPhone and Google Calendar choices. Opening Google Calendar displayed the explanation and the explicit message that an iOS Google client ID is required in a new build; no Connect action was available. No source was selected, no new permission was granted, and no availability was sent by this inspection.
- This corrects the earlier limitation that simulator controls were unavailable at the time of the initial-load check. The bundle identifier remains a placeholder and no Google iOS client ID is configured. Actual Google authorization, iPhone Calendar selection/ingestion, and physical-device proof remain open. **Do not publish to live users.**

## Local simulator launch against rebuilt services — 2026-09-27

- Docker Desktop restarted successfully. Compose rebuilt and started all four Go HTTP services, the Temporal worker, PostgreSQL, and Temporal; Experience readiness returned 200. Development-token requests for bootstrap, subjects, and opportunities returned 200.
- A fresh native iPhone 18 Pro simulator build installed and launched without Xcode errors. Its Today screen showed one learning path, matching the local API's track count. This proves local app launch and initial data loading, not a full tap-through: simulator UI controls were unavailable to automation. No Calendar permission, Google OAuth, real voice, physical-device delivery, or push was exercised. **Do not publish to live users.**

## Last-device location revocation — 2026-09-27

- Removing an account's last location-enabled device now atomically revokes location consent, disables its saved place, cancels active geofence events, and queues workflow/offer cancellation. Removing one of two enabled devices leaves the shared place active. This closes a server-side removal gap; it does not prove that an offline phone stopped monitoring or that a physical device delivered a background event.
- A migrated disposable PostgreSQL test covered both cases. The full database-backed Go suite and vet passed; both temporary databases were removed. Docker remained unavailable, and no provider, push, or physical-iPhone test ran. **Do not publish to live users.**

## Google Calendar revocation retry — 2026-09-27

- Sign-out and account switching now remove the active Google connection while retaining a refresh token only in secure revocation state if Google is unavailable. The app retries on opening/foregrounding, and blocks a new Google connection until older revocations succeed. This prevents an interrupted revocation from being silently forgotten; it does not fix offline server-device removal during sign-out.
- Eight focused Google Calendar tests, all 20 mobile tests, mobile typechecking, the Go suite, and iOS bundle export passed. Revocation retries used a mocked Google endpoint; no real account, OAuth grant, or physical iPhone was tested. **Do not publish to live users.**

## iPhone Calendar availability reaches the opportunity pipeline — 2026-09-26

- Selected iPhone calendars are read on app activation. The phone merges busy periods and sends only bounded free-window timestamps, never titles, guests, notes, or event IDs. The new owner/device/consent-bound endpoint reconciles edited windows without duplicate active events and queues cancellations when a window disappears. Disconnect, device removal, and observed iOS permission loss remove the source's windows; the last connected iPhone also revokes its consent. Existing Temporal evaluation recognizes the iPhone source, but automatic push is deliberately suppressed until closed-app freshness can be guaranteed.
- A freshly migrated disposable PostgreSQL database passed cross-account, idempotent refresh, changed-window, disconnect, device-revocation, consent-revocation, and no-unattended-push tests, then was removed. The full Go suite and vet, 14 Vitest tests (8 unrelated integrations skipped), mobile typecheck, and iOS bundle export passed. This is foreground refresh only: changes while the app stays closed may leave a stale in-app offer until expiry or next successful sync. No physical-iPhone EventKit run, APNs delivery, cross-source Google deduplication, or Google connection was verified. **Do not publish to live users.**

## Separate iPhone and Google Calendar connections — 2026-09-26

- The first-release requirement is now both Google Calendar and calendars available on the iPhone, each separately optional. The iOS Connections screen requests full Calendar read access, lists event calendars by account, and saves only selected calendar IDs bound to the signed-in account. It clears the selection on account change, sign-out, or observed OS permission revocation. Google is still labeled unavailable. No event details are stored or sent by this device-selection slice.
- The focused account/permission test, mobile TypeScript check, iOS bundle export, CocoaPods installation, and native iPhone 18 Pro simulator build passed; the built app contains both Calendar usage-description keys. The full Vitest run passed 11 tests and skipped 8 database integrations without a database. The simulator was not launched because local Docker services were unavailable. This is **not** Calendar availability ingestion, event-change handling, duplicate-source suppression, Google OAuth, live invitation delivery, or physical-iPhone proof. **Do not publish to live users.**

## Application evidence no longer trusts context labels — 2026-09-26

- The Experience service now sends only the owned offer and chosen outcome to Learning. Learning resolves the offered activity from its reviewed template or course manifest and verifies that the offer, context, consent, track, and activity are still current before recording self-reported evidence. It no longer takes a caller-supplied context category as permission to treat a course node as an application activity.
- A migrated PostgreSQL test rejects a forged `contextCategory` field while preserving saved-place, Calendar-availability, and reviewed Calendar-template flows. Full Go tests and vet passed; the disposable database was removed. Local Learning and Experience services were rebuilt, and public readiness returned 200. This hardens the existing internal boundary; it is not a live Calendar connector or release proof.

## Assessed opportunity practice advances learning — 2026-09-26

- Finishing an assessed course activity from a context opportunity now marks that activity complete, just as finishing it from the course does. A reviewed application template remains separate from course progress. The Experience API closes a successfully completed invitation through the existing Opportunity service; an interrupted handoff can be retried through the idempotent session finish. The iOS practice screen refreshes invitations, course, and progress before returning.
- A migrated PostgreSQL integration test completed a formerly missed opportunity, verified one course-progress credit after a repeated finish, and confirmed the next opportunity moved to the next activity. It also selected a reviewed Calendar-context template, completed its practice session, and verified that a context-only activity did not inflate course progress. A focused service test covered invitation completion, retry, canceled-offer conflict, and abandoned/curriculum sessions. Full Go tests and vet plus mobile typechecking passed; the disposable test database was removed. Local Learning and Experience services were rebuilt and public readiness returned 200. Real provider, push, and physical-device behavior remain unverified. **Do not publish to live users.**

## Calendar-availability learning selection — 2026-09-25

- A normalized `calendar_availability` context can now select the next activity or due review across active learning paths, including clearly labeled AI-generated courses. Other Calendar event categories still require a reviewed context-specific application template; the service will not invent a claim that an arbitrary event matches a skill. The offer retains a distinct reason, and iOS explains the free-window suggestion in plain language.
- Migrated PostgreSQL tests covered due-review priority, a later curriculum change, application self-report, unmatched-event suppression, and preservation of the Calendar reason through offer creation. Full Go tests and vet plus mobile TypeScript checking passed; the disposable test database was removed. The local Learning service and opportunity worker were rebuilt, and public readiness returned 200. No Calendar account was connected, no real event was ingested, and no push or physical-device interaction was tested. This is a candidate-selection contract, **not** a working Calendar connector or live-user release proof.

## Temporary voice-file cleanup — 2026-09-25

- iOS voice practice now deletes a captured recording even when reading it fails, retries stopping a recorder after a stop error, clears its recording timer on screen exit, and removes a coach-audio file if the screen closes while it is being written. The microphone mode is reset after failed recording setup or transcription.
- A focused test proved recording-file deletion after both a successful read and a read error; the mobile typecheck and iOS export passed. This is privacy/error-path hardening, **not** live ElevenLabs or physical-device voice-quality evidence. Those checks remain release gates.

## Saved-place retry without Learning-service dependency — 2026-09-25

- Foreground retry now submits each persisted saved-place removal with its original account ID directly to the authenticated Experience API. A 409 for another signed-in account preserves that account's pending removal and allows other queued location work to proceed. The server's owner check remains the authorization boundary. Retries no longer depend on the Learning service answering the composed bootstrap request.
- The mobile test exercised offline removal, cross-account rejection, same-account retry, permission revocation, and re-enrollment ordering. All three focused tests, mobile typecheck, and iOS export passed. A real phone, real sign-in provider, and offline/reconnect transition remain unverified. **Do not publish to live users.**

## iOS permission-revocation reconciliation — 2026-09-25

- On authenticated foreground, the app now checks whether a saved practice place still has iOS background-location permission. If the learner revoked that permission in Settings, it stops local monitoring, queues account-bound server removal, and retries that removal when online. This uses the existing disconnect path, so it cannot silently keep consented place opportunities active after an observed OS-level revocation.
- A focused mobile test covered OS permission loss with a failed server request and a persisted retry. The three geofence tests, mobile typecheck, and iOS bundle export passed. The simulator launched, but this did not exercise the iOS permission dialog or a physical background event. **Do not publish to live users.**

## Offline saved-place disconnect retry — 2026-09-25

- iOS now stores an account-bound removal intent before stopping a saved-place geofence. If the authenticated server deletion fails, the app retries on the next authenticated foreground session. A different signed-in account cannot consume the intent; the public API also rejects a deletion whose expected owner differs from the authenticated account. Re-enrollment for the same account first resolves any pending removal, preventing a delayed retry from deleting a newly saved place.
- Two focused mobile tests covered failed deletion, local monitoring shutdown, cross-account non-delivery, same-account retry, and blocking re-enrollment until a pending removal succeeds. A Go test covered the server-side account mismatch. Full Go tests and vet, mobile typechecking, and iOS export passed. The rebuilt local Experience service returned readiness 200, the authenticated account ID, and 409 for a mismatched owner. These are local/mock checks; no physical-iPhone offline/reconnect test was run. Existing saved places created before this owner marker may require an online reconnect to establish the account identity. **Do not publish to live users.**

## Atomic saved-place disconnect — 2026-09-25

- Disconnecting a saved place now makes one authenticated server request. The Context service takes the learner profile lock, records a location-consent revocation if still granted, disables the saved region, cancels every active location event, and queues offer/workflow cancellation in one PostgreSQL transaction. A repeat disconnect does not add another consent decision. The phone stops local geofence monitoring even if the request fails.
- A freshly migrated disposable PostgreSQL test covered active-event cancellation even when the region had already been disabled, idempotent revocation, and rejection of re-enrollment without new consent. Full database-backed Go tests, vet, mobile TypeScript checking, and iOS export passed; the test database was removed. The local Context service was rebuilt and API readiness returned 200. An offline disconnect still cannot reach the server until a retry, and physical-iPhone behavior remains unverified; **do not publish to live users**.

## Temporal integration in CI — 2026-09-25

- CI now starts a disposable Temporal server alongside PostgreSQL before `pnpm test`, so the Context-dispatch cancellation test runs on every push and pull request instead of silently skipping without `TEST_TEMPORAL_ADDRESS`.
- The workflow YAML parsed locally. A fresh migrated disposable PostgreSQL database, the running local Temporal server, the full `pnpm test` command, and a verbose targeted cancellation run all passed. A second disposable Temporal container using the CI host-gateway connection pattern also started, accepted the targeted cancellation test, and was removed with its three test databases. The GitHub-hosted runner itself has **not** run until these uncommitted files enter CI. This improves regression coverage, not live-provider/device readiness.

## Editable quiet hours and time-zone reconciliation — 2026-09-25

- Settings now lets the learner edit or disable quiet hours. On authenticated app opening/foregrounding, iOS compares the device IANA time zone with the account and updates it when different. A preference change atomically wakes deferred notification work so the existing dispatcher rechecks consent, quiet hours, expiry, and limits under the new schedule.
- A migrated disposable PostgreSQL test rejected an invalid time, saved valid hours, and verified a deferred invitation became due for re-evaluation. Full database-backed Go tests and vet, mobile TypeScript checking, and iOS export passed; the test database was removed. The local Experience service was rebuilt and readiness returned 200. The simulator visibly loaded Settings with an actionable Quiet hours row after relaunch, but the Mac was locked, so editing/tapping and a physical-device time-zone change were not manually verified. No live push was sent. **Do not publish to live users.**

## Future-context Temporal timing — 2026-09-25

- A local integration test now sends a future Calendar-shaped context event through the real Context dispatcher and Temporal server, confirms the workflow actually entered its one-hour timer, then dispatches cancellation. Temporal reached `CANCELED` and the offer-cancel endpoint was called once. The full database-backed Go suite and vet passed. The disposable migrated database was removed; no real Calendar account, push provider, or live user was involved. This closes the dispatcher-to-Temporal cancellation proof gap, not the connector or device gates.
- Migration 015 adds an optional absolute context evaluation time, constrained to precede expiry. The Context outbox forwards it to the existing Temporal workflow, which waits on Temporal's deterministic clock; location events without one retain their short relative delay. This is the scheduling primitive needed for a Calendar event found days ahead, not a Calendar connector or permission grant.
- A Temporal test fast-forwarded 48 hours and verified evaluation did not run at the two-second fallback; a migrated PostgreSQL test accepted a future context and rejected evaluation after expiry. The full database-backed Go suite and vet passed, and the disposable database was removed. Migration 015 applied to local Compose, the Context service and worker were rebuilt, all services are running, and public readiness returned 200. Real Calendar account connection, event edit/cancel rescheduling, and physical notification delivery remain unverified. **Do not publish to live users.**
- The Context dispatcher now locks each context event while claiming its outbox entry, skips canceled or expired observations, and asks Temporal to cancel a still-waiting workflow after closing its offer. A migrated PostgreSQL test verified canceled/expired observations were marked handled without starting Temporal; a virtual-time test canceled a 48-hour workflow before evaluation. The full database-backed Go suite and vet passed; the disposable database was removed. The rebuilt local Context service and public readiness passed. Real provider edit/cancel and notification recall remain unverified. **Do not publish to live users.**

## Honest skill evidence on Progress — 2026-09-25

- A missed opportunity is no longer counted as a practiced skill. Progress now counts distinct skills with real practice or attempted/applied self-report, and separately labels applied skills as **self-reported** instead of showing a permanently zero “Demonstrated” count. A missed moment does not advance the course or create a skill-review state.
- A migrated disposable PostgreSQL test covered two practiced skills, one applied self-report, and a missed third skill without progress inflation; the full Go suite and vet passed. The disposable database was removed. Mobile TypeScript checking and iOS bundle export passed. The local Learning service was rebuilt; public readiness returned 200 and the authenticated Progress response exposes the new field. This is evidence semantics, not verified physical-skill performance. **Do not publish to live users.**

## Moderated practice turns — 2026-09-25

- Practice now requests OpenAI input and output moderation through the [Responses API moderation contract](https://developers.openai.com/api/reference/resources/responses/methods/create), as course generation already did. Missing, errored, or flagged moderation results fail closed before a tutor turn or understanding assessment is saved. The learner receives the existing retryable coach-unavailable message; a flagged response cannot advance a course.
- A provider-double test checked the moderation request and rejection of flagged tutor output. The full Go suite against a freshly migrated disposable PostgreSQL database and `go vet ./...` passed; the disposable database was removed. The local Learning service was rebuilt, and public readiness and authenticated bootstrap returned 200. This is **not** live OpenAI proof or a multi-subject quality/safety evaluation. **Do not publish to live users.**

## Local service exposure — 2026-09-25

- Local Compose binds the public API, PostgreSQL, and Temporal ports to `127.0.0.1`. The shared development internal token is rejected outside `APP_ENV=local`; Compose sets local mode for the affected services. Full Go tests and vet passed, containers rebuilt, and authenticated bootstrap and readiness returned 200. These are local safeguards, not production identity or network-segmentation evidence. **Do not publish to live users.**

## Saved-place relevance — 2026-09-25

- A learner now chooses a learning path when saving an iOS practice place. The server verifies that the path is active and owned by the learner, snapshots the path on each location event, and passes it through the Temporal evaluation to candidate selection. A due review in an unrelated path can no longer hijack a selected place.
- Migration 014 retires old generic saved places and cancels their active context events; coordinates remain so a learner can reconnect the place to a path. An upgrade rehearsal with a legacy place produced `enabled=false`, `canceled=true`, and one cancellation outbox entry. The current local account had zero such places to retire.
- Freshly migrated PostgreSQL tests covered owner isolation, path binding through the Context API, selected-path candidate filtering, and workflow forwarding. Full Go tests and vet, mobile typechecking, and iOS bundle export passed. The affected local services were rebuilt, migration 014 applied, and API readiness passed. The simulator opened Connections, but its locked Mac prevented tapping the Location control for visual interaction proof. Physical-device geofencing remains unverified. **Do not publish to live users.**

## Source-specific consent and application boundary — 2026-09-25

- A context-consent revocation now cancels that source's active events and queues offer cancellation without canceling a separately consented source. The Location ingestion transaction rechecks device binding and consent under the profile lock, closing the stale eligibility-response race. Calendar and Location opportunity eligibility are source-specific; the offers list hides revoked, canceled, expired, or paused moments.
- The Learning service now locks the learner profile and verifies an offer's owner, activity, source consent, live context, status, and pause state before a new opportunity practice session or real-world application report. Already-recorded internal application evidence remains idempotent; a revoked source cannot create new evidence. This does not certify a physical skill from self-report.
- Full Go tests against a freshly migrated disposable PostgreSQL database and `go vet ./...` passed. The database was removed. The affected local services were rebuilt; Experience readiness and authenticated opportunity-list smoke checks passed. No live Calendar connector, physical-device proof, or external provider call was exercised. **Do not publish to live users.**

## Adaptive opportunity selection — 2026-09-25

- The Go Learning service now uses one next-activity decision for Today and saved-place opportunities. A due review outranks an unfinished activity across active learning paths; the earliest due review wins. Today and Course label the review; the offer preserves a distinct review reason, and iOS explains it without incorrectly labeling reviewed content as AI-generated.
- Assessed practice now schedules a next review after one day when assistance was used or three days after independent understanding. A new assessed practice moves the review forward from its old due time. Real-world self-reports remain distinct evidence and cannot postpone an earlier assessed review or claim demonstrated mastery.
- A migrated disposable PostgreSQL integration test covered assisted versus independent intervals, due guitar review outranking unfinished Spanish, matching Today and opportunity selection, self-report not deferring the review, and a successful review changing the next opportunity. Full database-backed Go tests, vet, mobile TypeScript checking, and iOS bundle export passed. This does **not** yet prove live model quality, physical-device delivery, or a proactive follow-up without a context event. **Do not publish to live users.**
- Simulator inspection found that the bundled, unreviewed Spanish course was incorrectly labeled “AI-generated.” A shared status label now distinguishes unreviewed bundled content from genuinely AI-generated content across onboarding, Today, Learn, Course, and learner profile. The iPhone 18 Pro simulator visibly shows “NOT EDUCATOR-REVIEWED” on Today and “Not educator-reviewed” after tapping through to Course; the label check, TypeScript checking, and iOS bundle export passed. Practice remains disabled locally without an OpenAI key.

## Consent/pause-before-offer race — 2026-09-25

- The Temporal evaluation activity now locks the learner profile and rechecks its pause state, the context event, and the latest source-specific consent immediately before creating an offer. Consent revocation and preference updates serialize on that profile row, so a stale remote eligibility result cannot create a new offer after either change commits.
- A disposable migrated PostgreSQL test changed consent or pause after the remote eligibility response but before the offer transaction. Both cases created no offer; a restored, unpaused control did create one. Full database-backed Go tests and vet passed; the disposable database was removed. The local opportunity worker was rebuilt and started; API readiness passed. The running app was not restarted. Physical-device and live-provider gates remain open. **Do not publish to live users.**

## Cancellation-before-offer race — 2026-09-25

- The Temporal evaluation activity now rechecks and locks the owner-bound context event immediately before inserting an offer. A cancellation committed after the remote read but before this transaction yields no offer; cancellation arriving after the insert waits and then queues the normal offer-cancel path.
- A disposable migrated PostgreSQL test deliberately returned a stale “active” remote event after canceling its database row, then verified no offer was created. A live control event still created one offer. Full database-backed Go tests and vet passed; the disposable database was removed. The rebuilt local Temporal worker started and public Experience readiness passed. This is local race evidence, not physical-device or live-notification proof. **Do not publish to live users.**

## Out-of-order geofence signals — 2026-09-25

- The context service now tracks the latest exit timestamp for each saved-place version. A delayed enter observed before that exit is ignored; a delayed exit cancels only enters it chronologically follows, leaving newer learning moments intact. Region validation and signal processing share a row lock, so a saved-place update cannot race a signal against an old version. Replayed enter IDs remain idempotent.
- A freshly migrated disposable PostgreSQL database exercised two enters, a late earlier exit, a later exit, a late enter, a replay, a fresh enter, place-version reset, and removal. Full database-backed Go tests and vet passed. The disposable database was removed. Migration 013 was applied to the running local Compose database; the rebuilt Context service and public Experience readiness checks passed. Physical-iPhone background delivery and notification display are still unverified; **do not publish to live users**.

## Practice progression check — 2026-09-24

- A practice turn now requests a [strict structured coach response](https://developers.openai.com/api/docs/guides/structured-outputs) with an explicit conceptual-understanding result. The server records that result per turn; a model-provided answer cannot count as learner understanding. Finishing a curriculum activity and awarding progress require at least one passing learner response. An unfinished session returns its assessment state when reopened, and iOS offers “Continue learning” only after a pass. Physical-skill performance is **not** certified from text or voice; real-world application remains separately self-reported.
- A fresh, migrated disposable PostgreSQL database verified weak answer → no progress, assisted model answer → no pass, independent answer → pass and progress, duplicate turn → no duplicate model call, and resume → assessment retained. Full database-backed Go tests, vet, mobile TypeScript checking, and iOS bundle export passed. The disposable database was removed. Migration 012 was also applied to the running local Compose database; the rebuilt Learning service and public Experience readiness checks passed.
- The evaluator has only mock-provider test evidence. A live OpenAI test and a multi-subject quality/safety evaluation remain required before release. **Do not publish to live users.**

## Account data export — 2026-09-24

- The authenticated `GET /v1/me/export` endpoint now returns a repeatable-read snapshot of profile, consent, device-permission, learning, saved-place, context, opportunity, and notification-status data. Push tokens, provider tickets, idempotency keys, and work-queue records are excluded. The iOS Settings screen warns that transcripts and saved-place coordinates are included, shares the JSON file, and removes its temporary cache copy after sharing or on next app start after an interruption.
- A freshly migrated disposable PostgreSQL database verified all 15 export sections, owner-only rows from two populated accounts, secret exclusion, and cleanup. Full Go tests and vet, mobile TypeScript checking, iOS bundle export, local API readiness, authenticated export, and unauthenticated rejection passed. The disposable database was removed.
- The iPhone 18 Pro simulator opened the native share sheet with the generated JSON file (6 KB); dismissing without sharing removed the temporary file. A physical-iPhone check remains. Account-data export is **not** account deletion: deleting the hosted sign-in identity requires a selected provider and its deletion API. Large-account export sizing and production privacy/operations checks remain open. **Do not publish to live users.**

## Receipt handling — 2026-09-24

- Migration 011 adds receipt state and due-check indexing. The opportunity worker batches due ticket IDs, checks Expo receipts after 15 minutes, records accepted/rejected/missing/failed outcomes, and disables a `DeviceNotRegistered` token only when the device and token hash still match the original send.
- A disposable migrated PostgreSQL database and mock Expo endpoints verified accepted and rejected receipts, invalid-token removal, and missing-receipt retries; `go test ./...` and `go vet ./...` pass. A real receipt has **not** been obtained. Expo states that even an `ok` receipt proves only APNs/FCM acceptance, not display on the phone: [Expo notification delivery documentation](https://docs.expo.dev/push-notifications/sending-notifications/).
- The first-release decision is still iOS-first with clearly labeled AI-generated courses. Physical-iPhone notification display and tap-through, provider credentials, and the remaining release gates below are open. **Do not publish to live users.**

## Notification path — 2026-09-24

- iOS-first and clearly labeled AI-generated courses are approved launch decisions. No educator review is required before an AI course is offered, but safety and subject-quality evaluation remain open.
- Notifications now have a separate iOS opt-in, account-bound Expo token registration, explicit revocation, and tap navigation to Moments. Location enrollment preserves a previously registered push token; initial device/consent enrollment creates the profile if necessary.
- Each Go opportunity atomically queues one notification. A worker locks the offer, context, profile, and selected device; it rechecks current consent, cancellation, expiry, pause, quiet hours, and a one-per-local-day cap, then submits generic text to Expo with bounded retries. Quiet-hour pushes wait only while the opportunity remains valid. Mock-provider and disposable-Postgres tests cover success, suppression, deferral, token ownership, and transient retry. A provider ticket is **not** a delivered notification receipt.
- `TEST_DATABASE_URL=… go test ./...`, `go vet ./...`, mobile TypeScript checking, iOS bundle export, local migration 010, Experience readiness, and notification-status smoke check pass. The simulator app launches. Local Compose keeps `PUSH_ENABLED=false`; EAS/APNs credentials and physical-iPhone delivery, tap, receipt, revocation, and background tests remain **unverified**. The disposable test database was removed after testing.
- Physical-device and large-account export checks, account deletion, Calendar, live OpenAI/ElevenLabs/OIDC, paid billing, AI-course quality evaluation, production operations, and native-device accessibility remain release gates. **Do not publish to live users.**

## Go replacement check — 2026-09-23

The Go/Temporal backend supersedes the TypeScript implementation measured below. The older rows are historical evidence, not proof that the Go system is ready for live users.

- `go test ./...` and `go vet ./...` pass; mobile TypeScript checking passes.
- Docker Compose builds and starts PostgreSQL, Temporal, the migration job, four Go HTTP services, and the opportunity worker. The app API readiness check passes.
- A local authenticated request created a learning track, started and completed a text session, and recorded separate practice/progress evidence.
- With a temporary reviewed test template and region, a consented location signal reached the context outbox, started a Temporal workflow, produced an opportunity, accepted a real-world self-report, and rejected a conflicting repeat. A location exit canceled a later offer. The temporary template, region, device, consent, track, events, sessions, and evidence were removed; quiet hours were restored.
- Real Calendar access, actual notification delivery, real voice, billing, privacy export/deletion, reviewed multi-subject content, and physical-device behavior have **not** been revalidated or completed. Do not release to live users yet.

## Current iOS-first work — 2026-09-23

- The user approved clearly labeled AI-generated courses before educator review and an iOS-first release. Android is not a first-release gate.
- A learner can now request a custom subject and goal. The Go learning service requests strict JSON-schema output and input/output moderation from OpenAI, validates a three-level course, stores it per track, limits generation attempts, and returns explicit AI provenance. Mock-provider integration tests cover creation, replay, practice, moderation rejection, and quota. No live OpenAI key or subject-quality evaluation was available.
- The iOS Connections screen now offers explicit one-place enrollment. Personal regions are account-scoped; consent and device binding gate creation and signal ingestion. The geofence task queues owner-bound events and retries uploads. Sign-out and removal clear local monitoring. A disposable Postgres test verified cross-account isolation, region versioning, event cancellation, and revocation rejection. Native background delivery and notification display are unverified.
- A saved practice-place event can select the next unfinished course activity, with a visible reason and AI-generated label. The full phone-to-offer path has not been run on a physical iPhone. Calendar, email, and messages remain unavailable rather than pretending to connect.
- `TEST_DATABASE_URL=… go test ./...`, `go vet ./...`, content validation, mobile TypeScript checking, iOS bundle export, and a native iPhone 18 Pro simulator build pass. The rebuilt app launches and shows the AI-generated-course disclosure; tap-through and permission checks were unavailable because the Mac UI was locked. Compose was rebuilt with migrations 007–009; Experience readiness and read-only track, subject, and region API smoke checks pass. Live OpenAI/ElevenLabs calls, a physical-device pass, and end-to-end notification delivery remain untested. **Do not release to live users yet.**

The contract table below originated with the earlier TypeScript architecture. Its release-critical rows are corrected for the active Go/iOS app; the command list after it is a historical 2026-09-21 snapshot, not the current test count. **Local** means executable local evidence, **prepared** means code/config exists but external proof is still required, and **blocked** names an unresolved gate.

| Contract | Status | Evidence / gate |
|---|---|---|
| R01 ordinary learning | local | API bootstrap → atomic reserve → text turn → assessment → progress path; iOS export succeeds. Content remains educator-unreviewed. |
| R02 Calendar opportunity | partial | Separate on-device iPhone and Google selection, foreground availability ingestion, cancellation, and same-phone overlap gate have local tests. Google OAuth credentials/test account, physical-iPhone proof, and dependable closed-app refresh are missing. |
| R03 location opportunity | prepared | Native top-level Expo task, owner/device/consent validation, dedupe, delayed workflow, expiry, outbox. Physical iPhone proof unrun. |
| R04 personalized live/text | partial | Text fallback and bounded OpenAI adapter implemented; model benchmark and live voice blocked by funded model access/device proof. |
| R05 durable feedback | local | Idempotent evidence/XP, `skill_state.next_review_at`, later due-review count. |
| R06 shared improvement | not evidenced | Cross-learner pattern extraction/review is not implemented in the active Go services and is not a first-release prerequisite. |
| R07 free/purchase | partial | A server-enforced free cap exists. Paid subscriptions are deferred from the first release; RevenueCat sandbox billing has not been verified. |
| R08 control/isolation | partial | Owner checks, pause/consent controls, and data export have local tests. Account deletion is not implemented; real identity-provider and privacy-operation proof remain mandatory. |
| R09 recovery/explainability | local/partial | The active Go path uses Context outbox plus Temporal wait/cancel and offer reasons, not pg-boss. Real-provider and operational recovery evidence remains open. |
| R10 mobile quality | prepared | Paper UI, 44-point actions, VoiceOver labels, Dynamic Type/system fonts, empty/error/offline cache, iOS bundle export. Physical accessibility pass unrun. |

| Test | Status | Evidence / gate |
|---|---|---|
| T01 | local/partial | Owner-bound API and database checks have local tests; active Go services do not implement the former runtime-role RLS design. |
| T02 | local | The active Go service accepted one of 30 concurrent starts from 24 existing sessions and rejected the rest at its server-enforced 25-session cap. |
| T03 | partial | Local snapshot/reconciliation and cross-source checks pass. Real Google-account create/edit/cancel and physical-device tests are unrun. |
| T04 | local/partial | Context outbox and Temporal wait/cancel have local tests; the historical pg-boss proof is not evidence for the active path. |
| T05 | local/prepared | DST/quiet/cooldown/cap/revocation policy tests; real push delivery unrun. |
| T06 | partial | Model receives bounded learning fields and no connector credentials/tools; adversarial live-model evaluation unrun. |
| T07 | blocked | A07 requires successful F00 physical Realtime control probe. |
| T08 | local | Assistance/uncertainty retained and next review changes. |
| T09 | not evidenced | The historical cross-learner pattern flow is not implemented in the active Go services. |
| T10 | blocked | RevenueCat/App Store sandbox configuration unavailable. |
| T11 | blocked | The active app has no account-deletion endpoint or UI flow. Identity-provider selection, app-data cleanup, and provider deletion proof remain open. |
| T12 | blocked | Physical iPhone, Apple team, approved regions, and provider access unavailable. |

## Historical commands run on 2026-09-21

- `pnpm content:check` — 25 nodes and 25 unique skills; prerequisite graph valid.
- `pnpm typecheck` — contracts, backend, and mobile pass.
- `pnpm test` — four policy/lineage tests pass; integration suite is explicitly skipped without `DATABASE_URL`.
- `DATABASE_URL=… pnpm test:integration` — eight Postgres/API invariant tests pass.
- `pnpm --filter @orchestrator/mobile build` — iOS JavaScript bundle exported successfully.
- A clean temporary Postgres database applied migrations 001–002 successfully and was removed afterward.
- `docker build -f infra/Dockerfile .` and the resulting container's `/health/ready` smoke check pass locally.
- Expo Doctor — 20/22 checks pass. Open findings: SDK 56's known Hermes regression and `react-native-webrtc` not marked New Architecture-tested. The plan fixes SDK 56 and direct WebRTC pending the physical proof, so these are release blockers, not hidden exclusions.
