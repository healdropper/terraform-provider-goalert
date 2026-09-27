# Specification: goalert_rotation Resource Contract

Status: Accepted contract for Milestone `Rotations and Escalation Targets` (Issue #44 / Gate ROT-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/rotation-feasibility.md) against GoAlert v0.35.0.

## 1. Resource Identity and Scope

The `goalert_rotation` resource manages the lifecycle of on-call shift rotations in GoAlert. A rotation defines an ordered repeating sequence of participants (`user_ids`) taking turns according to a schedule frequency (`daily`, `weekly`, `hourly`) and shift length.

## 2. Schema Specification

| Attribute | Type | Requirement | Constraints | Description |
| --- | --- | --- | --- | --- |
| `id` | String | Computed | UUID | Unique identifier of the rotation. |
| `name` | String | Required | `1 <= len <= 255`, regex `^[a-zA-Z0-9\-_' ]+$` | Human-readable name of the rotation. |
| `description` | String | Optional, Computed | Default `""` | Optional description of the rotation purpose. |
| `type` | String | Required | One of: `daily`, `weekly`, `hourly` | Frequency of shift handoffs. |
| `start_time` | String | Required | RFC3339 / ISO-8601 timestamp | Rotation anchor time determining when handoffs occur. |
| `time_zone` | String | Required | Valid IANA timezone | Timezone applied for shift calculation (e.g. `Europe/Madrid`, `UTC`). |
| `shift_length` | Int64 | Optional, Computed | Min `1`, default `1` | Multiplier applied to `type` (e.g. 2 for 2-week or 2-day shifts). |
| `user_ids` | List(String) | Required | Non-empty list of User UUIDs | Ordered sequence of participating operators. |
| `active_user_index` | Int64 | Computed | Zero-indexed integer | Current active index in `user_ids`. |

## 3. Lifecycle and Mutation Rules

1. **Creation (`Create`)**:
   - Executes GraphQL mutation `createRotation(input: CreateRotationInput!)`.
   - Passes `name`, `description`, `type`, `start`, `timeZone`, `shiftLength`, `userIDs`.
2. **Read (`Read`)**:
   - Executes GraphQL query `rotation(id: $id)`.
   - Synchronizes `name`, `description`, `type`, `start`, `timeZone`, `shiftLength`, `userIDs`, `activeUserIndex`.
   - If the rotation is not found, marks state as removed (`RemoveResource`).
3. **Update (`Update`)**:
   - All attributes are mutable in-place.
   - Executes GraphQL mutation `updateRotation(input: UpdateRotationInput!)`.
   - Reordering `user_ids` updates participant succession immediately without resource destruction.
4. **Delete (`Delete`)**:
   - Executes GraphQL mutation `deleteAll(input: [{type: rotation, id: $id}])`.

## 4. Import Contract

Rotations are imported by their UUID:
```bash
terraform import goalert_rotation.primary 55c3e4b3-2987-4635-9f60-c6ffc73672b7
```
