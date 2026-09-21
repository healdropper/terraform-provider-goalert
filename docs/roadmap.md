# Provider roadmap

Owner: healdropper.
Status: original service milestone accepted; future ordering and scope Proposed.
Authority: [vision](specs/00_VISION.md), [lifecycle](specs/spec-driven-lifecycle.md).
Evidence: [baseline](sprints/sprint-0-baseline.md).
No releases are currently observed; preserve the requested v0.0.x cadence.

## Current milestone

v0.0.1 candidate: [service contract](specs/goalert-service.md), under review in
[PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1).
Passing tests do not mark it merged, released, or deployed.
[Audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2)
tracks transition review; keep its baseline milestone open until reviewed.

## Backlog

| Item | Track / state | Owner | Scheduling / dependency |
| --- | --- | --- | --- |
| Review observed contracts and adoption gaps | Process transition, in review | healdropper | Baseline before normal feature delivery |
| Rebase-only governance, issue templates and private delivery board | Technical adoption gap, Proposed | healdropper | Scoped bootstrap; inspect settings before changing |
| Reconcile Release Please with GoReleaser and choose signing identity | Technical adoption gap, Proposed | healdropper | Preserve existing signing contract; before real release |
| Policy API PoC and policy/step resource ownership | C, Proposed | healdropper | Candidate v0.0.2; [Sprint 1 proposal](sprints/sprint-1.md) |
| Incoming integration-key lifecycle and secret/state semantics | C, Proposed | healdropper | After policy contract and scoped API PoC |
| Schedules, rotations and notification methods | C, unselected | healdropper | Only when concrete consumer needs justify them |
| Public visibility / Registry / production adoption | Three separate decisions | healdropper | [Existing gates](releases.md#three-independent-decisions) |

No active delivery Project URL exists yet; do not claim a configured board or
Done automation. Issue #2 is an adoption/audit tracker, not an accepted feature
issue. Proposed items are not commitments or authorization to implement their
unresolved behavior. There is no confirmed P0 or established new regression in
this transition.
