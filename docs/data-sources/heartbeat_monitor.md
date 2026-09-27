---
page_title: "goalert_heartbeat_monitor Data Source"
description: "Lookup a GoAlert heartbeat monitor by its ID."
---

# goalert_heartbeat_monitor (Data Source)

Fetches details of an existing GoAlert heartbeat monitor by UUID.

## Example

```hcl
data "goalert_heartbeat_monitor" "backup" {
  id = "f98e1234-5678-9abc-def0-123456789abc"
}

output "ping_url" {
  value     = data.goalert_heartbeat_monitor.backup.href
  sensitive = true
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Required heartbeat monitor UUID |
| `service_id` | Computed UUID of the parent GoAlert service |
| `name` | Computed name of the heartbeat monitor |
| `timeout_minutes` | Computed timeout window in minutes |
| `href` | Computed sensitive ping URL |
