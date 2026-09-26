# Integration key resource contract

Owner: healdropper.
Status: Accepted specification for v0.0.3 under V003-SPEC.
Recorded: 2026-09-26.
Related: [Issue #24](https://github.com/healdropper/terraform-provider-goalert/issues/24),
[Milestone v0.0.3](../milestones/v0.0.3.md),
[Feasibility research](../research/integration-key-feasibility.md).

## Requirements and scope

| ID | Requirement | Specification |
| --- | --- | --- |
| INT-01 | Generic Plugin Framework resource | Resource type `goalert_integration_key` implemented with `hashicorp/terraform-plugin-framework`. |
| INT-02 | Resource attributes | `id` (Computed String), `service_id` (Required String), `name` (Required String), `type` (Optional/Computed String, default `"grafana"`), `href` (Computed String, Sensitive). |
| INT-03 | Strict immutability | GoAlert v0.34.1 exposes no `updateIntegrationKey` mutation. Any modification to `name`, `type`, or `service_id` forces resource replacement via `RequiresReplace()`. |
| INT-04 | Key type validation | Allowed types match upstream `IntegrationKeyType` enum: `generic`, `grafana`, `site24x7`, `prometheusAlertmanager`, `email`, `universal`. Default is `"grafana"`. |
| INT-05 | Direct O(1) read & drift detection | Read uses `Query.integrationKey(id: $id)` directly. When a key is deleted externally, GraphQL returns `null`, causing the resource to be removed from state cleanly (`resp.State.RemoveResource(ctx)`). |
| INT-06 | Deletion lifecycle | Deletion executes `deleteAll(input: [{type: integrationKey, id: $id}])`. Repeated deletion is idempotent. |
| INT-07 | Secret-bearing URL handling | `href` embeds the integration token query parameter (`.../incoming?token=<id>`). It is marked `Sensitive: true` to prevent token leakage in CLI output and logs. |
| INT-08 | Import syntax | Supports standard import by UUID (`terraform import goalert_integration_key.example <id>`) or compound ID (`<service_id>/<id>`). |
| INT-09 | Canonical document expansion | The provider embeds `internal/client/operations.graphql` with three new operations: `ProviderReadIntegrationKey`, `ProviderCreateIntegrationKey`, `ProviderDeleteIntegrationKey`. |

## Resource schema

```hcl
resource "goalert_integration_key" "grafana" {
  service_id = goalert_service.watch_cluster.id
  name       = "Grafana Ingress"
  type       = "grafana"
}

output "grafana_webhook_url" {
  value     = goalert_integration_key.grafana.href
  sensitive = true
}
```

### Attribute definitions

- `id` (String, Computed): The unique UUID of the integration key.
- `service_id` (String, Required): UUID of the parent GoAlert service. Requires replacement if changed.
- `name` (String, Required): Name of the integration key. Requires replacement if changed.
- `type` (String, Optional, Computed): Type of integration key. Must be one of `generic`, `grafana`, `site24x7`, `prometheusAlertmanager`, `email`, `universal`. Defaults to `"grafana"`. Requires replacement if changed.
- `href` (String, Computed, Sensitive): The complete incoming webhook URL (including authentication token) to which alerts should be posted.

## API lifecycle and operation contract

The provider embeds the canonical multi-operation document (`internal/client/operations.graphql`):

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

### Create
Executes `ProviderCreateIntegrationKey` with variables:
```json
{
  "input": {
    "serviceID": "867f48af-8f86-486f-b258-19e3a27be2aa",
    "name": "Grafana Ingress",
    "type": "grafana"
  }
}
```
State is populated with returned `id`, `name`, `type`, `href`, and `service_id`.

### Read
Executes `ProviderReadIntegrationKey` with variables:
```json
{
  "id": "1cb84163-06f1-460c-8795-8777a7267268"
}
```
- If returned `integrationKey` is `null`: calls `resp.State.RemoveResource(ctx)`.
- Otherwise: synchronizes `name`, `type`, `href`, and `service_id`.

### Update
Because GoAlert integration keys are strictly immutable, PlanModifiers ensure any attribute change triggers resource replacement. No in-place update operation is executed.

### Delete
Executes `ProviderDeleteIntegrationKey` with variables:
```json
{
  "id": "1cb84163-06f1-460c-8795-8777a7267268"
}
```

### Import
Accepts either `<id>` or `<service_id>/<id>`. If compound syntax is provided, the key ID after the slash is extracted. `ProviderReadIntegrationKey` is invoked to populate all attributes from the remote instance.

## API key document migration

Adding operations to `operations.graphql` alters the document AST hash.
When upgrading to provider v0.0.3, any pre-existing GoAlert API key restricted to the v0.0.2 operations document will reject integration key operations with:
```
wrong query for API key
```
To migrate existing environments:
1. In GoAlert UI, navigate to Admin -> API Keys.
2. Edit or regenerate the provider API key with the updated `internal/client/operations.graphql` content.
