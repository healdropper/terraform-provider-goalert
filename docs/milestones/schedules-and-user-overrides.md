# Schedules and User Overrides

Owner: healdropper. Recorded: 2026-09-28.
Status: Planned and tracked under [GitHub Milestone 7](https://github.com/healdropper/terraform-provider-goalert/milestone/7).

## Recorded owner decisions

- Implement schedule management via `goalert_schedule` (name, description, time_zone).
- Support schedule target assignment and time filters via `goalert_schedule_rule` (schedule_id, start_time, end_time, weekday_filter, user_id, rotation_id).
- Support temporary human shift replacements via `goalert_user_override` (schedule_id, start_time, end_time, add_user_id, remove_user_id).
- Support schedule on-call notification rules via `goalert_schedule_on_call_notification_rule` (schedule_id, time, weekday_filter, channel_id / contact_method_id).
- Provide data source `data.goalert_schedule` for querying schedules by ID or exact name.
- Expand `goalert_escalation_policy` steps to support `schedule_ids` alongside `user_ids`, `rotation_ids`, and `webhook_action`.

## Planning envelope

Proposed outcomes:
1. Declare schedules with names, descriptions, and designated timezones.
2. Define recurring schedule rules that assign rotations or individual users to specific daytime windows and days of the week.
3. Schedule planned user overrides (vacations, coverage swaps) with start and end timestamps.
4. Configure on-call reminder/notification rules triggered when an on-call shift begins or is upcoming.
5. Query existing schedule records via data source.
6. Target schedules directly in escalation policy steps (`schedule_ids`).
7. Verify complete lifecycle via disposable GoAlert v0.35.0 acceptance tests.

| ID | Planning/acceptance gate | Evidence required before closing |
| --- | --- | --- |
| SCHED-DISC | Establish schedule, rule, override, and on-call notification API feasibility | Probe suite against disposable GoAlert v0.35.0 characterizing GraphQL mutations and queries, input shapes, validation limits, and escalation policy step integration |
| SCHED-SPEC | Resolve and accept resource and data source contracts | Canonical specifications in `docs/specs/` defining schemas, validations, compound import syntax, and immutability rules; matching documentation in `docs/resources/` and `docs/data-sources/` |
| SCHED-IMPL | Implement accepted contracts | Provider resources (`goalert_schedule`, `goalert_schedule_rule`, `goalert_user_override`, `goalert_schedule_on_call_notification_rule`), data source `data.goalert_schedule`, expanded escalation policy steps, and typed client methods |
| SCHED-VERIFY | Establish provider lifecycle and acceptance confidence | Unit tests and real disposable GoAlert acceptance tests in `scripts/acceptance.py` verifying full CRUD, time window filters, override application, import, and drift repair |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| SCHED-DISC | [#52: feat: investigate schedule, rule, override, and on-call notification API feasibility](https://github.com/healdropper/terraform-provider-goalert/issues/52) | In progress |
| SCHED-SPEC | [#53: docs: resolve schedule, rule, override, and on-call notification contracts](https://github.com/healdropper/terraform-provider-goalert/issues/53) | Pending DISC |
| SCHED-IMPL | [#54: feat: implement schedule, rule, override resources, notification rules, and data sources](https://github.com/healdropper/terraform-provider-goalert/issues/54) | Pending SPEC |
| SCHED-VERIFY | [#55: test: verify schedule, rule, override, and notification rule lifecycles](https://github.com/healdropper/terraform-provider-goalert/issues/55) | Pending IMPL |
