# Provider roadmap

Owner: healdropper.
Status: original service milestone accepted; future ordering and scope Proposed.
Authority: [vision](specs/00_VISION.md), [lifecycle](specs/spec-driven-lifecycle.md).
Evidence: [baseline](sprints/sprint-0-baseline.md).
No releases are currently observed; preserve the requested v0.0.x cadence.

## Current milestone

v0.0.1 candidate: [service contract](specs/goalert-service.md), merged through
[PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1) on
2026-09-21 at `77b6c1c13f5608328206f76ac4668cb1bb0524b1`.
Merged does not mean released or deployed; no provider release was observed.
[Audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2)
tracks transition review; keep its baseline milestone open until reviewed.

## Backlog

| Item | Track / state | Owner | Scheduling / dependency |
| --- | --- | --- | --- |
| Review observed contracts and adoption gaps | Process transition, in review | healdropper | Baseline before normal feature delivery |
| Rebase-only governance, issue templates and private delivery board | Technical adoption gap, Proposed | healdropper | Scoped bootstrap; inspect settings before changing |
| Reconcile Release Please with GoReleaser and choose signing identity | Technical adoption gap, Proposed | healdropper | Preserve existing signing contract; before real release |
| Alert delivery discovery and policy/step ownership | C, Proposed | healdropper | D0/M1 in [milestone options](milestones/alert-delivery.md); no selected v0.0.2 scope |
| Incoming integration-key lifecycle and secret/state semantics | C, Proposed | healdropper | M2 in [milestone options](milestones/alert-delivery.md); order awaits owner decision |
| External chat adapter and real source-to-chat demonstration | C, Proposed; consumer-owned delivery | healdropper | M3/M4 in [milestone options](milestones/alert-delivery.md); outside provider runtime |
| Schedules, rotations and notification methods | C, unselected | healdropper | Only when concrete consumer needs justify them |
| Public visibility / Registry / production adoption | Three separate decisions | healdropper | [Existing gates](releases.md#three-independent-decisions) |

No active delivery Project URL exists yet; do not claim a configured board or
Done automation. Issue #2 is an adoption/audit tracker, not an accepted feature
issue. Proposed items are not commitments or authorization to implement their
unresolved behavior. There is no confirmed P0 or established new regression in
this transition.

## Alert delivery discovery

The [proposed contract](specs/alert-delivery.md) defines generic ownership and
candidate outcomes. The [source review](research/alert-delivery-feasibility.md)
records evidence and limits. The [milestone cards and decision worksheet](milestones/alert-delivery.md)
compare possible v0.0.2 slices. Their priorities and release assignments are not
accepted. Keep [Sprint 1](sprints/sprint-1.md) proposed until the owner chooses;
this intake does not silently activate or expand it.
