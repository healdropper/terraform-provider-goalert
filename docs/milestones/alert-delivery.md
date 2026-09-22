# Proposed milestones: from service management to useful alert delivery

Status: Proposed options for owner decision, not an active sprint.
Owner: healdropper. Date: 2026-09-21.
Track C intake: analyze Grafana -> GoAlert -> Telegram delivery.
Authority: [proposed outcomes AD-01 through AD-08](../specs/alert-delivery.md).
Evidence and uncertainty: [source review](../research/alert-delivery-feasibility.md).
The original brainstorm remains evidence of alternatives. Subsequent owner
direction and GitHub tracking are recorded in [v0.0.2](v0.0.2.md); no release
is created by either planning document.

## Milestone cards

| ID / outcome | Proposed work and deliverable | Exit evidence | Dependencies / exclusions |
| --- | --- | --- | --- |
| D0: choose a feasible delivery route | Disposable API/transport experiments; decide ordinary webhook adapter vs alternative; characterize sender failure behavior | Named results for the experiment matrix below, accepted contract delta | Owner selects investigation; no production outage or real recipients |
| M1: declare routing | Candidate escalation-policy resource with owned ordered steps and typed webhook action; document global prerequisites | AD-04/05; real CRUD, import, reordering, drift, partial failure and clean second plan | D0; no schedules/users/channel resource unless evidence requires one |
| M2: connect the alert source | Candidate Grafana integration-key resource and secure consumer handoff | AD-02/03/07 ingress tests, import/delete/replacement, credential isolation | D0 and service; can be designed alongside M1 after owner selection |
| M3: demonstrate actual notification and recovery | Consumer-owned synthetic Grafana rule/contact point, adapter and chosen test chat | AD-01/02/03/06; exact incident-to-message correlation and teardown evidence | M1/M2 or explicitly recorded manual PoC setup; adapter is not provider code |
| M4: establish delivery and upgrade confidence | Retry/dedup policy, observed delivery failures, restart recovery, key migration and release packaging | Selected AD-05/07/08 thresholds; retained evidence and known limitations | M3 plus reviewed reliability disposition; publication remains separate |
| LATER: human on-call operations | Optional user/schedule references, schedules/rotations, acknowledgement interactions, independent availability monitoring | New accepted specs and separate PoCs | Concrete user value; no implicit inclusion in v0.0.2 |

D0 is discovery, not a promised release. M3/M4 include consumer/adapter work that
must be scoped in their owning repositories; a provider release alone cannot
deliver the complete scenario. These cards are dependencies, not fixed dates or
accepted priorities.

## Proposed D0 experiment matrix

All experiments below are unexecuted proposals. Start with fake receivers and
synthetic credentials; real Telegram delivery is an explicit later test.

| Probe | Observation to record | Decision enabled |
| --- | --- | --- |
| Canonical key creates/reads a policy with one builtin-webhook action | Role requirement, accepted input shape, enabled/URL prerequisites | Minimum provider surface |
| Insert/reorder/delete steps; change repeat; interrupt a partial update | Identity, delay semantics, order, recovery and referenced-policy behavior | Nested versus independent step ownership |
| Create/read/delete/import a Grafana key | Sensitive fields, not-found behavior, replacement and old-token invalidation | Resource lifecycle and migration |
| POST pinned Grafana version-1 firing/duplicate/resolved payloads | Incident identity, counts and final state | Source compatibility |
| Run an actual synthetic rule in pinned Grafana | Evaluation, grouping, contact-point routing and recovery | End-to-end evidence beyond a sample HTTP request |
| Fake receiver records alert, status and bundle events | Payloads, sequence, metadata and repeats | Adapter contract and correlation |
| Receiver returns 401/500, delays and refuses connections | GoAlert message state versus actual receipt | Transport limitation disposition |
| Adapter uses fake Telegram API with success, 429, 5xx and timeout | Queue durability, retries, uncertain delivery and sanitized errors | Reliability contract |
| Operator-approved real test destination | Telegram acceptance and human-visible correlated message/recovery | Actual user value |

A batch above ten Grafana alerts and a missing fingerprint require explicit
experiments before claiming broader ingress support. The first PoC should not
manufacture an outage: use an isolated deterministic metric/rule. Observing an
actual Grafana outage needs an evaluator outside Grafana's failure domain.

Proposed timing measure: record condition-change, rule-firing, GoAlert creation,
adapter acceptance, Telegram acceptance and observed-receipt timestamps. Let the
owner select an SLO; neither “instant” nor an arbitrary fixed deadline is assumed.

## Options for the next provider increment

| Option | What v0.0.2 could contain | Advantage | Tradeoff |
| --- | --- | --- | --- |
| A: routing foundation | M1 after D0, with an isolated receiver demonstration | Small reusable increment, consistent with the original cadence | Does not yet deliver the complete Grafana-to-Telegram scenario |
| B: ingress foundation | M2 after D0, using an existing policy | Earlier Grafana incident creation and strong secret-lifecycle focus | Routing stays manually configured |
| C: thin end-to-end slice | Minimal M1 + M2 in provider; separate M3 adapter/consumer work | Earliest complete product demonstration | Larger scope and multiple owners; should not conceal delivery reliability debt |

Recommendation for discussion: perform D0 first, then prefer A if small provider
increments matter most, or C if the immediate learning objective is an end-to-end
demo. Do not assign M1/M2/M3 automatically to v0.0.2/v0.0.3/v0.0.4. No provider
version is selected for production, and the first v0.0.1 candidate is still
unreleased.

## Common completion and entry gates

For any selected provider resource: accepted spec and issue committed before
implementation, red-first behavior tests, real disposable API/Framework
acceptance, import, drift, auth-error state preservation, no-change second plan,
docs and migration guidance, reviewed PR. Keep service regression coverage.

Before normal implementation, resolve/schedule the remaining
[SpecDD adoption gaps](../roadmap.md), accept the selected scope and criteria,
and replace the [Sprint 1 proposal](../sprints/sprint-1.md) with the owner's
actual decision. A suggested defect buffer is planning input, not an estimate.
Before a real release, reconcile release management/signing and verify the
maintained signer. Public visibility, Registry publication and production
adoption remain three independent decisions.

## Owner decision worksheet

Recorded answers: Grafana firing first; small increments; group preferred with
private visibility possible; outage detection later. See [v0.0.2](v0.0.2.md).
The remaining questions below are not answered by implication:
1. First failure to demonstrate: synthetic workload condition, Grafana outage,
   or an ordered pair of separate scenarios?
2. Destination: private chat, group or channel; existing or dedicated bot?
3. User experience: firing only, firing/recovery, or interactive acknowledgement?
4. Next increment: A, B or C after discovery?
5. Delivery owner: small standalone adapter, extension of an existing bot
   service, or investigate a GoAlert notification plugin?
6. Acceptable latency, reminder frequency, retry limit and message retention?

An existing application that sends Telegram messages is not automatically a
webhook adapter. Verify its input contract, failure reporting, secret handling
and deployment ownership before selecting reuse.
