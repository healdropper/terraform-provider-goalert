# goalert_system_limit (Resource)

Manages a global system limit configuration in GoAlert.

## Example Usage

```terraform
resource "goalert_system_limit" "rules_per_schedule" {
  id    = "RulesPerSchedule"
  value = 50
}

resource "goalert_system_limit" "ep_actions_per_step" {
  id    = "EPActionsPerStep"
  value = 25
}
```

## Schema

### Required

- `id` (String) Identifier of the system limit (must match a valid GoAlert `SystemLimitID`, e.g. `RulesPerSchedule`, `EPActionsPerStep`, `EPStepsPerPolicy`, etc.). Requires replacement if changed.
- `value` (Number) Configured threshold for this system limit. Must be greater than or equal to 0.

### Read-Only

- `description` (String) Description of the limit returned by GoAlert.

## Import

Import is supported using the limit ID:

```bash
terraform import goalert_system_limit.rules_per_schedule RulesPerSchedule
```
