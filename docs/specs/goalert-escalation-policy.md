# Escalation policy resource contract

Owner: healdropper.
Status: Accepted specification for v0.0.2 under V002-SPEC.
Recorded: 2026-09-22.
Related: [Issue #6](https://github.com/healdropper/terraform-provider-goalert/issues/6),
[Milestone Escalation Policies and Webhook Routing](../milestones/escalation-policies-and-webhook-routing.md),
[Feasibility research](../research/alert-delivery-feasibility.md).

## Requirements and scope

| ID | Requirement | Specification |
| --- | --- | --- |
| POL-01 | Generic Plugin Framework resource | Resource type `goalert_escalation_policy` implemented with `hashicorp/terraform-plugin-framework`. |
| POL-02 | Attributes | `id` (Computed String), `name` (Required String), `description` (Optional/Computed String, default `""`), `repeat` (Optional/Computed Int64, default `0`). |
| POL-03 | Nested step blocks | Ordered list of `step` blocks: `id` (Computed String), `delay_minutes` (Required Int64, minimum 1), `step_number` (Computed Int64, 0-indexed), `multi_ack` (Optional/Computed Bool, default `false`), `user_ids` (Optional List of String), `rotation_ids` (Optional List of String), `schedule_ids` (Optional List of String), `webhook_action` (Optional list of blocks). |
| POL-04 | Webhook action block | `url` (Required String, must include valid scheme `http://` or `https://`). Mapped to GoAlert action `builtin-webhook` with argument `webhook_url`. |
| POL-05 | Atomic create | `createEscalationPolicy` submits inline steps, `multiAck`, and actions in a single atomic mutation when creating the policy. |
| POL-06 | Step reconciliation & ordering | Updates reconcile step definitions: modifying steps (including `multiAck`) via `updateEscalationPolicyStep`, creating new steps via `createEscalationPolicyStep`, and reordering/pruning via `updateEscalationPolicy(stepIDs: [...])`. Omitted step IDs are removed automatically by GoAlert. |
| POL-07 | Referential integrity | Deletion uses `deleteAll(type: escalationPolicy, id: $id)` which cascades to steps. If attached to a service, deletion fails with an explanatory diagnostic error. |
| POL-08 | Import and drift | Supports standard import by UUID (`terraform import goalert_escalation_policy.example <uuid>`). Refreshes remote drift accurately; clean second plan. |
| POL-09 | Key migration | Documented migration for keys: keys must be created with the expanded canonical document (`internal/client/operations.graphql`) to satisfy GoAlert AST hash validation. |

## Resource schema

```hcl
resource "goalert_escalation_policy" "example" {
  name        = "Production High Priority"
  description = "Escalates critical alerts to on-call destinations"
  repeat      = 3

  step {
    delay_minutes = 5
    multi_ack     = true
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

### Attribute definitions

- `id` (String, Computed): The unique UUID of the escalation policy.
- `name` (String, Required): The name of the escalation policy. Must not be empty.
- `description` (String, Optional, Computed): A description of the escalation policy. Defaults to `""`.
- `repeat` (Int64, Optional, Computed): Number of times the escalation policy repeats its steps if unacknowledged. Must be `>= 0`. Defaults to `0`.
- `step` (List of Objects, Optional): Ordered escalation steps (0-indexed):
  - `id` (String, Computed): UUID of the step.
  - `step_number` (Int64, Computed): Position of the step in the escalation sequence.
  - `delay_minutes` (Int64, Required): Delay in minutes before escalating to the next step. Must be `>= 1`.
  - `multi_ack` (Bool, Optional, Computed): When `true`, acknowledging an alert does not silence notifications for other users on the same step; each user continues to be notified until they individually acknowledge (GoAlert v0.35.0+). Defaults to `false`.
  - `user_ids` (List of String, Optional): Operator user UUIDs targeted in this step.
  - `rotation_ids` (List of String, Optional): Rotation UUIDs targeted in this step.
  - `schedule_ids` (List of String, Optional): Schedule UUIDs targeted in this step.
  - `webhook_action` (List of Objects, Optional): Webhook notification targets:
    - `url` (String, Required): Target webhook URL. Scheme `http://` or `https://` is required.

## API lifecycle and operation contract

The provider embeds the canonical multi-operation document (`operations.graphql`):

```graphql
query ProviderReadEscalationPolicy($id: ID!) {
  escalationPolicy(id: $id) {
    id
    name
    description
    repeat
    steps {
      id
      stepNumber
      delayMinutes
      actions {
        type
        args
      }
    }
  }
}

mutation ProviderCreateEscalationPolicy($input: CreateEscalationPolicyInput!) {
  createEscalationPolicy(input: $input) {
    id
    name
    description
    repeat
    steps {
      id
      stepNumber
      delayMinutes
      actions {
        type
        args
      }
    }
  }
}

mutation ProviderUpdateEscalationPolicy($input: UpdateEscalationPolicyInput!) {
  updateEscalationPolicy(input: $input)
}

mutation ProviderDeleteEscalationPolicy($id: ID!) {
  deleteAll(input: [{type: escalationPolicy, id: $id}])
}

mutation ProviderCreateEscalationPolicyStep($input: CreateEscalationPolicyStepInput!) {
  createEscalationPolicyStep(input: $input) {
    id
    stepNumber
    delayMinutes
    actions {
      type
      args
    }
  }
}

mutation ProviderUpdateEscalationPolicyStep($input: UpdateEscalationPolicyStepInput!) {
  updateEscalationPolicyStep(input: $input)
}
```

## Migration & prerequisites

1. **GoAlert Configuration**:
   The GoAlert instance must have `Webhook.Enable: true` configured (via `GOALERT_WEBHOOK_ENABLE=true` environment variable or GraphQL `setConfig`). If disabled, GoAlert rejects step creation with `"destination type is not enabled"`.
2. **API Key Migration**:
   GoAlert validates the query AST hash against the API key. To support both `goalert_service` and `goalert_escalation_policy`, operators must generate an API key using the complete `operations.graphql` document. Using an old v0.0.1 key results in `"wrong query for API key"`.
