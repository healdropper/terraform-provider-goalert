# System Configuration Management

Owner: healdropper. Recorded: 2026-10-28.
Status: Active planning under [GitHub Milestone 11](https://github.com/healdropper/terraform-provider-goalert/milestone/11).
Parent intake: [Issue #88](https://github.com/healdropper/terraform-provider-goalert/issues/88).

## Recorded owner decisions

- Provide a singleton `goalert_system_config` resource managing only the GoAlert `Admin › Config` keys declared in Terraform (`settings` for `password: false` keys, `secrets_wo` for `password: true` keys).
- Enforce zero plaintext secret persistence in `terraform.tfstate` and saved plan files by using Terraform 1.11+ write-only attributes (`secrets_wo`, `WriteOnly: true`).
- Preserve full remote drift detection across all 68 configuration keys (including the 9 `password: true` keys) by computing salted `HMAC-SHA256` digests (`secret_hashes`) and scrubbing plaintext secret values immediately in `internal/client/client.go`.
- Unmanage keys removed from `settings` or `secrets_wo` during `Update` (drop from state without sending a mutation to GoAlert, matching `Delete` no-op semantics); explicit `""` clears a value while keeping the key declared.
- Surface GraphQL validation errors as `<ConfigID>: <reason>` with echoed input values scrubbed.
- Support zero-diff import via `system` and `system:<key1>,<key2>,...` (normalized to `id = "system"`).

## Planning envelope

| ID | Planning/acceptance gate | Evidence required before closing | Status |
| --- | --- | --- | --- |
| CFG-DISC | Establish GoAlert v0.35.0 `config` and `setConfig` API feasibility | Empirical probe against disposable GoAlert v0.35.0 characterizing `config(all: true)`, `setConfig` partial updates, newline `stringList` encoding, `validateEnable` cross-field checks, and GraphQL field error extensions in `docs/research/system-config-feasibility.md` | Completed ([#89](https://github.com/healdropper/terraform-provider-goalert/issues/89)) |
| CFG-SPEC | Resolve and accept `goalert_system_config` contract and documentation | Canonical specification `docs/specs/goalert-system-config.md` and Registry documentation `docs/resources/system_config.md` (with OIDC + `secrets_wo` example, zero-diff import, `ephemeral` secret source note, and high-entropy password note) | Completed ([#90](https://github.com/healdropper/terraform-provider-goalert/issues/90)) |
| CFG-CLIENT | Implement GraphQL operations, HMAC scrubbing, and sanitized field errors | `ProviderReadSystemConfig` and `ProviderSetSystemConfig` in `internal/client/operations.graphql`, client methods with immediate plaintext scrubbing and salted `HMAC-SHA256` digests, sanitized field validation errors, and unit tests in `internal/client/client_test.go` | Completed ([#91](https://github.com/healdropper/terraform-provider-goalert/issues/91)) |
| CFG-RESOURCE | Implement `goalert_system_config` resource with write-only secrets and zero-diff import | `internal/provider/system_config_resource.go` (`settings`, `secrets_wo`, `secret_hashes`, `ModifyPlan` HMAC comparison, key unmanagement on `Update`, state-only `Delete`, and `system:<keys>` import) plus unit tests in `internal/provider/system_config_resource_test.go` | Completed ([#92](https://github.com/healdropper/terraform-provider-goalert/issues/92)) |
| CFG-VERIFY | Verify `goalert_system_config` end-to-end lifecycle against disposable GoAlert v0.35.0 | `system_config_acceptance` suite in `scripts/acceptance.py` proving OIDC creation, zero plaintext secrets in `.tfstate`, in-place updates, secret rotation, remote drift repair, key unmanagement, zero-diff import, sanitized validation errors, and AST key migration | Planned ([#93](https://github.com/healdropper/terraform-provider-goalert/issues/93)) |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| Intake | [#88: feat: manage system configuration (Admin › Config) with a goalert_system_config resource](https://github.com/healdropper/terraform-provider-goalert/issues/88) | Triaged (`Track C`, `enhancement`) |
| CFG-DISC | [#89: feat(config): probe GoAlert v0.35.0 config and setConfig GraphQL behavior](https://github.com/healdropper/terraform-provider-goalert/issues/89) | Completed |
| CFG-SPEC | [#90: docs(config): specify goalert_system_config contract and user documentation](https://github.com/healdropper/terraform-provider-goalert/issues/90) | Completed |
| CFG-CLIENT | [#91: feat(client): add system config GraphQL operations, HMAC scrubbing, and sanitized field errors](https://github.com/healdropper/terraform-provider-goalert/issues/91) | Completed |
| CFG-RESOURCE | [#92: feat(provider): implement goalert_system_config resource with write-only secrets and zero-diff import](https://github.com/healdropper/terraform-provider-goalert/issues/92) | Completed |
| CFG-VERIFY | [#93: test(config): verify goalert_system_config end-to-end lifecycle and close #88](https://github.com/healdropper/terraform-provider-goalert/issues/93) | Ready |
