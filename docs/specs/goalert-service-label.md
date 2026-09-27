# Service label resource contract

Owner: healdropper.
Status: Accepted specification for v0.0.4 under V004-SPEC.
Recorded: 2026-09-27.
Related: [Issue #32](https://github.com/healdropper/terraform-provider-goalert/issues/32),
[Milestone Heartbeat Monitors and Service Labels](../milestones/heartbeat-monitors-and-service-labels.md),
[Feasibility research](../research/heartbeat-feasibility.md).

## Requirements and scope

| ID | Requirement | Specification |
| --- | --- | --- |
| LBL-01 | Generic Plugin Framework resource | Resource type `goalert_service_label` implemented with `hashicorp/terraform-plugin-framework`. |
| LBL-02 | Resource attributes | `id` (Computed String: compound `<service_id>/<key>`), `service_id` (Required String, RequiresReplace), `key` (Required String, RequiresReplace), `value` (Required String). |
| LBL-03 | Key validation | Key must be formatted as `<prefix>/<suffix>`, where prefix follows domain naming rules (min 3 chars). |
| LBL-04 | Value validation | Value must be 3-255 characters, printable characters, no leading/trailing whitespace. |
| LBL-05 | In-place update | Modifying `value` executes `setLabel(input: {target: {type: service, id}, key, value})` in place. |
| LBL-06 | Deletion lifecycle | Deletion sets `value: ""` in `setLabel`, which deletes the label in GoAlert. |
| LBL-07 | Drift detection | Read queries service labels via `Query.service(id: $serviceID) { labels { key value } }`. If the key is missing or service does not exist, the resource is removed from state cleanly (`resp.State.RemoveResource(ctx)`). |
| LBL-08 | Import syntax | Supports compound ID import: `terraform import goalert_service_label.example <service_id>/<key>`. |

## Resource schema

```hcl
resource "goalert_service_label" "env" {
  service_id = goalert_service.main.id
  key        = "example.com/environment"
  value      = "production"
}
```

### Attribute definitions

- `id` (String, Computed): Compound identifier in the form `<service_id>/<key>`.
- `service_id` (String, Required): UUID of the parent GoAlert service. Requires replacement if changed.
- `key` (String, Required): Label key in `<domain>/<name>` format (e.g. `example.com/team`). Requires replacement if changed.
- `value` (String, Required): Label value string (length between 3 and 255 characters). Modifiable in place.
