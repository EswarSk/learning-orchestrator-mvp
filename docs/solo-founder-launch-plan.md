# Solo-founder launch, free tier, cloud credits, and company setup

Updated September 17, 2026. This supplement supersedes the original Android-first/two-builder assumptions. Confirmed: one developer, iOS is required, Android is desirable, and the intended free offer is 25 interactive sessions across five Spanish-learning milestones. Jurisdiction, work authorization, existing stack, company name, incorporation, and prior credits remain unconfirmed.

## Product and engineering decision

Recommend **React Native + TypeScript + Expo development builds, with iOS released first**, if React/TypeScript is familiar. If Swift is substantially stronger, native SwiftUI is likely the faster route. Do not maintain independent Swift and Kotlin apps as a solo founder initially.

React Native shares UI and business logic; it does not remove native permissions, platform-specific background execution, audio handling, billing, or separate Android QA. Expo Go is insufficient for validating background location. Use a real iPhone, a development build, and TestFlight. Expo documents an iOS 20-region geofencing limit and Always authorization for background geofencing. Use a bounded region set and investigate native modules if necessary. [N06]

Keep the TypeScript API, Postgres records, durable jobs/scheduler, permission gates, personal opportunity pool, and provenance-restricted shared pool from the main plan. Shared learning remains part of the pilot; its first version is a small catalog with automatic extraction and operator quality review.

No general iOS API for observing every other app's activity has been verified. React Native does not unlock that capability. The first autonomous integrations are server-side Google Calendar and native location. Cross-app music stays gated; do not substitute a simulated cue and describe it as real autonomous detection.

### What “on par with Duolingo” means for this release

Match the quality of the **first 25-session experience**: coherent visual design, clear onboarding, fluid navigation, excellent audio, reliable progress, rewarding feedback, useful reminders, accessible interactions, and graceful recovery.

Beat alternatives on one demonstrable outcome: recognizing useful real-life opportunities and carrying the learner's difficulties into later practice. Full feature/content/experimentation parity with a mature product is not a credible solo-founder release commitment. Do not advertise overall superiority before comparative evidence.

Ship: original visual identity, five-milestone progress path, reusable session player, audio controls, contextual invitation cards, error review, progress recap, a modest weekly goal, subscription screen, account/privacy/settings, and billing restoration. Keep the screens consistent and polished rather than building a large social game.

Avoid copying Duolingo's assets, mascot, sounds, text, or distinctive branding. Budget for a Spanish educator's review even if all development is done alone.

### Solo schedule

- Weeks 1–2: five to ten user conversations, iOS background proof, curriculum skeleton, design system, domain/landing page, credit application materials.
- Weeks 3–4: first five polished sessions, real autonomous offer, personal progress, and a working demonstration.
- Weeks 5–8: all 25 session structures, contextual adaptation, delayed review, private TestFlight pilot, shared-pattern proof, analytics, privacy/deletion.
- Weeks 9–12: paid continuation, receipt/entitlement handling, quality/reliability fixes, real retention analysis, public-release preparation.
- Android follows after the iOS core works and actual tester demand justifies its QA; shared code makes it an extension, not zero additional work.

These assume substantial weekly development time and existing programming competence. Public store approval, content review, and unknown integrations can extend the timeline. Apply to accelerators with genuine progress before the complete release rather than wait for perfection.

## Free product: 25 sessions, five milestones

Use five **product milestones**, not five CEFR levels or a promise of Spanish mastery.

| Milestone | Functional goal | Five-session structure |
|---|---|---|
| 1. Start and repair a conversation | Introduce yourself; ask someone to repeat or slow down | Two guided sessions, two context-adapted sessions, one transfer check |
| 2. Get what you need | Order, ask prices, express preferences | Same structure, increasingly independent responses |
| 3. Find your way | Directions, transport, time, appointments | Same structure with listening emphasis |
| 4. Connect with people | Interests, plans, follow-up questions | Same structure with sustained dialogue |
| 5. Describe your experiences | Explain a recent event and resolve a misunderstanding | Same structure with earlier skills recombined |

This is 25 bounded session slots with adaptive content, not 25 static scripts and not five exhaustive proficiency bands. If real context is absent, offer a clearly labeled fallback; do not lock progression behind location/calendar permissions or claim a fallback was triggered externally.

