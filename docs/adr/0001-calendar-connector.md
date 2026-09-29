# ADR 0001 — Calendar connector gate C1

Status: **Google C1 unrun; iPhone Calendar foreground availability sync implemented, device proof unrun**  
Date: 2026-09-21

The first iOS release must let a learner connect both Google Calendar and calendars on their iPhone, with separate opt-in and disconnect controls. The iPhone source may include a Google calendar already connected directly; normalize overlapping busy intervals before proposing an opportunity so one real event cannot create duplicate invitations. iOS device permission, account-bound selection, and foreground-only, timestamp-only availability ingestion are implemented. Closed-app event changes, real-device behavior, and cross-source deduplication are not proven.

No Nango project, Google OAuth client, consenting test account, registered native redirect, or provider approval was available. C1 therefore cannot honestly select Nango auth/proxy, Nango managed sync, or direct Google for the separate cloud connection.

The application contains provider-neutral owner/resource/event lineage, a normalized signal boundary, and an absolute Temporal evaluation time for future context. It contains neither app-stored Google tokens nor a Nango production implementation, so there are no parallel connector paths to retire.

Before Google production work, run the exact C1 matrix from `docs/build-vs-buy-review.md`: native return, owner binding, selected secondary calendar, create/edit/cancel, recurrence exception, seven-day horizon, missed notification recovery, revoked credential, disconnect cleanup, fetched/stored/logged fields, webhook authenticity, duplicated/unmatched/out-of-order delivery, Google verification, and Nango processing/cost terms. Publish one replacement Google decision here and implement only that cloud path. The iPhone source still needs dependable closed-app updates or a stricter stale-window delivery gate, cross-source deduplication, and physical-device proof.
