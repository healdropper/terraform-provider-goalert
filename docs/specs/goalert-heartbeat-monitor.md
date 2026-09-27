# Heartbeat monitor resource contract

Owner: healdropper.
Status: Accepted specification for v0.0.4 under V004-SPEC.
Recorded: 2026-09-27.
Related: [Issue #32](https://github.com/healdropper/terraform-provider-goalert/issues/32),
[Milestone Heartbeat Monitors and Service Labels](../milestones/heartbeat-monitors-and-service-labels.md),
[Feasibility research](../research/heartbeat-feasibility.md).

## Requirements and scope

| ID | Requirement | Specification |
| --- | --- | --- |
| HB-01 | Generic Plugin Framework resource | Resource type `goalert_heartbeat_monitor` implemented with `hashicorp/terraform-plugin-framework`. |
| HB-02 | Resource attributes | `id` (Computed String), `service_id` (Required String, RequiresReplace), `name` (Required String), `timeout_minutes` (Required Int64), `href` (Computed String, Sensitive). |
| HB-03 | In-place update | `name` and `timeout_minutes` support in-place updates via mutation `updateHeartbeatMonitor`. `service_id` cannot be modified in GoAlert and requires resource recreation via `stringplanmodifier.RequiresReplace()`. |
| HB-04 | Timeout boundary validation | `timeout_minutes` must be at least 5 minutes (`int64validator.AtLeast(5)`). |
| HB-05 | Direct O(1) read & drift detection | Read uses `Query.heartbeatMonitor(id: $id)`. When missing, GraphQL returns `null`, causing the resource to be removed from state cleanly (`resp.State.RemoveResource(ctx)`). |
| HB-06 | Deletion lifecycle | Deletion executes `deleteAll(input: [{type: heartbeatMonitor, id: $id}])`. Repeated deletion is idempotent. |
| HB-07 | Secret-bearing ping URL | `href` embeds the monitor ping endpoint (`.../api/v2/heartbeat/<id>`). Marked `Sensitive: true` to prevent leakage in CLI plans and logs. |
| HB-08 | Import syntax | Supports standard import by UUID: `terraform import goalert_heartbeat_monitor.example <id>`. |
| HB-09 | Canonical document expansion | Embeds `ProviderReadHeartbeatMonitor`, `ProviderCreateHeartbeatMonitor`, `ProviderUpdateHeartbeatMonitor`, `ProviderDeleteHeartbeatMonitor` in `internal/client/operations.graphql`. |

## Resource schema

```hcl
resource "goalert_heartbeat_monitor" "cron_backup" {
  service_id      = goalert_service.main.id
  name            = "Daily DB Backup"
  timeout_minutes = 15
}

output "backup_ping_url" {
  value     = goalert_heartbeat_monitor.cron_backup.href
  sensitive = true
}
```

### Attribute definitions

- `id` (String, Computed): The unique UUID of the heartbeat monitor.
- `service_id` (String, Required): UUID of the parent GoAlert service. Requires replacement if changed.
- `name` (String, Required): Name of the heartbeat monitor. Modifiable in place.
- `timeout_minutes` (Int64, Required): Heartbeat timeout threshold in minutes (must be >= 5). If no ping arrives within this window, an alert is triggered on the service. Modifiable in place.
- `href` (String, Computed, Sensitive): The HTTP ping URL to which periodic keep-alive requests (GET or POST) must be sent.