**Recommended free entitlement:**
- 25 sessions per account, once, not a monthly reset.
- Up to three minutes of active AI interaction in a session, maximum 75 minutes for the full allowance.
- Contextual opportunities are available during the free experience, including an early opportunity after connection; the main differentiator must be experienced before the paywall.
- No payment card required for the free allowance.
- After the allowance, new AI interaction requires payment. Keep existing results, account controls, data export/deletion, and cached non-AI review accessible.
- Do not require a perfect score to progress. Offer appropriate retry/scaffolding and label evidence honestly.

Define the allowance in onboarding and before session start. Notifications and previews do not spend a session. Reserve a slot atomically on start, finalize it once useful AI interaction is delivered, and restore it on a verified technical failure before useful delivery. Reconnection resumes the same session. A deliberately abandoned session after useful interaction still consumes its slot; disclose this. Support reasonable recovery without letting endless abandon/restart cycles create unlimited audio cost.

Enforce access on the server, including concurrency limits, remaining active audio time, and entitlement checks. Do not trust a client counter. Handle subscription purchases, renewals, cancellation at period end, refunds/revocations, restore purchases, and duplicate webhook delivery. Add runnable checks for the 25th/26th sessions and concurrent starts.

On iOS, use StoreKit in-app subscriptions as the uncomplicated initial route for digital access. RevenueCat is an optional installed service choice if its current pricing/terms justify reducing billing work; it does not replace Apple billing rules. Regional external-payment exceptions exist, but are not needed to validate this MVP. [N08]

**Paid hypothesis:** one monthly plan around $14.99, with an explicit AI-minute allowance chosen after benchmarking. A provisional 100 minutes/month can be evaluated, but must not be promised until cost is measured. Subscription value is ongoing contextual adaptation and practice, not just a static pack of 25 lessons. Ensure users retain access to what was promised and see renewal price/terms clearly.

### Economics of the free allowance

These are sensitivity assumptions, **not quotes from an AI vendor**. Effective cost must include speech input/output, model work, and session overhead.

| Assumed effective AI cost per active minute | Fully used 75-minute free allowance | 1,000 fully used allowances |
|---|---:|---:|
| $0.02 | $1.50 | $1,500 |
| $0.05 | $3.75 | $3,750 |
| $0.10 | $7.50 | $7,500 |

Background generation, hosting, support, taxes, and fees are additional. Measure both actual consumption per signup and the fully consumed liability. At $3.75 per free user and 5% conversion, free-session cost alone implies $75 per converted payer before acquisition/support. This is why credit-funded signups are not the same as a sustainable business.

Start with 20–50 invitees, cap the pilot's total usage, use reusable reviewed lesson assets, and generate personalized content only when needed. Cloud credits should reduce experimentation cost, not conceal poor paid margins.

Measure the free funnel: signup → first useful session → first autonomous session → milestone progression → allowance exhausted → paid conversion → paid renewal. Users who finish early encounter the paywall sooner; segment retention by entitlement state instead of calling every post-limit inactive user an engagement failure. Report TestFlight/sandbox purchases separately from actual revenue.

## Cloud-credit shortlist and application strategy

Current official pages were inspected; amounts are conditional, not guaranteed cash awards. Pick one primary cloud after a small voice/hosting benchmark. Do not deploy three clouds simply because three offer credits.

| Program | Realistic starting route | Larger offer and conditions | Recommendation |
|---|---|---|---|
| AWS Activate Founders | $1,000 initially; selected participants can grow to $5,000 | Portfolio up to $200,000 requires an Activate Provider Org ID; further AI credits are invitation-based | Apply directly as bootstrapped if eligible; do not invent an accelerator affiliation |
| Google for Startups Cloud, Start tier | Up to $2,000 valid one year | Scale up to $200,000 or $350,000 AI-first is tied to funded-stage criteria; year-two coverage is partial | Strong candidate for Google Cloud/Firebase/Gemini; current Start description asks for a working MVP and clear business model |
| Microsoft for Startups | Up to $200 after initial signup/identity step; $5,000 after business verification | Current docs describe milestone progression up to $150,000 based on sustained service usage and other requirements | Apply after legal entity verification if Azure fits; do not add unnecessary services to chase higher milestones |

**AWS eligibility inspected:** pre-Series B, founded within ten years, AWS account on a Paid Tier Plan, and new to credits or applying for more than previously received. Founders is for self-funded startups; Portfolio requires a valid provider Org ID. Credits do not prevent charges outside eligible coverage. [N01]

