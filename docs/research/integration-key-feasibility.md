# Integration key feasibility: live probe and API analysis

Status: Verified runtime evidence against disposable GoAlert v0.34.1 container.
Owner: healdropper. Date: 2026-09-26.
GoAlert target image: `goalert/goalert:v0.34.1@sha256:c3ebee8abdf6035cbf7ea2aa855807ce484d44e5125d413cde1f9811039ddb75`.
Tracking: [Issue #9](https://github.com/healdropper/terraform-provider-goalert/issues/9) (`V003-DISC`).
Probe script: [`scripts/test_ingress_key_feasibility.py`](../../scripts/test_ingress_key_feasibility.py).

## Executive summary

This feasibility study proves the end-to-end lifecycle and GraphQL contract for managing
GoAlert integration keys (`goalert_integration_key`) on services, specifically targeting
Grafana webhook ingress.

All findings below were verified via automated runtime probes against a disposable GoAlert
v0.34.1 instance.

## Verified findings

| Finding | Runtime evidence | Architectural & Terraform implication |
| --- | --- | --- |
| `createIntegrationKey` creates keys on services | Executed `createIntegrationKey(input: {serviceID, name, type})` returning `id`, `name`, `type`, `href`, `serviceID` | Clean single-operation creation. |
| Supported key types | Enum `IntegrationKeyType` exposes: `generic`, `grafana`, `site24x7`, `prometheusAlertmanager`, `email`, `universal` | Default `type = "grafana"`; validate type string against enum set. |
| Direct read query exists | `Query.integrationKey(id: ID!): IntegrationKey` | O(1) `Read` operation without paginating or scanning `Service.integrationKeys`. |
| Missing key returns `null` | Querying missing ID returns `{"data": {"integrationKey": null}}` without errors | Clean drift detection; `resp.State.RemoveResource(ctx)` when `integrationKey` is null. |
| No update mutation exists | `Mutation.updateIntegrationKey` does not exist in schema | Integration keys are **strictly immutable**. Schema attributes `name`, `type`, and `service_id` must declare `PlanModifiers: [stringplanmodifier.RequiresReplace()]`. |
| Deletion via `deleteAll` | `deleteAll(input: [{type: integrationKey, id: $id}])` successfully removes key | Reuses existing `deleteAll` pattern implemented in provider. |
| Token matches Key ID in URL | `href` format: `<url>/api/v2/<type>/incoming?token=<id>` | Token query parameter is identical to the key UUID. `href` is secret-bearing and must be marked `Sensitive: true`. |
| Role permissions | `user` role API key successfully executed `createIntegrationKey` | Key creation does not require system admin role, adhering to least-privilege principles. |
| Webhook delivery generates alerts | Sent Grafana v1 payload to `href` -> HTTP 200 returned -> alert generated on service | Proven end-to-end compatibility with Grafana alerting webhook format. |

## GraphQL contract

The canonical fixed-document operations required for `goalert_integration_key`:

```graphql
query ProviderReadIntegrationKey($id: ID!) {
  integrationKey(id: $id) {
    id
    name
    type
    href
    serviceID
  }
}

mutation ProviderCreateIntegrationKey($input: CreateIntegrationKeyInput!) {
  createIntegrationKey(input: $input) {
    id
    name
    type
    href
    serviceID
  }
}

mutation ProviderDeleteIntegrationKey($id: ID!) {
  deleteAll(input: [{type: integrationKey, id: $id}])
}
```

## Security and sensitivity analysis

1. **Ingress Token**: The token required to trigger alerts via `/api/v2/grafana/incoming?token=<token>` or `/api/v2/generic/incoming?token=<token>` is the key's UUID.
2. **Terraform State**:
   - `href`: Must be marked `Sensitive: true` so the webhook URL (bearing the authentication token) is masked in CLI plans and logs.
   - `id`: Standard resource ID (UUID).
3. **Import syntax**: `terraform import goalert_integration_key.example <id>` or `<service_id>/<id>`. Since `Query.integrationKey(id: $id)` only requires the key ID, importing by `<id>` is fully supported and resolves `serviceID` directly.
