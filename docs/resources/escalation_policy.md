---
page_title: "goalert_escalation_policy Resource"
description: "Manage a GoAlert escalation policy, ordered steps, and webhook notification actions."
---

# goalert_escalation_policy

## Example

```hcl
resource "goalert_escalation_policy" "production" {
  name        = "Production High Priority"
  description = "Escalates critical alerts to on-call destinations"
  repeat      = 3

  step {
    delay_minutes = 5
    webhook_action {
      url = "https://events.example.com/alerts"
    }
  }

  step {
    delay_minutes = 10
    webhook_action {
      url = "https://events.example.com/secondary"
    }
  }
}
```

| Attribute / Block | Type | Description |
| --- | --- | --- |
| `id` | String, Computed | Escalation policy UUID, stable across updates. |
| `name` | String, Required | Policy name; must not be empty. |
| `description` | String, Optional | Policy description; defaults to `""`. Removing clears the API value. |
| `repeat` | Int64, Optional | Number of times the escalation sequence repeats unacknowledged alerts. Defaults to `0`. |
| `step` | Block (Ordered List) | Ordered list of escalation steps (0-indexed). |
| `step.id` | String, Computed | Unique UUID of the escalation step. |
| `step.step_number` | Int64, Computed | 0-indexed position of the step in the escalation sequence. |
| `step.delay_minutes` | Int64, Required | Delay in minutes before escalating to the next step. Minimum `1`. |
| `step.webhook_action` | Block (List) | Webhook notification target for the step. |
| `step.webhook_action.url` | String, Required | Webhook endpoint URL. Must include scheme `http://` or `https://`. |

## Behavior and Lifecycle

- **Atomic creation**: Creating a policy with `step` and `webhook_action` blocks executes in a single atomic GraphQL mutation, avoiding partial resources.
- **In-place updates**: Changing name, description, repeat count, step delays, or actions reconciles in place without recreating the policy.
- **Step reordering and pruning**: Step order is preserved. Reordering, adding, or removing steps updates GoAlert's step sequence and step list accordingly.
- **Drift detection**: Full refresh captures out-of-band changes to the policy, steps, and webhook actions.
- **Referential integrity**: GoAlert prevents deleting escalation policies currently attached to services. Deletion attempts on in-use policies fail with an explanatory diagnostic message.
- **Webhook prerequisite**: GoAlert must have `Webhook.Enable` set to `true` (`GOALERT_WEBHOOK_ENABLE=true` or via system configuration) to accept webhook actions.

## Import

```console
terraform import goalert_escalation_policy.production 11111111-1111-4111-8111-111111111111
```

The UUID is illustrative. Use the existing escalation policy's actual UUID and configure matching attributes before applying. Import reads existing state and populates steps and webhook actions; it does not infer unconfigured local state.
