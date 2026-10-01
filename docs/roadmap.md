# Provider roadmap

Owner: healdropper.
Status: Milestones 1 through 9 completed. Registry Publication and GA (Milestone 10) active next.
Authority: [vision](specs/00_VISION.md), [lifecycle](specs/spec-driven-lifecycle.md).
Evidence: [baseline](sprints/sprint-0-baseline.md).
Delivery milestones represent human-scoped goals and functional capability themes decoupled from strict SemVer patch increments.

## Completed milestones

- **Service Contract**: [service contract](specs/goalert-service.md), merged through
  [PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1) on
  2026-09-21. Resources: `goalert_service`.
- **Escalation Policies and Webhook Routing**: [routing increment](milestones/escalation-policies-and-webhook-routing.md), merged through
  [PR #17](https://github.com/healdropper/terraform-provider-goalert/pull/17) on
  2026-09-22. Resources: `goalert_escalation_policy` (with webhook step target). Successfully deployed and adopted in production cluster `cenarion-watch`.
- **Ingress and Integration Keys**: [ingress and integration keys](milestones/ingress-and-integration-keys.md), merged through
  [PR #30](https://github.com/healdropper/terraform-provider-goalert/pull/30) on
  2026-09-26. Resources: `goalert_integration_key`. Closed on GitHub.
- **Heartbeat Monitors and Service Labels**: [heartbeats, labels and data sources](milestones/heartbeat-monitors-and-service-labels.md), merged through
  [PR #36](https://github.com/healdropper/terraform-provider-goalert/pull/36) on
  2026-09-27. Resources: `goalert_heartbeat_monitor`, `goalert_service_label`.
  Data sources: `goalert_service`, `goalert_escalation_policy`, `goalert_integration_key`, `goalert_heartbeat_monitor`.
  Closed on GitHub.
- **User Identity and Notification Rules**: [human identity, contact methods, and notification rules](milestones/user-identity-and-notification-rules.md), merged through
  [PR #42](https://github.com/healdropper/terraform-provider-goalert/pull/42) on 2026-09-27.
  Resources: `goalert_user`, `goalert_user_contact_method`, `goalert_user_notification_rule`.
  Data sources: `goalert_user`. Closed on GitHub.
- **Rotations and Escalation Targets**: [on-call shift rotations and escalation targets](milestones/rotations-and-escalation-targets.md), merged through
  [PR #50](https://github.com/healdropper/terraform-provider-goalert/pull/50) and [PR #51](https://github.com/healdropper/terraform-provider-goalert/pull/51) on 2026-09-28.
  Resources: `goalert_rotation`, extended `goalert_escalation_policy` step targets.
  Data sources: `goalert_rotation`. Closed on GitHub.
- **Schedules and User Overrides**: [schedules, rules, overrides, and on-call notification rules](milestones/schedules-and-user-overrides.md), merged through
  [PR #59](https://github.com/healdropper/terraform-provider-goalert/pull/59) and [PR #60](https://github.com/healdropper/terraform-provider-goalert/pull/60) on 2026-09-28.
  Resources: `goalert_schedule`, `goalert_schedule_rule`, `goalert_user_override`, extended `goalert_escalation_policy` step targets (`schedule_ids`).
  Data sources: `goalert_schedule`. Closed on GitHub.
- **Collaboration Channels and System Limits**: [collaboration channels and system limits](milestones/collaboration-channels-and-system-limits.md), merged through
  [PR #68](https://github.com/healdropper/terraform-provider-goalert/pull/68) and [PR #69](https://github.com/healdropper/terraform-provider-goalert/pull/69) on 2026-09-28.
  Resources: `goalert_system_limit`.
  Data sources: `goalert_slack_channel`, `goalert_slack_user_group`.
  Closed on GitHub.
- **Embrace Latest GoAlert Version**: [embrace latest goalert version](milestones/embrace-latest-goalert-version.md), merged through
  [PR #75](https://github.com/healdropper/terraform-provider-goalert/pull/75), [PR #76](https://github.com/healdropper/terraform-provider-goalert/pull/76), [PR #77](https://github.com/healdropper/terraform-provider-goalert/pull/77), and [PR #78](https://github.com/healdropper/terraform-provider-goalert/pull/78) on 2026-10-01.
  Resources: `goalert_label`, extended `goalert_escalation_policy` (`multi_ack`), extended `goalert_user_contact_method` (`enable_status_updates`, `private`, `status_updates`).
  Closed on GitHub.

## Active milestone

- **Registry Publication and GA**: [registry publication and GA](milestones/registry-publication-and-ga.md).
  - Scope: Privacy and git-history sanitization, public repository governance (`CODEOWNERS`, `CONTRIBUTING.md`, `SECURITY.md`, fork PR Actions gating `all_external_contributors`, `main` branch protection ruleset), Registry documentation validation, GPG-signed `v1.0.0` release, Terraform Registry publication, and production consumer adoption (GoAlert v0.35.0 upgrade).
  - Tracking: [GitHub Milestone 10](https://github.com/healdropper/terraform-provider-goalert/milestone/10).

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

[Milestone: Registry Publication and GA](https://github.com/healdropper/terraform-provider-goalert/milestone/10)
is the active delivery planning container.
Its canonical record links readiness and dependencies across
DISC, SPEC, IMPL, and VERIFY gates.
