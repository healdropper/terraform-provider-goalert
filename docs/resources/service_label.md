---
page_title: "goalert_service_label Resource"
description: "Manage a key-value label attached to a GoAlert service."
---

# goalert_service_label

Manages a label on a GoAlert service for categorization and filtering.

## Example

```hcl
resource "goalert_service" "app" {
  name                 = "Customer Gateway"
  escalation_policy_id = goalert_escalation_policy.primary.id
}

resource "goalert_service_label" "environment" {
  service_id = goalert_service.app.id
  key        = "example.com/environment"
  value      = "production"
}

resource "goalert_service_label" "team" {
  service_id = goalert_service.app.id
  key        = "example.com/team"
  value      = "core-infra"
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Computed compound identifier `<service_id>/<key>` |
| `service_id` | Required UUID of the parent GoAlert service. Requires replacement if changed |
| `key` | Required string in `<domain>/<suffix>` format (e.g. `example.com/environment`). Requires replacement if changed |
| `value` | Required string (3-255 characters). Modifiable in place |

## Import

```console
terraform import goalert_service_label.environment <service_id>/<key>
```
