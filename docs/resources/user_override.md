# goalert_user_override (Resource)

Manages temporary shift coverage or replacement on a GoAlert schedule.

## Example Usage

```terraform
resource "goalert_user_override" "vacation_swap" {
  schedule_id    = goalert_schedule.engineering.id
  start_time     = "2026-10-05T00:00:00Z"
  end_time       = "2026-10-12T00:00:00Z"
  add_user_id    = goalert_user.bob.id
  remove_user_id = goalert_user.alice.id
}
```

## Schema

### Required

- `end_time` (String) RFC3339 / ISO-8601 timestamp when override expires.
- `schedule_id` (String) UUID of the target schedule. Forces replacement.
- `start_time` (String) RFC3339 / ISO-8601 timestamp when override starts.

### Optional

- `add_user_id` (String) UUID of the user taking on coverage.
- `remove_user_id` (String) UUID of the user being replaced.

At least one of `add_user_id` or `remove_user_id` must be set.

### Read-Only

- `id` (String) UUID of the user override.

## Import

Import is supported using the override UUID:

```bash
terraform import goalert_user_override.vacation_swap 2dbf9d46-417c-46cd-b840-40fd674932d7
```
