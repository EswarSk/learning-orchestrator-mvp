# Research brief: experience-led language-learning startup

**Research date:** September 17, 2026, Pacific time.  
**Coverage:** 25 reviewed primary guidance/product/documentation sources or research-paper records; additional search results were used for discovery. Saved extracts are linked below. Official platform pages were requested with fresh fetches where supported.

## The research-backed recommendations

**Keep the ambition; narrow the deployment.** YC's MVP guidance favors a quick, bounded product that helps a small group and starts the learning process. It does not imply that this product should become a manual check-in app. Here, autonomous context detection is the hypothesis to preserve. Limit language, platform, audience, activity types, and connector count instead. [S01–S03]

**Treat the founder's first task as finding an urgent user.** Interview recent real experiences, current workarounds, spending, and actual phone/service usage. Twelve to fifteen interviews is this plan's time-box, not a scientifically sufficient market study. Recruit and build together rather than wait for perfect research. [S02–S03]

**Separate product-user fit from market scale.** A handful of enthusiastic learners can identify a useful product, but do not establish a venture-scale market or repeatable acquisition. Measure actual recurring practice and willingness to keep paying. The shared-pool network effect remains a hypothesis. [S06–S08]

**Make usefulness, not interruption volume, the retention strategy.** a16z explicitly treats notification frequency and relevance as double-edged. A language-learning field study likewise reports an exposure-versus-disturbance trade-off. The agent must learn when to remain quiet. [S07, S23]

**Avoid mistaking generated content for a differentiated company.** Speak already offers personalized conversational practice; Google has explored situation-based lessons. The distinctive bet is autonomous discovery plus longitudinal skill evidence and permitted shared improvement. The research does not establish that no other company has this combination. [S10–S11]

**Connector restrictions change the build order.** Calendar and geofencing have documented implementation routes. Neither provides perfect knowledge of activity. Spotify is not an uncomplicated foundation for an AI music workflow; food-order APIs examined serve merchants. Test real access and distribution in week one. [S12–S22]

**Separate personal context from shared knowledge.** The shared pool cannot indiscriminately ingest all connected experiences. Google's rules cover derived data and restrict AI improvement beyond the particular user. Preserve provenance and exclude restricted ancestors from pooled extraction. This is a concrete architectural constraint, not an optional privacy polish. [S15]

## What the evidence does not establish

- No interviews, product usage, paid conversions, or integration prototypes were conducted during this research.
- No verified claim that more contextual practice always leads to faster retention, or that this product is superior to existing instruction.
- No guaranteed access to Spotify at pilot scale, no verified cross-platform live playback signal, and no verified consumer food-order-history integration.
- No claim that Android notification-listener access resolves third-party content terms.
- No guarantee of background-location store approval or second-exact event delivery.
- No proven network effect, data moat, product-market fit, or accelerator acceptance odds.
- No researched market-size estimate. Subscription scenarios and budget figures in the plan are labeled assumptions.
- Paper coverage was narrower than a systematic literature review. The smartphone paper was inspected through its abstract and retrieved body passages; the variable-retrieval paper had incomplete abstract text and no retrievable full text. The plan deliberately avoids numerical efficacy claims.

## Source register

### S01. YC — How to build an MVP

