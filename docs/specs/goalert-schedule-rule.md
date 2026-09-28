# Specification: goalert_schedule_rule Resource Contract

Status: Accepted contract for Milestone `Schedules and User Overrides` (Issue #53 / Gate SCHED-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/schedules-feasibility.md) against GoAlert v0.35.0.

## 1. Resource Identity and Scope

The `goalert_schedule_rule` resource defines a recurring coverage window on a `goalert_schedule`, binding either a rotation or an individual user to specified times of day and days of the week.

In GoAlert, target rules are keyed by `(scheduleID, targetType, targetID)`.

## 2. Schema Specification

| Attribute | Type | Requirement | Constraints | Description |
| --- | --- | --- | --- | --- |
| `id` | String | Computed | `<schedule_id>:<target_type>:<target_id>` | Compound identifier. |
| `schedule_id` | String | Required | UUID, RequiresReplace | Schedule UUID to attach the rule to. |
| `target_type` | String | Required | One of: `rotation`, `user`, RequiresReplace | Target entity type. |
| `target_id` | String | Required | UUID, RequiresReplace | Target entity UUID (rotation or user). |
| `start_time` | String | Required | ClockTime `HH:MM` (24h) | Beginning of active shift window (inclusive). |
| `end_time` | String | Required | ClockTime `HH:MM` (24h) | End of active shift window (inclusive). |
| `weekday_filter` | List(Bool) | Optional, Computed | Exactly 7 booleans (Sun..Sat) | Active days of the week. Defaults to `[true, true, true, true, true, true, true]`. |

## 3. Lifecycle and Mutation Rules

1. **Creation (`Create`)**:
   - Executes GraphQL mutation `updateScheduleTarget(input: ScheduleTargetInput!)`.
   - Passes `scheduleID`, `target: { id: target_id, type: target_type }`, and `rules: [{ start: start_time, end: end_time, weekdayFilter: weekday_filter }]`.
2. **Read (`Read`)**:
   - Executes GraphQL query `schedule(id: $schedule_id)`.
   - Locates target matching `target_id` and `target_type`.
   - Extracts active rule parameters (`start`, `end`, `weekdayFilter`).
   - If schedule or target rule is not found, removes resource from state.
3. **Update (`Update`)**:
   - `start_time`, `end_time`, and `weekday_filter` are mutable in-place.
   - Executes `updateScheduleTarget`.
4. **Delete (`Delete`)**:
   - Executes `updateScheduleTarget` with `rules: []`, cleanly removing the target assignment from the schedule.

## 4. Import Contract

Schedule rules are imported using their compound key `<schedule_id>:<target_type>:<target_id>`:
```bash
terraform import goalert_schedule_rule.weekday 98b3d474-9ae5-4acd-b3ea-dbf6e1ec308c:rotation:f7783057-5542-4bc2-b4fe-a202e13a25c9
```
