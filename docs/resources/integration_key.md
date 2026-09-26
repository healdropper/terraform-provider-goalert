---
page_title: "goalert_integration_key Resource"
description: "Manage a GoAlert integration key and incoming webhook endpoint for a service."
---

# goalert_integration_key

Manages an integration key on a GoAlert service. Integration keys provide webhook endpoints
allowing monitoring systems such as Grafana, Prometheus Alertmanager, or generic webhooks
to trigger and resolve alerts in GoAlert.

## Example

```hcl
resource "goalert_service" "app" {
  name                 = "Web Application"
  escalation_policy_id = goalert_escalation_policy.primary.id
}

resource "goalert_integration_key" "grafana" {
  service_id = goalert_service.app.id
  name       = "Grafana Ingress"
  type       = "grafana"
}

output "grafana_incoming_url" {
  value     = goalert_integration_key.grafana.href
  sensitive = true
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Computed integration key UUID |
| `service_id` | Required UUID of the parent GoAlert service. Requires replacement if changed |
| `name` | Required string. Requires replacement if changed |
| `type` | Optional string, defaults to `"grafana"`. Allowed values: `generic`, `grafana`, `site24x7`, `prometheusAlertmanager`, `email`, `universal`. Requires replacement if changed |
| `href` | Computed string, sensitive. The full incoming webhook URL including token |

## Immutability

GoAlert does not support updating integration keys in place. Modifying any attribute
(`service_id`, `name`, or `type`) forces Terraform to destroy and re-create the key,
generating a new token and webhook URL.

## Import

```console
terraform import goalert_integration_key.grafana 11111111-1111-4111-8111-111111111111
```

Or using compound format:

```console
terraform import goalert_integration_key.grafana <service_id>/<key_id>
```
