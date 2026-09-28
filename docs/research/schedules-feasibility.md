# Schedules, Overrides, and Notification Rules Feasibility Probe

Status: Complete empirical evidence for Gate SCHED-DISC (Issue #52).
Executed: 2026-09-28 against disposable GoAlert v0.35.0 (`scripts/test_schedule_feasibility.py`).

## 1. Schedule GraphQL Schema Characterization

### Queries
- **`schedule(id: ID!)`**:
  - `id: ID!`
  - `name: String!`
  - `description: String!`
  - `timeZone: String!`
  - `isFavorite: Boolean!`
  - `targets: [ScheduleTarget!]!`
    - `target: Target!` (`id`, `name`, `type`)
    - `rules: [ScheduleRule!]!` (`id`, `start`, `end`, `weekdayFilter`)
- **`schedules(input: ScheduleSearchOptions)`**:
  - Arguments: `input: { search: String, first: Int }`
  - Returns connection with `nodes: [Schedule!]!`

### Mutations
- **`createSchedule(input: CreateScheduleInput!): Schedule!`**:
  - Required: `name: String!`, `timeZone: String!`
  - Optional: `description: String`, `favorite: Boolean`, `targets: [ScheduleTargetInput!]`
- **`updateSchedule(input: UpdateScheduleInput!): Boolean!`**:
  - `id: ID!` (Required)
  - `name: String`, `description: String`, `timeZone: String` (in-place updates)
- **`deleteAll(input: [{id: ID!, type: schedule}])`**: Clean deletion.

## 2. Schedule Target and Rules Characterization

In GoAlert, a schedule's active coverage is defined by **Target Rules**:
- Mutation: **`updateScheduleTarget(input: ScheduleTargetInput!): Boolean!`**
  - `scheduleID: ID!`
  - `target: TargetInput!` (e.g. `{ id: <rotation_id>, type: "rotation" }` or `{ id: <user_id>, type: "user" }`)
  - `rules: [ScheduleRuleInput!]!`
    - `start: ClockTime!` (Format `HH:MM`, 24-hour, e.g. `"09:00"`)
    - `end: ClockTime!` (Format `HH:MM`, 24-hour, e.g. `"17:00"`)
    - `weekdayFilter: WeekdayFilter!` (Array of 7 booleans: `[Sun, Mon, Tue, Wed, Thu, Fri, Sat]`)
- **Target Deletion**: Calling `updateScheduleTarget` with `rules: []` removes the target assignment from the schedule.

## 3. User Override Characterization

Temporary shift coverage or human replacements on a schedule:
- **`createUserOverride(input: CreateUserOverrideInput!): UserOverride!`**:
  - `scheduleID: ID!` (direct attribute)
  - `start: ISOTimestamp!` (ISO-8601 UTC timestamp)
  - `end: ISOTimestamp!` (ISO-8601 UTC timestamp)
  - `addUserID: ID` (Operator covering the shift)
  - `removeUserID: ID` (Operator being replaced)
  - At least one of `addUserID` or `removeUserID` is required.
- **`updateUserOverride(input: UpdateUserOverrideInput!): Boolean!`**:
  - `id: ID!` (UUID)
  - `start`, `end`, `addUserID`, `removeUserID`
- **`userOverride(id: ID!): UserOverride`**: Query by UUID.
- **`deleteAll(input: [{id: ID!, type: userOverride}])`**: Deletes override.

## 4. Schedule On-Call Notification Rules

- **`setScheduleOnCallNotificationRules(input: SetScheduleOnCallNotificationRulesInput!): Boolean!`**:
  - `scheduleID: ID!`
  - `rules: [OnCallNotificationRuleInput!]!`
    - `time: ClockTime` (Time of day for notification)
    - `weekdayFilter: WeekdayFilter` (Days of week)
    - `dest: DestinationInput` (e.g. contact method or slack channel)

## 5. Escalation Policy Step Integration

Escalation policy steps natively support scheduling destinations:
- **`DestinationInput`**:
  - `type: "builtin-schedule"`
  - `args: { "schedule_id": "<UUID>" }`
- Verified empirical result: GoAlert registers `builtin-schedule` in `step.actions` alongside `builtin-user`, `builtin-rotation`, and `builtin-webhook`.

## 6. Terraform Modeling Architecture

1. **`goalert_schedule`**:
   - `id`: Computed String (UUID)
   - `name`: Required String (`^[a-zA-Z0-9\-_' ]+$`)
   - `description`: Optional String (default `""`)
   - `time_zone`: Required String (IANA timezone)
   - Import: `<schedule_id>`

2. **`goalert_schedule_rule`**:
   - `id`: Computed String (`<schedule_id>:<target_type>:<target_id>`)
   - `schedule_id`: Required String (UUID, RequiresReplace)
   - `target_type`: Required String (`rotation` or `user`, RequiresReplace)
   - `target_id`: Required String (UUID, RequiresReplace)
   - `start_time`: Required String (ClockTime `HH:MM`)
   - `end_time`: Required String (ClockTime `HH:MM`)
   - `weekday_filter`: Optional/Computed List of Boolean (default all true `[true, true, true, true, true, true, true]`)

3. **`goalert_user_override`**:
   - `id`: Computed String (UUID)
   - `schedule_id`: Required String (UUID, RequiresReplace)
   - `start_time`: Required String (ISO-8601)
   - `end_time`: Required String (ISO-8601)
   - `add_user_id`: Optional String (UUID)
   - `remove_user_id`: Optional String (UUID)
   - Import: `<override_id>`

4. **`data.goalert_schedule`**:
   - Query by `id` or exact `name`.
   - Exports `name`, `description`, `time_zone`.

5. **`goalert_escalation_policy` Step Expansion**:
   - Add `schedule_ids` (List of String, Optional) to `step` block.
   - Maps to `builtin-schedule` destination action.
