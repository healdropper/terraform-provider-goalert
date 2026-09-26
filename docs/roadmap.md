# Provider roadmap

Owner: healdropper.
Status: v0.0.1 (services), v0.0.2 (escalation policies with webhook routing), and v0.0.3 (ingress and integration keys) completed.
Authority: [vision](specs/00_VISION.md), [lifecycle](specs/spec-driven-lifecycle.md).
Evidence: [baseline](sprints/sprint-0-baseline.md).
No releases are currently observed; preserve the requested v0.0.x cadence.

## Completed milestones

- **v0.0.1**: [service contract](specs/goalert-service.md), merged through
  [PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1) on
  2026-09-21.
- **v0.0.2**: [routing increment](milestones/v0.0.2.md), merged through
  [PR #17](https://github.com/healdropper/terraform-provider-goalert/pull/17) on
  2026-09-22. Successfully deployed and adopted in production cluster `downstream-consumer`.
- **v0.0.3**: [ingress and integration keys](milestones/v0.0.3.md), merged through
  [PR #30](https://github.com/healdropper/terraform-provider-goalert/pull/30) on
  2026-09-26. Resources: `goalert_integration_key`. Closed on GitHub.


## Backlog

| Item | Track / state | Owner | Scheduling / dependency |
| --- | --- | --- | --- |
| Review observed contracts and adoption gaps | Process transition, in review | healdropper | Baseline before normal feature delivery |
| Remaining governance and delivery prerequisites | Technical adoption gap, Complete | healdropper | Resolved in issue #4; Project 1 automation and templates active; [v0.0.2](milestones/v0.0.2.md) |
| Reconcile Release Please with GoReleaser and choose signing identity | Technical adoption gap, Proposed | healdropper | Preserve existing signing contract; before real release |
| Alert delivery discovery and policy/step ownership | C, Complete | healdropper | Resolved in issue #5 and #6; canonical spec accepted; [v0.0.2](milestones/v0.0.2.md) |
| Incoming integration-key lifecycle and secret/state semantics | C, Proposed | healdropper | M2 in [milestone options](milestones/alert-delivery.md); order awaits owner decision |
| External chat adapter and real source-to-chat demonstration | C, Proposed; consumer-owned delivery | healdropper | M3/M4 in [milestone options](milestones/alert-delivery.md); outside provider runtime |
| Independent Grafana availability | C, Backlog | healdropper | [Availability RFC](specs/grafana-availability.md); excluded from v0.0.2 |
| Schedules, rotations and notification methods | C, unselected | healdropper | Only when concrete consumer needs justify them |
| Public visibility / Registry / production adoption | Three separate decisions | healdropper | [Existing gates](releases.md#three-independent-decisions) |

## Delivery tracking

[Private Project](https://github.com/users/healdropper/projects/1): Delivery board,
Board layout with Status columns, verified through GitHub API on 2026-09-21.
Additional views separate the v0.0.2 milestone from unassigned Backlog issues.
The repository is linked; ten items were verified: five v0.0.2 issues, three
later backlog issues, baseline issue #2 and documentation PR #3.
Item closed and Pull request merged workflows report enabled. Their actual
closure-to-Done transition must be checked on real reviewed work; no throwaway
issue was closed to manufacture evidence. Auto-close issue also reports enabled:
never move an unfinished card to Done, which can close its issue.

[Milestone v0.0.2](https://github.com/healdropper/terraform-provider-goalert/milestone/2)
is a delivery planning container, not a release. Its
[canonical record](milestones/v0.0.2.md) links readiness and dependencies.
Audit issue #2 and its baseline milestone remain open pending their review.

## Alert delivery discovery

The [proposed contract](specs/alert-delivery.md) defines generic ownership and
candidate outcomes. The [source review](research/alert-delivery-feasibility.md)
records evidence and limits. The [milestone cards and decision worksheet](milestones/alert-delivery.md)
preserve alternative slices. The owner selected small increments and Grafana
firing first; [v0.0.2 planning](milestones/v0.0.2.md) and
[Sprint 1](sprints/sprint-1.md) now record readiness without implying an accepted
resource schema or completed PoC. Later release assignments remain undecided.
