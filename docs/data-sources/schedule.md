# goalert_schedule (Data Source)

Fetches an existing GoAlert schedule by its unique UUID or exact name.

## Example Usage

```terraform
# Look up schedule by name
data "goalert_schedule" "engineering" {
  name = "Engineering On-Call"
}

# Look up schedule by ID
data "goalert_schedule" "by_id" {
  id = "98b3d474-9ae5-4acd-b3ea-dbf6e1ec308c"
}
```

## Schema

### Optional

- `id` (String) UUID of the schedule. Exactly one of `id` or `name` must be set.
- `name` (String) Name of the schedule to search for. Exactly one of `id` or `name` must be set.

### Read-Only

- `description` (String) Description of the schedule.
- `time_zone` (String) Configured IANA timezone.
