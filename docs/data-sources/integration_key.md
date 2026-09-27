---
page_title: "goalert_integration_key Data Source"
description: "Lookup a GoAlert integration key by its ID."
---

# goalert_integration_key (Data Source)

Fetches details of an existing GoAlert integration key by UUID.

## Example

```hcl
data "goalert_integration_key" "grafana" {
  id = "e87f1234-5678-9abc-def0-123456789abc"
}

output "webhook_url" {
  value     = data.goalert_integration_key.grafana.href
  sensitive = true
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Required integration key UUID |
| `service_id` | Computed UUID of the parent GoAlert service |
| `name` | Computed name of the integration key |
| `type` | Computed key type (e.g. `grafana`, `generic`) |
| `href` | Computed sensitive webhook URL |
