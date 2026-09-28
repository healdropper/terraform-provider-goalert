# Specification: goalert_schedule Resource Contract

Status: Accepted contract for Milestone `Schedules and User Overrides` (Issue #53 / Gate SCHED-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/schedules-feasibility.md) against GoAlert v0.35.0.

## 1. Resource Identity and Scope

The `goalert_schedule` resource manages the lifecycle of on-call schedules in GoAlert. A schedule acts as a calendar container binding target rotations or individual users to specific daytime windows and days of the week, as well as holding planned human user overrides.

## 2. Schema Specification

| Attribute | Type | Requirement | Constraints | Description |
| --- | --- | --- | --- | --- |
| `id` | String | Computed | UUID | Unique identifier of the schedule. |
| `name` | String | Required | `1 <= len <= 255`, regex `^[a-zA-Z0-9\-_' ]+$` | Human-readable name of the schedule. |
| `description` | String | Optional, Computed | Default `""` | Optional description of the schedule. |
| `time_zone` | String | Required | Valid IANA timezone | Timezone for daylight savings, shifts, and rule evaluations. |

## 3. Lifecycle and Mutation Rules

1. **Creation (`Create`)**:
   - Executes GraphQL mutation `createSchedule(input: CreateScheduleInput!)`.
   - Passes `name`, `description`, `timeZone`.
2. **Read (`Read`)**:
   - Executes GraphQL query `schedule(id: $id)`.
   - Synchronizes `name`, `description`, `timeZone`.
   - If not found, removes resource from state.
3. **Update (`Update`)**:
   - All attributes (`name`, `description`, `time_zone`) are mutable in-place.
   - Executes GraphQL mutation `updateSchedule(input: UpdateScheduleInput!)`.
4. **Delete (`Delete`)**:
   - Executes GraphQL mutation `deleteAll(input: [{type: schedule, id: $id}])`.

## 4. Import Contract

Schedules are imported by their UUID:
```bash
terraform import goalert_schedule.primary 98b3d474-9ae5-4acd-b3ea-dbf6e1ec308c
```