[Original source](https://www.ycombinator.com/library/Io-how-to-build-an-mvp) · [Saved extract](./research/S01-yc-mvp.md)

Launch a bounded useful product quickly, give it a deadline, and iterate with a small group. Applied by shipping a real autonomous loop by week four rather than waiting for the full pilot.

### S02. YC — How to talk to users

[Original source](https://www.ycombinator.com/library/Iq-how-to-talk-to-users) · [Saved extract](./research/S02-yc-interviews.md)

Study concrete problems and existing behavior; avoid hypothetical willingness-to-use questions. Applied to the interview script and first-user selection.

### S03. YC — Essential startup advice

[Original source](https://www.ycombinator.com/library/4D-yc-s-essential-startup-advice) · [Saved extract](./research/S03-yc-advice.md)

Launch, recruit manually, talk to users, avoid premature scaling, and track a small set of meaningful metrics. Human recruitment and support fit this guidance; simulated autonomy would not validate this product.

### S04. YC — FAQ

[Original source](https://www.ycombinator.com/faq) · [Saved extract](./research/S04-yc-faq.md)

Idea-stage and solo-founder applications are permitted; incorporation is not necessary to apply; building capability and full-time commitment matter. No acceptance probability can be inferred.

### S05. YC — Apply

[Original source](https://www.ycombinator.com/apply) · [Saved extract](./research/S05-yc-apply.md)

At retrieval, Winter 2027 runs January–March in San Francisco, with November 2 at 8 p.m. Pacific as the on-time deadline. Recheck before submission.

### S06. a16z — Product-user fit before product-market fit

[Original source](https://a16z.com/product-user-fit-comes-before-product-market-fit/) · [Saved extract](./research/S06-a16z-userfit.md)

Early product-user fit is distinct from product-market fit. Small-cohort enthusiasm is a starting point, not proof of broad demand. The essay has an enterprise emphasis; its distinction is applied here as general startup reasoning.

### S07. a16z — Improving AI-native retention

[Original source](https://a16z.com/7-ways-ai-native-companies-can-improve-retention/) · [Saved extract](./research/S07-a16z-retention.md)

Core value, thoughtful onboarding, useful notifications, and progress summaries can support retention. Notifications can also cause permanent opt-out. These are practitioner recommendations, not causal proof for this app.

### S08. Speedrun — Measuring product-market fit

[Original source](https://speedrun.substack.com/p/how-to-measure-product-market-fit) · [Saved extract](./research/S08-sr-pmf.md)

Consumer PMF assessment should look beyond easily gamed retention toward real use, monetization, and customer testimony. No quoted high-growth metrics are treated as admission requirements.

### S09. Speedrun — FAQ and selection criteria

[Original source](https://speedrun.a16z.com/faq) · [Saved extract](./research/S09-sr-faq.md)

Execution, unique insights, and founder capability are selection signals; early metrics are encouraged rather than mandatory. The current FAQ lists SR008 for late January–April 2027 and off-cycle applications.

### S10. Google — Little Language Lessons

[Original source](https://blog.google/products-and-platforms/products/education/little-language-lessons/) · [Saved extract](./research/S10-google-competitor.md)

Google has already explored situation-specific lessons, informal dialogue, and camera vocabulary. Everyday-context learning alone is not a defensible novelty claim.

### S11. Speak — Product

[Original source](https://www.speak.com/) · [Saved extract](./research/S11-speak.md)

Speak markets expert-crafted curriculum, AI personalization, real conversation, and feedback. The proposed differentiation must go beyond having a conversational tutor.

### S12. Google Calendar — Push notifications

[Original source](https://developers.google.com/workspace/calendar/api/guides/push) · [Saved extract](./research/S12-calendar-push.md)

Calendar push identifies changed resources without carrying event data; watch renewal is required. The application needs synchronization and scheduling, not only a webhook.

### S13. Google Calendar — Incremental sync

[Original source](https://developers.google.com/workspace/calendar/api/guides/sync) · [Saved extract](./research/S13-calendar-sync.md)

Sync-token pagination and invalidation must be handled; HTTP 410 requires a new sync of the affected calendar cache. This must not erase the learner's unrelated progress.

### S14. Google — Sensitive OAuth scope verification

[Original source](https://developers.google.com/identity/protocols/oauth2/production-readiness/sensitive-scope-verification) · [Saved extract](./research/S14-oauth-sensitive.md)

Sensitive scope verification requires preparation and disclosures. Public rollout depends on actual requested scopes and approval status, not a guaranteed review deadline.

### S15. Google Workspace — User data policy

[Original source](https://developers.google.com/workspace/workspace-api-user-data-developer-policy) · [Saved extract](./research/S15-google-workspace-policy.md)

Limited use applies to sensitive/restricted data and derivations; cross-user AI improvement from Workspace data is restricted. Calendar-based content stays private; consent alone does not waive provider restrictions.

### S16. Android — Geofencing

[Original source](https://developer.android.com/develop/sensors-and-location/location/geofencing) · [Saved extract](./research/S16-android-geofence.md)

Android supports geofences with entry/exit/dwell, a 100-region app/device-user limit, background permissions, and potentially minutes of latency. Native documentation recommends avoiding alert spam and not starting visible UI without user action.

### S17. Apple — Geographic condition monitoring

[Original source](https://developer.apple.com/documentation/corelocation/monitoring-the-user-s-proximity-to-geographic-regions) · [Saved extract](./research/S17-apple-geofence.md)

iOS supports geographic-condition monitoring and system wake/relaunch behavior under documented conditions, including recreation of monitors and post-reboot unlock. This supports feasibility, not a guarantee across every device state.

### S18. Google Play — Background location approval

[Original source](https://support.google.com/googleplay/android-developer/answer/9799150?hl=en) · [Saved extract](./research/S18-play-background.md)

Google Play reviews background location against core functionality and significant benefit. Technical access and store approval are separate gates.

### S19. Android — MediaSessionManager

[Original source](https://developer.android.com/reference/android/media/session/MediaSessionManager) · [Saved extract](./research/S19-android-media.md)

Android active media sessions require privileged access or enabled notification-listener access. It provides an investigable route, not approval to collect arbitrary notifications or a general right to use third-party media.

### S20. Spotify — 2026 development-mode migration

[Original source](https://developer.spotify.com/documentation/web-api/tutorials/february-2026-migration-guide) · [Saved extract](./research/S20-spotify-migration.md)

Spotify's current migration guide retains five users per new development app and a Premium requirement for the owner. It notes the July 2026 increase in client IDs per developer; that does not expand each app's user cap or authorize quota evasion.

### S21. Spotify — Developer policy

[Original source](https://developer.spotify.com/policy) · [Saved extract](./research/S21-spotify-policy.md)

Spotify policy prohibits ingesting Spotify content into AI, user profiling/analysis and other relevant uses, including games/trivia. A generative song-learning workflow needs a permitted basis; API access alone is insufficient.

### S22. Uber Eats — Marketplace API

[Original source](https://developer.uber.com/docs/eats/introduction) · [Saved extract](./research/S22-uber-eats.md)

Uber Eats Marketplace APIs are merchant/store/POS-oriented and may require written approval. The inspected documentation does not establish a general consumer order-history connector.

### S23. Schneegass et al. — Embedded smartphone language learning

[Original source](https://arxiv.org/abs/2109.07944) · [Saved extract](./research/S23-learning-phone.md)

The three-week field study reports more answers with embedded vocabulary interactions and individual differences in preference/disturbance. It does not prove that this proposed product accelerates fluency or long-term retention.

### S24. Butowska-Buczyńska et al. — Variable retrieval

[Original source](https://doi.org/10.1073/pnas.2413511121) · [Saved extract](./research/S24-learning-variable.md)

The paper record concerns variable retrieval and spacing in foreign-vocabulary learning. Only metadata and a truncated abstract were available through the index; the full-text query returned no passages. No effect size or detailed finding is relied upon.

### S25. Schwoebel et al. — Distinct episodic contexts

[Original source](https://doi.org/10.1080/09658211.2018.1464190) · [Saved extract](./research/S25-learning-spacing.md)

The abstract supports spaced retrieval and distinct episodic contexts as relevant to retention. This is narrower than real-world conversational competence and does not validate the complete product.

## Decisions supported by research versus planning assumptions

| Decision | Evidence status |
|---|---|
| Apply before a complete commercial product exists | Explicit YC and speedrun guidance |
| Recruit users individually and observe them using the MVP | Explicit YC guidance |
| Preserve a real autonomous loop while limiting deployment | Product-specific interpretation of MVP guidance |
| Calendar needs synchronization, renewal, and an independent scheduler | Documented API behavior plus implementation design |
| Geofencing can support timely background opportunities with limitations | Documented platform behavior; real-device validation still required |
| Spotify integration is gated | Current policy and access restrictions, with exact intended usage still needing review |
| Cross-user pool must exclude restricted connector-derived data | Provider policy plus conservative implementation boundary |
| A2–B1 Spanish learners abroad are the first audience | Unvalidated founder hypothesis |
| iOS first, React Native recommended | Updated user requirement; stack recommendation remains conditional on founder expertise. See solo-founder supplement |
| 20–30 skills, two prompts/day, three-hour cooldown | Initial bounded product settings to test |
| 30–50 pilot users and specified retention/conversion gates | Proposed operating targets, not investment benchmarks |
| $15/month and $450–1,800 monthly pilot cash allowance | Pricing/budget hypotheses, not quoted vendor totals |
| Context improves delayed transfer more than scheduled practice | Core product hypothesis requiring an experiment |

## How to use this research

Use the [MVP plan](./mvp-plan.md) as the working specification. Before implementation, confirm audience/platform using the first testers and run the connector feasibility gate. Recheck platform policies and accelerator deadlines before distribution/submission. Replace assumptions with observed behavior weekly.

The source material is preserved for traceability. Research is intentionally finished at a decision-useful point: the next uncertainty is best resolved through customers and live integration tests, not more general startup advice.

