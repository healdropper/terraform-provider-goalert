# goalert_schedule_rule (Resource)

Manages an active target coverage rule on a GoAlert schedule. Binds a rotation or user to specific shift hours and weekdays.

## Example Usage

```terraform
resource "goalert_schedule" "engineering" {
  name      = "Engineering On-Call"
  time_zone = "Europe/Madrid"
}

resource "goalert_rotation" "primary" {
  name       = "Primary SRE"
  type       = "daily"
  start_time = "2026-10-01T08:00:00Z"
  time_zone  = "Europe/Madrid"
  user_ids   = [goalert_user.alice.id, goalert_user.bob.id]
}

resource "goalert_schedule_rule" "weekday_coverage" {
  schedule_id    = goalert_schedule.engineering.id
  target_type    = "rotation"
  target_id      = goalert_rotation.primary.id
  start_time     = "09:00"
  end_time       = "17:00"
  weekday_filter = [false, true, true, true, true, true, false] # Mon-Fri
}
```

## Schema

### Required

- `schedule_id` (String) UUID of the schedule.
- `target_type` (String) Target entity type (`rotation` or `user`). Forces replacement.
- `target_id` (String) UUID of the target entity. Forces replacement.
- `start_time` (String) Shift start clock time in 24-hour `HH:MM` format.
- `end_time` (String) Shift end clock time in 24-hour `HH:MM` format.

### Optional

- `weekday_filter` (List of Boolean) 7-element boolean array for `[Sun, Mon, Tue, Wed, Thu, Fri, Sat]`. Defaults to all `true`.

### Read-Only

- `id` (String) Compound identifier `<schedule_id>:<target_type>:<target_id>`.

## Import

Import is supported using the compound key:

```bash
terraform import goalert_schedule_rule.weekday_coverage 98b3d474-9ae5-4acd-b3ea-dbf6e1ec308c:rotation:f7783057-5542-4bc2-b4fe-a202e13a25c9
```
