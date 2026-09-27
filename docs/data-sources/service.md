---
page_title: "goalert_service Data Source"
description: "Lookup a GoAlert service by its ID or name."
---

# goalert_service (Data Source)

Fetches details of an existing GoAlert service by UUID or name search.

## Example

```hcl
data "goalert_service" "by_name" {
  name = "Production API"
}

output "service_policy_id" {
  value = data.goalert_service.by_name.escalation_policy_id
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Optional UUID. Exactly one of `id` or `name` must be specified |
| `name` | Optional service name. Exactly one of `id` or `name` must be specified |
| `description` | Computed service description string |
| `escalation_policy_id` | Computed UUID of the assigned escalation policy |
