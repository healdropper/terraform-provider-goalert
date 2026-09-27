# Foundational data sources contract

Owner: healdropper.
Status: Accepted specification for v0.0.4 under V004-SPEC.
Recorded: 2026-09-27.
Related: [Issue #32](https://github.com/healdropper/terraform-provider-goalert/issues/32),
[Milestone v0.0.4](../milestones/v0.0.4.md),
[Feasibility research](../research/heartbeat-feasibility.md).

## Requirements and scope

| ID | Data Source | Lookup Keys | Attributes Exposed |
| :--- | :--- | :--- | :--- |
| DS-01 | `data.goalert_service` | `id` (Optional) or `name` (Optional) | `id`, `name`, `description`, `escalation_policy_id` |
| DS-02 | `data.goalert_escalation_policy` | `id` (Optional) or `name` (Optional) | `id`, `name`, `description`, `repeat` |
| DS-03 | `data.goalert_integration_key` | `id` (Required) | `id`, `service_id`, `name`, `type`, `href` (Sensitive) |
| DS-04 | `data.goalert_heartbeat_monitor` | `id` (Required) | `id`, `service_id`, `name`, `timeout_minutes`, `href` (Sensitive) |

## Schemas and usage

### 1. `data "goalert_service"`

```hcl
# Lookup by exact ID
data "goalert_service" "by_id" {
  id = "a669e588-2770-4056-baef-fb7a6909643f"
}

# Or lookup by name search
data "goalert_service" "by_name" {
  name = "Production API"
}
```

Validation: Exactly one of `id` or `name` must be specified. If multiple services match the name search, the data source errors with a clear diagnostic instructing to refine or use `id`.

### 2. `data "goalert_escalation_policy"`

```hcl
data "goalert_escalation_policy" "default" {
  name = "Default Escalation"
}
```

Validation: Exactly one of `id` or `name` must be specified.

### 3. `data "goalert_integration_key"`

```hcl
data "goalert_integration_key" "grafana" {
  id = "e87f1234-5678-9abc-def0-123456789abc"
}
```

Validation: `id` is Required.

### 4. `data "goalert_heartbeat_monitor"`

```hcl
data "goalert_heartbeat_monitor" "backup" {
  id = "f98e1234-5678-9abc-def0-123456789abc"
}
```

Validation: `id` is Required.
