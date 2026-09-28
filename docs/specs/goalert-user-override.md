# Specification: goalert_user_override Resource Contract

Status: Accepted contract for Milestone `Schedules and User Overrides` (Issue #53 / Gate SCHED-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/schedules-feasibility.md) against GoAlert v0.35.0.

## 1. Resource Identity and Scope

The `goalert_user_override` resource manages temporary shift coverage on a schedule (e.g. planned vacations, swap coverage, or emergency substitutions).

## 2. Schema Specification

| Attribute | Type | Requirement | Constraints | Description |
| --- | --- | --- | --- | --- |
| `id` | String | Computed | UUID | Unique identifier of the override. |
| `schedule_id` | String | Required | UUID, RequiresReplace | Schedule UUID where the override applies. |
| `start_time` | String | Required | RFC3339 / ISO-8601 | Start timestamp of coverage override. |
| `end_time` | String | Required | RFC3339 / ISO-8601 | End timestamp of coverage override. |
| `add_user_id` | String | Optional | UUID | User ID taking on-call coverage. |
| `remove_user_id` | String | Optional | UUID | User ID being replaced. |

Validation rule: At least one of `add_user_id` or `remove_user_id` must be provided.

## 3. Lifecycle and Mutation Rules

1. **Creation (`Create`)**:
   - Executes GraphQL mutation `createUserOverride(input: CreateUserOverrideInput!)`.
   - Passes `scheduleID`, `start`, `end`, and either/both `addUserID`, `removeUserID`.
2. **Read (`Read`)**:
   - Executes GraphQL query `userOverride(id: $id)`.
   - Synchronizes `start`, `end`, `addUserID`, `removeUserID`, `schedule_id`.
   - If not found, removes from state.
3. **Update (`Update`)**:
   - `start_time`, `end_time`, `add_user_id`, `remove_user_id` are mutable in-place.
   - Executes GraphQL mutation `updateUserOverride(input: UpdateUserOverrideInput!)`.
4. **Delete (`Delete`)**:
   - Executes GraphQL mutation `deleteAll(input: [{type: userOverride, id: $id}])`.

## 4. Import Contract

User overrides are imported by their UUID:
```bash
terraform import goalert_user_override.swap 2dbf9d46-417c-46cd-b840-40fd674932d7
```
