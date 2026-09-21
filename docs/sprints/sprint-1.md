# Sprint 1 proposal: policy-driven service configuration

Status: Proposed, not an active accepted delivery sprint.
Owner: healdropper.
Candidate increment: v0.0.2, not a published/selected production version.
Source: [roadmap Backlog](../roadmap.md);
entry gate: [baseline review](sprint-0-baseline.md).
No feature issue is accepted yet. [Issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2)
tracks adoption only.

## Proposed goal and story

As a GoAlert operator, declare an escalation policy and connect a managed service
to it without putting installation-specific assumptions into the provider.

Proposed scope: investigate policy and step APIs, then agree the minimal typed
resource contract and key-document migration. Preserve SVC-01 through SVC-08 in
[the service spec](../specs/goalert-service.md).
Explicit exclusions: real recipients, production apply, public visibility,
Registry publication, schedules and integration-key resources.

## Decisions needed before dependent implementation

- Whether policy steps are nested or separately managed; ordering and ownership.
- Minimum useful target/action support and handling of empty policies.
- Referenced-policy deletion and safe dependency behavior.
- Exact acceptance identifiers in a proposed policy specification after the PoC.
- Sprint selection, actionable feature issue, capacity and DoD acceptance.

## Proposed capacity and defect buffer

One policy capability at a time; reserve 20% of the agreed effort for confirmed
defects. This is a planning proposal, not an accepted time estimate or deadline.
No ongoing accepted delivery item is displaced by writing this proposal.

## Proposed DoD

Accepted policy spec committed before implementation; linked issue with its
criteria; real failing behavior tests before the implementation change; relevant
Make verification; service regression and disposable policy/service acceptance;
import, drift and no-change plans; documented key migration; reviewed PR.
Merge, issue/board closure, release and deployment are recorded separately.
No implementation, TDD result, sprint acceptance or completion is asserted here.

## 2026-09-21: new discovery input, no sprint activation

The owner requested analysis of source-to-chat delivery before choosing work.
See [the proposed contract](../specs/alert-delivery.md) and
[milestone options](../milestones/alert-delivery.md). This earlier policy-first
proposal remains one option; it is not an accepted choice over ingress-first or
an end-to-end slice. No capacity, displacement, feature issue or implementation
has been committed by the new research.
