---
page_title: "goalert_heartbeat_monitor Resource"
description: "Manage a GoAlert dead-man switch heartbeat monitor for a service."
---

# goalert_heartbeat_monitor

Manages a heartbeat monitor on a GoAlert service. Heartbeat monitors act as a "dead-man switch":
if an external worker, backup process, or cron job fails to send an HTTP GET or POST request
to the monitor's ping endpoint (`href`) within the specified timeout duration, an alert
is triggered on the service.

## Example

```hcl
resource "goalert_service" "app" {
  name                 = "Batch Processing Worker"
  escalation_policy_id = goalert_escalation_policy.primary.id
}

resource "goalert_heartbeat_monitor" "nightly_backup" {
  service_id      = goalert_service.app.id
  name            = "Nightly Database Backup"
  timeout_minutes = 15
}

output "backup_ping_url" {
  value     = goalert_heartbeat_monitor.nightly_backup.href
  sensitive = true
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Computed heartbeat monitor UUID |
| `service_id` | Required UUID of the parent GoAlert service. Requires replacement if changed |
| `name` | Required string. Modifiable in place |
| `timeout_minutes` | Required integer (minimum 5). Timeout window before triggering an alert. Modifiable in place |
| `href` | Computed string, sensitive. The HTTP ping URL used by external jobs to send keep-alive signals |

## Import

```console
terraform import goalert_heartbeat_monitor.nightly_backup 11111111-1111-4111-8111-111111111111
```