**Google Start eligibility inspected:** technology startup planning venture funding, founded within 24 months, no prior Cloud credits beyond the free trial, clear roadmap; the tier description specifies a working MVP and business model. Eligible credits cover Google models such as Gemini/Gemma; the page says third-party models are billed directly and not covered. Do not assume credits pay for an external AI API or every Maps expense. [N02]

**Microsoft inspected:** initial identity verification uses a personal LinkedIn account; $5,000 milestone requires a registered legal entity. Higher milestones are based on verified usage over time; the $150,000 milestone currently specifies 10+ workloads and around $3,000/month sustained Azure usage. These are not reasons to inflate the architecture. [N03]

Recommended application order: prepare company identity and a live landing page now; apply to AWS Founders if eligibility is met; complete Microsoft business verification when the company exists; submit Google Start once a small working demo/MVP is available. If the actual primary stack is already settled, prioritize its program first.

**Cloud architecture choice:** retain one backend and managed database, using the selected provider's container hosting or equivalent managed runtime. If choosing GCP, Cloud Run + managed Postgres + managed scheduling/queue is a reasonable mapping; if choosing another cloud, use equivalent native managed services. Database baseline costs can dominate a tiny app, so price the minimum before committing. Do not assume a cloud grant covers Supabase, Expo builds, Apple membership, or third-party billing services.

Use budget alerts plus application-enforced session/usage limits. Alerts alone are not hard spending caps. Track grant amount, eligible services, activation date, expiry, remaining balance, and expected cost after expiration. Avoid activating a time-limited grant months before useful development.

### Application packet

Ready-to-adapt description:

> We are building an iOS-first AI language-learning companion that turns everyday experiences into personalized Spanish practice. With user permission, calendar and location signals create timely speaking and listening opportunities. The product tracks skill evidence and revisits difficulties in new contexts. The first release offers 25 interactive sessions across five practical milestones, followed by paid ongoing practice. One founder is developing the product.

Use future tense for unbuilt capabilities. Do not claim users, revenue, funding, incorporation, or partnerships that do not exist.

Suggested technical-use statement:

> We expect to use managed application hosting, a database, background jobs, secure credential storage, monitoring, and eligible speech/language-model services. Initial deployment is a small invite-only pilot. We will measure cost per active learner and scale with retained usage.

Fields still needed:
- Startup/product name and public website/domain email.
- Founder legal name and authorized contact.
- Country/state, entity type, legal name and formation date, if incorporated.
- Funding stage, funding dates/amounts, investor/accelerator affiliations only if real.
- Previous credits for each provider.
- Current prototype status and honest user/revenue figures.
- Provider account/billing identifiers and planned services.
- A short demo, LinkedIn profile, and incorporation evidence where required.

