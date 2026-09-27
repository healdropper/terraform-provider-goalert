# Provider roadmap

Owner: healdropper.
Status: v0.0.1 (services), v0.0.2 (escalation policies with webhook routing), and v0.0.3 (ingress and integration keys) completed. v0.0.4 (heartbeats, labels, data sources, and v0.35.0) in planning.
Authority: [vision](specs/00_VISION.md), [lifecycle](specs/spec-driven-lifecycle.md).
Evidence: [baseline](sprints/sprint-0-baseline.md).
Preserve the v0.0.x cadence during private development until the v1.0.0 Registry publication milestone.

## Completed milestones

- **v0.0.1**: [service contract](specs/goalert-service.md), merged through
  [PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1) on
  2026-09-21.
- **v0.0.2**: [routing increment](milestones/v0.0.2.md), merged through
  [PR #17](https://github.com/healdropper/terraform-provider-goalert/pull/17) on
  2026-09-22. Successfully deployed and adopted in production cluster `cenarion-watch`.
- **v0.0.3**: [ingress and integration keys](milestones/v0.0.3.md), merged through
  [PR #30](https://github.com/healdropper/terraform-provider-goalert/pull/30) on
  2026-09-26. Resources: `goalert_integration_key`. Closed on GitHub.

## Active milestone

- **v0.0.4**: [heartbeats, labels and data sources](milestones/v0.0.4.md), planned in
  [Sprint 3](sprints/sprint-3.md) and tracked in
  [GitHub Milestone 4](https://github.com/healdropper/terraform-provider-goalert/milestone/4).
  - Scope: `goalert_heartbeat_monitor` resource (#18), `goalert_service_label` resource, foundational data sources (`service`, `escalation_policy`, `integration_key`, `heartbeat_monitor`), and GoAlert v0.35.0 compatibility (#22).

## Planned milestones (full API coverage & Registry launch)

- **v0.0.5**: [human identity, contact methods, and notification rules](milestones/v0.0.5-preview.md).
  - Scope: `goalert_user`, `goalert_user_contact_method`, `goalert_user_notification_rule` resources and user data source.
- **v0.0.6**: [on-call shift rotations and escalation targets](milestones/v0.0.6-preview.md).
  - Scope: `goalert_rotation` resource, data source, and expanding `goalert_escalation_policy` steps to target users and rotations.
- **v0.0.7**: [schedules, shifts, targets, and user overrides](milestones/v0.0.7-preview.md).
  - Scope: `goalert_schedule`, `goalert_schedule_rule`, `goalert_user_override`, `goalert_schedule_on_call_notification_rule`, and schedule data source.
- **v0.0.8**: [collaboration channels and system limits](milestones/v0.0.8-preview.md).
  - Scope: Slack channel/user group data sources and `goalert_system_limit` resource.
- **v1.0.0**: [Terraform Registry publication and GA release](milestones/v1.0.0-preview.md).
  - Scope: `tfplugindocs` pipeline integration, GPG signing identity setup, public repository transition, registry publication, and production consumer adoption.

## Backlog

| Item | Track / state | Owner | Scheduling / dependency |
| --- | --- | --- | --- |
| Review observed contracts and adoption gaps | Process transition, in review | healdropper | Baseline before normal feature delivery |
| Upstream GoAlert v0.35.0 compatibility | Maintenance, Ready | healdropper | Assigned to v0.0.4 (#22) |
| Heartbeat monitor lifecycle & dead-man switch | C, In planning | healdropper | Assigned to v0.0.4 (#18, #32, #33, #34) |
| User identity, contact methods, notification rules | C, Backlog | healdropper | Scheduled for v0.0.5 |
| Schedules, rotations and notification methods | C, Backlog | healdropper | Scheduled for v0.0.6 (#19) and v0.0.7 |
| External chat adapter and real source-to-chat demonstration | C, Proposed; consumer-owned delivery | healdropper | Outside provider runtime (#11) |
| Independent Grafana availability | C, Backlog | healdropper | [Availability RFC](specs/grafana-availability.md) (#10) |
| Public visibility / Registry / production adoption | Three separate decisions | healdropper | Scheduled for v1.0.0; [Existing gates](releases.md#three-independent-decisions) |

## Delivery tracking

[Private Project](https://github.com/users/healdropper/projects/1): Delivery board,
Board layout with Status columns.
Additional views track active milestones and unassigned Backlog issues.
Item closed and Pull request merged workflows report enabled.

[Milestone v0.0.4](https://github.com/healdropper/terraform-provider-goalert/milestone/4)
is the active delivery planning container.
Its [canonical record](milestones/v0.0.4.md) links readiness and dependencies across
V004-UPSTREAM, V004-DISC, V004-SPEC, V004-IMPL, and V004-VERIFY gates.
