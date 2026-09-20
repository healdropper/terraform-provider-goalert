---
page_title: "goalert_service Resource"
description: "Manage a GoAlert service and its escalation policy reference."
---

# goalert_service

## Example

```hcl
resource "goalert_service" "api" {
  name                 = "Example API"
  description          = "API team alerts"
  escalation_policy_id = var.escalation_policy_id
}
```

| Attribute | Behavior |
| --- | --- |
| `id` | Computed service UUID, stable across updates |
| `name` | Required string; checked further by GoAlert |
| `description` | Optional string, defaults to `""`; removing it clears the API value |
| `escalation_policy_id` | Required existing lowercase policy UUID |

All three mutable attributes update in place and participate in drift detection.
Only confirmed remote absence removes the service from Terraform state.
Deleting a service uses GoAlert's service deletion semantics, including its
dependent service data; the referenced escalation policy is not deleted.

## Import

```console
terraform import goalert_service.api 11111111-1111-4111-8111-111111111111
```

The UUID is illustrative. Use the existing service's actual UUID and configure
matching attributes before applying. Import reads existing state; it does not
create the resource or infer the intended configuration.
