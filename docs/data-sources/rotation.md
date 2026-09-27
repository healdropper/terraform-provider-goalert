# goalert_rotation (Data Source)

Lookup an existing on-call shift rotation in GoAlert by unique ID or exact name.

## Example Usage

```terraform
data "goalert_rotation" "by_name" {
  name = "Primary SRE On-Call"
}

output "primary_participants" {
  value = data.goalert_rotation.by_name.user_ids
}
```

## Schema

### Optional

- `id` (String) UUID of the rotation. Exactly one of `id` or `name` must be set.
- `name` (String) Exact name of the rotation to search.

### Read-Only

- `active_user_index` (Number) Current active participant index in `user_ids`.
- `description` (String) Rotation description.
- `shift_length` (Number) Duration multiplier.
- `start_time` (String) RFC3339 start timestamp.
- `time_zone` (String) IANA timezone string.
- `type` (String) Shift frequency (`daily`, `weekly`, `hourly`).
- `user_ids` (List of String) Ordered list of participating user UUIDs.
