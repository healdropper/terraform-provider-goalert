# Heartbeat monitor, label, and data source feasibility: live probe and API analysis

Status: Verified runtime evidence against disposable GoAlert v0.35.0 container.
Owner: healdropper. Date: 2026-09-27.
GoAlert target image: `goalert/goalert:v0.35.0@sha256:f090a90538d7e61446aad245c21a7a1694d36d7b4f6c56fc55e9cc7405d9c03f`.
Tracking: [Issue #18](https://github.com/healdropper/terraform-provider-goalert/issues/18) (`V004-DISC`) and [Issue #22](https://github.com/healdropper/terraform-provider-goalert/issues/22) (`V004-UPSTREAM`).
Probe script: [`scripts/test_heartbeat_feasibility.py`](../../scripts/test_heartbeat_feasibility.py).

## Executive summary

This feasibility study proves the end-to-end lifecycle and GraphQL contract for managing
GoAlert heartbeat monitors (`goalert_heartbeat_monitor`), service labels (`goalert_service_label`),
and foundational data sources (`data.goalert_service`, `data.goalert_escalation_policy`,
`data.goalert_integration_key`, `data.goalert_heartbeat_monitor`).

All findings below were verified via automated runtime probes against a disposable GoAlert
v0.35.0 instance.

## Verified findings

| Finding | Runtime evidence | Architectural & Terraform implication |
| :--- | :--- | :--- |
| `createHeartbeatMonitor` creates monitors on services | Executed `createHeartbeatMonitor(input: {serviceID, name, timeoutMinutes})` returning `id`, `serviceID`, `name`, `timeoutMinutes`, `href` | Clean single-operation creation. |
| Timeout boundary validation | `timeoutMinutes: 1` rejected with `must not be below 5m0s` | Minimum timeout is 5 minutes. The Terraform resource schema must enforce `int64validator.AtLeast(5)`. |
| Update mutation exists | `updateHeartbeatMonitor(input: {id, name, timeoutMinutes})` returns `true` | Unlike integration keys, heartbeat monitors support **in-place updates** for `name` and `timeout_minutes`. `service_id` is immutable and requires replacement. |
| Direct read query exists | `Query.heartbeatMonitor(id: ID!): HeartbeatMonitor` | O(1) `Read` operation for both resource and data source. Missing ID returns `null` (enables clean drift detection). |
| Ping URL format | `href` format: `<url>/api/v2/heartbeat/<id>` | The ping endpoint is secret-bearing and triggers the keep-alive signal. Must be marked `Sensitive: true` in Terraform schema. |
| Ping delivery verified | HTTP POST to `href` returns HTTP 200 | Proved ping receipt. Runtime state (`lastState`) is operational runtime data, not managed desired state. |
| Deletion via `deleteAll` | `deleteAll(input: [{type: heartbeatMonitor, id: $id}])` successfully removes monitor | Reuses existing generic `deleteAll` mutation pattern. |
| Service Labels: `setLabel` | `setLabel(input: {target: {type: service, id}, key, value})` returns `true` | Creates and updates labels. Setting `value: ""` removes the label. |
| Label key validation rules | Key must be `<prefix>/<suffix>`, prefix must be valid domain format (>= 3 chars) | E.g. `example.com/environment`. Schema should document or validate `<domain>/<key>` structure. |
| Label value validation rules | Value must be 3-255 characters, printable chars, no leading/trailing spaces | Schema validation with `stringvalidator.LengthBetween(3, 255)`. |
| Data source search queries | `services(input: {search: $name})` and `escalationPolicies(input: {search: $name})` return matching nodes | Enables lookup by `name` or `id` in data sources. |
| Least privilege permissions | `user` role API key successfully creates/updates heartbeat monitors and labels | Least privilege confirmed; administrative role is not strictly required. |

## Canonical GraphQL operations contract

The operations to register in `internal/client/operations.graphql`:

```graphql
query ProviderReadHeartbeatMonitor($id: ID!) {
  heartbeatMonitor(id: $id) {
    id
    serviceID
    name
    timeoutMinutes
    href
  }
}

mutation ProviderCreateHeartbeatMonitor($input: CreateHeartbeatMonitorInput!) {
  createHeartbeatMonitor(input: $input) {
    id
    serviceID
    name
    timeoutMinutes
    href
  }
}

mutation ProviderUpdateHeartbeatMonitor($input: UpdateHeartbeatMonitorInput!) {
  updateHeartbeatMonitor(input: $input)
}

mutation ProviderDeleteHeartbeatMonitor($id: ID!) {
  deleteAll(input: [{type: heartbeatMonitor, id: $id}])
}

mutation ProviderSetServiceLabel($input: SetLabelInput!) {
  setLabel(input: $input)
}

query ProviderReadServiceLabels($id: ID!) {
  service(id: $id) {
    id
    labels {
      key
      value
    }
  }
}

query ProviderSearchServices($search: String!) {
  services(input: {search: $search, first: 15}) {
    nodes {
      id
      name
      description
      escalationPolicy { id }
    }
  }
}

query ProviderSearchEscalationPolicies($search: String!) {
  escalationPolicies(input: {search: $search, first: 15}) {
    nodes {
      id
      name
      description
      repeat
    }
  }
}
```

## Security and sensitivity analysis

1. **Ping URL (`href`)**: The URL contains the monitor ID (`/api/v2/heartbeat/<id>`). Sending HTTP GET or POST to this URL records a heartbeat ping. The URL should be marked `Sensitive: true` in state.
2. **Import Syntax**:
   - `goalert_heartbeat_monitor`: `terraform import goalert_heartbeat_monitor.example <id>`. Resolves `service_id`, `name`, `timeout_minutes`, and `href` directly via `Query.heartbeatMonitor`.
   - `goalert_service_label`: `terraform import goalert_service_label.example <service_id>/<key>`.
