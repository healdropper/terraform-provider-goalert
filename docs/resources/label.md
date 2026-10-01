# goalert_label (Resource)

Manages a key-value label attached to a GoAlert service, escalation policy, schedule, or rotation (GoAlert v0.35.0+).

## Example Usage

```terraform
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

## Schema

### Required

- `key` (String) Label key in `<domain>/<suffix>` format (e.g. `example.com/team`). Changing this forces replacement.
- `target_id` (String) UUID of the target GoAlert resource. Changing this forces replacement.
- `target_type` (String) Type of GoAlert target resource (`service`, `escalation_policy`, `schedule`, or `rotation`). Changing this forces replacement.
- `value` (String) Label value (3 to 255 printable characters).

### Read-Only

- `id` (String) Compound identifier in the format `<target_type>:<target_id>/<key>`.

## Import

Import is supported using the compound ID `<target_type>:<target_id>/<key>`:

```bash
terraform import goalert_label.policy_team escalation_policy:98b3d474-9ae5-4acd-b3ea-dbf6e1ec308c/example.com/team
```
