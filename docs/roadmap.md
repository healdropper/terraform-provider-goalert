# Provider roadmap

Owner: healdropper.
Status: Service Contract, Escalation Policies and Webhook Routing, Ingress and Integration Keys, Heartbeat Monitors and Service Labels, and User Identity and Notification Rules completed. Rotations and Escalation Targets active next.
Authority: [vision](specs/00_VISION.md), [lifecycle](specs/spec-driven-lifecycle.md).
Evidence: [baseline](sprints/sprint-0-baseline.md).
Delivery milestones represent human-scoped goals and functional capability themes decoupled from strict SemVer patch increments.

## Completed milestones

- **Service Contract**: [service contract](specs/goalert-service.md), merged through
  [PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1) on
  2026-09-21. Resources: `goalert_service`.
- **Escalation Policies and Webhook Routing**: [routing increment](milestones/v0.0.2.md), merged through
  [PR #17](https://github.com/healdropper/terraform-provider-goalert/pull/17) on
  2026-09-22. Resources: `goalert_escalation_policy` (with webhook step target). Successfully deployed and adopted in production cluster `cenarion-watch`.
- **Ingress and Integration Keys**: [ingress and integration keys](milestones/v0.0.3.md), merged through
  [PR #30](https://github.com/healdropper/terraform-provider-goalert/pull/30) on
  2026-09-26. Resources: `goalert_integration_key`. Closed on GitHub.
- **Heartbeat Monitors and Service Labels**: [heartbeats, labels and data sources](milestones/v0.0.4.md), merged through
  [PR #36](https://github.com/healdropper/terraform-provider-goalert/pull/36) on
  2026-09-27. Resources: `goalert_heartbeat_monitor`, `goalert_service_label`.
  Data sources: `goalert_service`, `goalert_escalation_policy`, `goalert_integration_key`, `goalert_heartbeat_monitor`.
  Closed on GitHub.
- **User Identity and Notification Rules**: [human identity, contact methods, and notification rules](milestones/v0.0.5.md), merged through
  [PR #42](https://github.com/healdropper/terraform-provider-goalert/pull/42) on 2026-09-27.
  Resources: `goalert_user`, `goalert_user_contact_method`, `goalert_user_notification_rule`.
  Data sources: `goalert_user`. Closed on GitHub.

## Active milestone

- **Rotations and Escalation Targets**: [on-call shift rotations and escalation targets](milestones/rotations-and-escalation-targets.md).
  - Scope: `goalert_rotation` resource, `data.goalert_rotation` data source, and expanding `goalert_escalation_policy` steps to target users and rotations.
  - Tracking: [GitHub Milestone 6](https://github.com/healdropper/terraform-provider-goalert/milestone/6).

## Planned milestones (full API coverage & Registry launch)

- **Schedules and User Overrides**:
  - Scope: `goalert_schedule`, `goalert_schedule_rule`, `goalert_user_override`, `goalert_schedule_on_call_notification_rule`, and schedule data source.
- **Collaboration Channels and System Limits**:
  - Scope: Slack/chat channel/user group data sources and `goalert_system_limit` resource.
- **Registry Publication and GA**:
  - Scope: `tfplugindocs` pipeline integration, GPG signing identity setup, public repository transition, registry publication, and production consumer adoption.

## Backlog

| Item | Track / state | Owner | Scheduling / dependency |
| --- | --- | --- | --- |
| Review observed contracts and adoption gaps | Process transition, in review | healdropper | Baseline before normal feature delivery |
| External chat adapter and real source-to-chat demonstration | C, Proposed; consumer-owned delivery | healdropper | Outside provider runtime (#11) |
| Independent Grafana availability | C, Backlog | healdropper | [Availability RFC](specs/grafana-availability.md) (#10) |
| Public visibility / Registry / production adoption | Three separate decisions | healdropper | Scheduled for Registry Publication milestone; [Existing gates](releases.md#three-independent-decisions) |

## Delivery tracking

[Private Project](https://github.com/users/healdropper/projects/1): Delivery board,
Board layout with Status columns.
Additional views track active milestones and unassigned Backlog issues.
Item closed and Pull request merged workflows report enabled.

[Milestone: Rotations and Escalation Targets](https://github.com/healdropper/terraform-provider-goalert/milestone/6)
is the active delivery planning container.
Its canonical record links readiness and dependencies across
DISC, SPEC, IMPL, and VERIFY gates.
