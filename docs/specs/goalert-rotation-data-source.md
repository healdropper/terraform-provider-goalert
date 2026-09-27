# Specification: goalert_rotation Data Source Contract

Status: Accepted contract for Milestone `Rotations and Escalation Targets` (Issue #44 / Gate ROT-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/rotation-feasibility.md).

## 1. Scope and Criteria

The `data.goalert_rotation` data source resolves an existing GoAlert rotation by unique ID or exact name.

## 2. Schema Specification

### Query Arguments
- Exactly one of `id` or `name` must be specified:
  - `id` (String, Optional, Computed): UUID of the rotation.
  - `name` (String, Optional, Computed): Exact name of the rotation to search.

### Exported Attributes
- `id` (String): Rotation UUID.
- `name` (String): Rotation name.
- `description` (String): Rotation description.
- `type` (String): Rotation type (`daily`, `weekly`, `hourly`).
- `start_time` (String): Anchor start timestamp.
- `time_zone` (String): IANA time zone.
- `shift_length` (Int64): Duration multiplier.
- `user_ids` (List of String): Ordered list of participant user IDs.
- `active_user_index` (Int64): Zero-based index of the currently active participant.

## 3. Resolution Logic

1. When `id` is specified:
   - Execute query `rotation(id: $id)`.
   - If not found, return diagnostic error.
2. When `name` is specified:
   - Execute query `rotations(search: $name, first: 50)`.
   - Filter results for exact string match on `name`.
   - Return error if 0 or >1 matches are found.
