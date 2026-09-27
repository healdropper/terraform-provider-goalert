---
page_title: "goalert_escalation_policy Data Source"
description: "Lookup a GoAlert escalation policy by its ID or name."
---

# goalert_escalation_policy (Data Source)

Fetches details of an existing GoAlert escalation policy by UUID or name search.

## Example

```hcl
data "goalert_escalation_policy" "default" {
  name = "Default Escalation"
}

output "policy_id" {
  value = data.goalert_escalation_policy.default.id
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Optional UUID. Exactly one of `id` or `name` must be specified |
| `name` | Optional policy name. Exactly one of `id` or `name` must be specified |
| `description` | Computed policy description string |
| `repeat` | Computed repeat count integer |
