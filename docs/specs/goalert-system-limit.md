# System Limit Resource Contract

Owner: healdropper.
Status: Accepted specification for Milestone `Collaboration Channels and System Limits` under Gate COL-SPEC (Issue #62).
Recorded: 2026-09-28.
Related: [Issue #62](https://github.com/healdropper/terraform-provider-goalert/issues/62),
[Milestone Collaboration Channels and System Limits](../milestones/collaboration-channels-and-system-limits.md),
[Feasibility research](../research/collaboration-and-limits-feasibility.md).

## Requirements and scope

| ID | Requirement | Specification |
| --- | --- | --- |
| SL-01 | Generic Plugin Framework resource | Resource type `goalert_system_limit` implemented with `hashicorp/terraform-plugin-framework`. |
| SL-02 | Resource attributes | `id` (Required String, RequiresReplace), `value` (Required Int64), `description` (Computed String). |
| SL-03 | SystemLimitID enum validation | `id` must be one of the valid GoAlert `SystemLimitID` enum values (e.g., `RulesPerSchedule`, `EPActionsPerStep`, `EPStepsPerPolicy`, etc.). |
| SL-04 | Non-negative threshold | `value` must be >= 0 (`int64validator.AtLeast(0)`). |
| SL-05 | In-place update | `value` supports in-place update via mutation `setSystemLimits(input: [{id: $id, value: $value}])`. `id` modification requires replacement. |
| SL-06 | Read and drift detection | Query `systemLimits`, filter by `id`. If the limit is not found in the list, remove the resource from state (`resp.State.RemoveResource(ctx)`). |
| SL-07 | Lifecycle & deletion | Deletion in Terraform removes the resource from state (no-op delete), because GoAlert system limits are permanent platform configuration objects that cannot be destroyed via GraphQL. |
| SL-08 | Import syntax | Supports import by ID: `terraform import goalert_system_limit.example <limit_id>` (e.g. `RulesPerSchedule`). |

## Resource schema

```hcl
resource "goalert_system_limit" "rules_per_schedule" {
  id    = "RulesPerSchedule"
  value = 50
}
```

### Attribute definitions

- `id` (String, Required): The system limit identifier matching GoAlert's `SystemLimitID` enum. Requires replacement if changed.
- `value` (Int64, Required): The threshold value assigned to this system limit. Must be greater than or equal to 0. Modifiable in place.
- `description` (String, Computed): Human-readable description of the limit provided by GoAlert.
