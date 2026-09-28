# Specification: data.goalert_schedule Data Source Contract

Status: Accepted contract for Milestone `Schedules and User Overrides` (Issue #53 / Gate SCHED-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/schedules-feasibility.md) against GoAlert v0.35.0.

## 1. Data Source Identity and Scope

The `goalert_schedule` data source enables Terraform configurations to query and reference existing GoAlert schedules by their unique UUID or exact name.

## 2. Schema Specification

| Attribute | Type | Requirement | Constraints | Description |
| --- | --- | --- | --- | --- |
| `id` | String | Optional, Computed | UUID | UUID of the schedule. Required if `name` is omitted. |
| `name` | String | Optional, Computed | Exact match | Name of the schedule. Required if `id` is omitted. |
| `description` | String | Computed | | Description of the schedule. |
| `time_zone` | String | Computed | | Configured IANA timezone of the schedule. |

Validation rule: Exactly one of `id` or `name` must be specified.

## 3. Query Logic

1. When `id` is specified:
   - Executes `schedule(id: $id)` query.
2. When `name` is specified:
   - Executes `schedules(input: { search: $name, first: 50 })` query.
   - Filters results for an exact case-sensitive match on `node.name == name`.
   - If zero matches found, returns error with diagnostic advice.
   - If multiple exact matches found, returns error with diagnostic advice to use `id`.
