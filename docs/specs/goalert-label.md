# Polymorphic label resource contract

Owner: healdropper.
Status: Accepted specification for Milestone `Embrace Latest GoAlert Version` under Gate UPST-SPEC ([Issue #71](https://github.com/healdropper/terraform-provider-goalert/issues/71)).
Recorded: 2026-10-01.
Related: [Issue #71](https://github.com/healdropper/terraform-provider-goalert/issues/71),
[Milestone Embrace Latest GoAlert Version](../milestones/embrace-latest-goalert-version.md),
[Feasibility research](../research/goalert-v035-capabilities-feasibility.md).

## Requirements and scope

| ID | Requirement | Specification |
| --- | --- | --- |
| PLBL-01 | Generic Plugin Framework resource | Resource type `goalert_label` implemented with `hashicorp/terraform-plugin-framework`. |
| PLBL-02 | Resource attributes | `id` (Computed String: `<target_type>:<target_id>/<key>`), `target_type` (Required String, RequiresReplace), `target_id` (Required String UUID, RequiresReplace), `key` (Required String, RequiresReplace), `value` (Required String). |
| PLBL-03 | Target type validation | `target_type` must be one of `"service"`, `"escalation_policy"`, `"schedule"`, or `"rotation"` (mapped to GoAlert `TargetType`: `service`, `escalationPolicy`, `schedule`, `rotation`). |
| PLBL-04 | Key & value validation | `key` must be 3-255 characters in `<domain>/<suffix>` format. `value` must be 3-255 printable characters with no leading/trailing whitespace. |
| PLBL-05 | In-place update | Modifying `value` executes `setLabel(input: {target: {type, id}, key, value})` in place. |
| PLBL-06 | Deletion lifecycle | Deletion sets `value: ""` in `setLabel`, which removes the label from the target entity in GoAlert. |
| PLBL-07 | Drift detection | Read queries the target entity's `labels { key value }` list. If the target or key is missing, the resource is removed from state cleanly (`resp.State.RemoveResource(ctx)`). |
| PLBL-08 | Import syntax | Supports compound ID import: `terraform import goalert_label.example <target_type>:<target_id>/<key>`. |

## Resource schema

```hcl
resource "goalert_label" "policy_team" {
  target_type = "escalation_policy"
  target_id   = goalert_escalation_policy.primary.id
  key         = "example.com/team"
  value       = "sre-core"
}

resource "goalert_label" "schedule_tier" {
  target_type = "schedule"
  target_id   = goalert_schedule.engineering.id
  key         = "example.com/tier"
  value       = "tier-1"
}

resource "goalert_label" "rotation_region" {
  target_type = "rotation"
  target_id   = goalert_rotation.primary.id
  key         = "example.com/region"
  value       = "eu-west"
}
```
