---
page_title: "goalert_rotation Resource"
description: "Manages an on-call shift rotation in GoAlert."
---

# goalert_rotation (Resource)

Manages an on-call shift rotation in GoAlert.

## Example Usage

```terraform
resource "goalert_user" "sre1" {
  name  = "Alex SRE"
  email = "alex@example.com"
  role  = "user"
}

resource "goalert_user" "sre2" {
  name  = "Sam SRE"
  email = "sam@example.com"
  role  = "user"
}

resource "goalert_rotation" "primary" {
  name         = "Primary SRE On-Call"
  description  = "Weekly primary on-call rotation"
  type         = "weekly"
  start_time   = "2026-10-01T09:00:00Z"
  time_zone    = "Europe/Madrid"
  shift_length = 1
  user_ids     = [
    goalert_user.sre1.id,
    goalert_user.sre2.id,
  ]
}
```

## Schema

### Required

- `name` (String) Name of the rotation. Can contain letters, digits, hyphens, underscores, apostrophes, and spaces.
- `start_time` (String) RFC3339 / ISO-8601 start timestamp.
- `time_zone` (String) IANA timezone string (e.g. `Europe/Madrid`, `UTC`).
- `type` (String) Shift frequency (`daily`, `weekly`, `hourly`).
- `user_ids` (List of String) Ordered list of participating user UUIDs.

### Optional

- `description` (String) Description of the rotation. Default is `""`.
- `shift_length` (Number) Duration multiplier for the rotation type. Minimum `1`, default `1`.

### Read-Only

- `active_user_index` (Number) Current active participant index in `user_ids`.
- `id` (String) Unique identifier of the rotation.

## Import

Import is supported using the rotation UUID:

```bash
terraform import goalert_rotation.primary 55c3e4b3-2987-4635-9f60-c6ffc73672b7
```
