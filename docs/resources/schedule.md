---
page_title: "goalert_schedule Resource"
description: "Manages an on-call schedule in GoAlert."
---

# goalert_schedule (Resource)

Manages an on-call schedule in GoAlert.

## Example Usage

```terraform
resource "goalert_schedule" "engineering" {
  name        = "Engineering On-Call"
  description = "Primary engineering shifts and coverage calendar"
  time_zone   = "Europe/Madrid"
}
```

## Schema

### Required

- `name` (String) Name of the schedule.
- `time_zone` (String) IANA timezone string (e.g. `Europe/Madrid`, `UTC`).

### Optional

- `description` (String) Description of the schedule. Default is `""`.

### Read-Only

- `id` (String) Unique identifier of the schedule.

## Import

Import is supported using the schedule UUID:

```bash
terraform import goalert_schedule.engineering 98b3d474-9ae5-4acd-b3ea-dbf6e1ec308c
```