Application links:
- [AWS Activate](https://aws.amazon.com/startups/join?destination=/credits/apply)
- [Google Start](https://cloud.google.com/startup/apply)
- [Microsoft for Startups](https://www.microsoft.com/en-us/startups)

No applications have been submitted. The packet is prepared; missing identity/eligibility information must be supplied before forms can be completed accurately. Enter passwords, identity documents, and billing details directly in the providers' secure interfaces.

## Company setup: what is required at each stage

Jurisdiction and employment/visa status are not yet known. The steps below are a **conditional US venture-startup route**, not a determination that a US entity is best for this founder. A local startup lawyer/accountant should confirm entity choice and local registrations; immigration authorization is a separate issue from owning a company.

### Before interviewing or pitching the concept

You generally do not need a corporation just to describe an idea, interview potential customers, or demonstrate a prototype. YC also explicitly says incorporation is not necessary to apply. This is distinct from running a paid service, signing contracts, employing people, or soliciting investments.

Use truthful prototype claims. If collecting email addresses, explain the purpose and provide a suitable privacy notice and opt-out. Review employment/IP obligations before putting employer-related code or inventions into the product. If work authorization is restricted, do not infer that unpaid startup development is automatically permitted.

### Before a data-collecting pilot

Establish who operates the service and is responsible for personal data. Prepare a privacy notice and terms matching location, calendar, voice, AI processing, retention, and sharing. Provide consent and account/data deletion. Review third-party processor terms. Initially recruit adults and avoid recording bystanders.

An entity is prudent before taking on material contractual/data liability, even where not universally mandatory. Choose the entity based on residence, taxation, funding plans, and operating location.

### If forming a US venture-backed company

A Delaware C corporation is a common fit for US institutional fundraising; it is not mandatory for every startup or every foreign founder.

1. Confirm entity/jurisdiction and relevant founder work authorization with counsel.
2. Check company name availability and avoid obvious brand conflicts.
3. Appoint a Delaware registered agent and file the formation document.
4. Complete bylaws, initial board actions, founder stock issuance, any vesting arrangements, cap table, and assignment of company IP.
5. If restricted founder stock is issued, promptly evaluate the Section 83(b) election deadline with the formation provider/tax adviser; this can be time-sensitive. Do not assume filing is complete merely because incorporation is complete.
6. Obtain an EIN after forming the entity. The IRS charges no EIN fee; online eligibility depends on the responsible party and business location. [N10]
7. Open a business bank account and set up bookkeeping.
8. Determine whether the operating state/country requires foreign qualification, local licenses, or additional taxes. Delaware incorporation does not eliminate those obligations.
9. Calendar federal/state tax filings, Delaware annual report/franchise tax, registered-agent renewal, and other applicable obligations—even in a zero-revenue year.
10. Establish contractor IP/confidentiality agreements before outsourcing design/content work.

One available formation route is Stripe Atlas: its inspected documentation lists $500 covering incorporation/state fees and the first year of registered-agent service, then $100/year for the agent. It does not eliminate ongoing taxes, operating-state fees, accounting, or special cross-border advice. Its standard documents may not suit existing valuable IP or unusual ownership situations. [N12]

Avoid stale checklists: FinCEN's current page says US-created companies are exempt from BOI reporting; some foreign-formed companies registered in the US can remain subject to requirements. Recheck the current rule when forming. [N13]

### Apple distribution and payments

Apple allows individual enrollment; your personal legal name is then shown as seller. Organization enrollment requires a legal entity, D-U-N-S number, binding authority, a working company website, and organization-domain email. Membership is $99/year in the US, with regional variation. A single founder can own a qualifying corporation; “solo” does not force individual enrollment. [N07]

If the brand must be the seller from day one, form the entity and start organization verification early. Do not delay interviews while waiting for it.

Before public paid release: complete store agreements and tax/banking information; use appropriate subscription billing; show price and renewal terms; support restore, cancellation guidance, and in-app account deletion. Disclose AI data sharing and obtain required permission; don't expose calendar details in lock-screen notifications. [N08]

## Next concrete steps

1. Confirm residence/operating location and current technology strengths.
2. Choose a working name/domain and publish a truthful landing page.
3. Build the first iOS background-trigger proof and five-session milestone.
4. Prepare entity formation only after jurisdiction/work/IP considerations are clear.
5. Complete the cloud applications using the prepared packet and actual eligibility facts.
6. Recruit the first 20–50 free users, measure costs, then expand the 25-session offer.

## New sources


- **N01:** [AWS Activate credits](https://aws.amazon.com/startups/credits/) · [saved extract](./research/N01-aws.md)

- **N02:** [Google startup benefits](https://cloud.google.com/startup/benefits) · [saved extract](./research/N02-gcp.md)

- **N03:** [Microsoft startup milestones](https://learn.microsoft.com/en-us/startups/microsoft-for-startups/getting-started-mfs) · [saved extract](./research/N03-azure.md)

- **N06:** [Expo Location](https://docs.expo.dev/versions/latest/sdk/location/) · [saved extract](./research/N06-expo.md)

- **N07:** [Apple Developer enrollment](https://developer.apple.com/programs/enroll/) · [saved extract](./research/N07-apple.md)

- **N08:** [Apple App Review Guidelines](https://developer.apple.com/app-store/review/guidelines/) · [saved extract](./research/N08-apple-review.md)

- **N10:** [IRS EIN guidance](https://www.irs.gov/businesses/small-businesses-self-employed/get-an-employer-identification-number) · [saved extract](./research/N10-irs.md)

- **N11:** [Delaware entity formation](https://corp.delaware.gov/howtoform/) · [saved extract](./research/N11-delaware.md)

- **N12:** [Stripe Atlas incorporation](https://docs.stripe.com/atlas/signup) · [saved extract](./research/N12-atlas.md)

- **N13:** [FinCEN BOI guidance](https://www.fincen.gov/boi) · [saved extract](./research/N13-boi.md)

