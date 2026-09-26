# Development log

## 2026-09-26 — Integration key implementation and verification (Issues #25 & #26, Gates V003-IMPL & V003-VERIFY)

- Implemented `goalert_integration_key` resource using Terraform Plugin Framework (`internal/provider/integration_key_resource.go`).
- Expanded canonical operations document `internal/client/operations.graphql` and implemented client methods in `internal/client/client.go`.
- Added unit tests in `internal/client/client_test.go` and `internal/provider/integration_key_resource_test.go`.
- Added comprehensive end-to-end acceptance test `integration_key_acceptance` in `scripts/acceptance.py`:
  - Full CRUD lifecycle and sensitive `href` handling.
  - Real Grafana v1 payload ingestion to `href` returning HTTP 200.
  - Replacement on attribute change (deleting old key and creating new key with fresh token).
  - Drift detection and automatic recreation upon remote external deletion.
  - Import via standalone `<id>` and compound `<service_id>/<key_id>`.
  - Upstream AST hash validation: old v0.0.2 API key rejected with clean state preservation.
  - Complete remote teardown.
- All unit and acceptance tests passed against real disposable GoAlert v0.34.1 container.

## 2026-09-26 — Integration key contract specification (Issue #24, Gate V003-SPEC)

- Defined canonical specification `docs/specs/goalert-integration-key.md` and resource reference `docs/resources/integration_key.md`:
  - Resource: `goalert_integration_key`
  - Attributes: `id` (Computed), `service_id` (Required, RequiresReplace), `name` (Required, RequiresReplace), `type` (Optional, default `"grafana"`, RequiresReplace), `href` (Computed, Sensitive).
  - Explicit immutability model: since GoAlert has no `updateIntegrationKey`, PlanModifiers enforce resource recreation on any attribute modification.
  - O(1) direct read using `Query.integrationKey(id: ID!)` with `resp.State.RemoveResource(ctx)` when returning `null`.
  - Canonical GraphQL document expansion in `internal/client/operations.graphql` adding `ProviderReadIntegrationKey`, `ProviderCreateIntegrationKey`, and `ProviderDeleteIntegrationKey`.
  - Documented migration path for existing API key tokens.

## 2026-09-26 — Ingress key feasibility probe (Issue #9, Gate V003-DISC)

- Executed automated feasibility probe against disposable GoAlert v0.34.1:
  - Validated `createIntegrationKey` mutation with `grafana` and `generic` types.
  - Proved direct `Query.integrationKey(id: ID!)` returning `null` on missing IDs for O(1) drift detection.
  - Confirmed absence of `updateIntegrationKey`: keys are strictly immutable in GoAlert, requiring `RequiresReplace()` for all schema attributes.
  - Proved `href` format `<url>/api/v2/<type>/incoming?token=<id>`, requiring `Sensitive: true`.
  - Verified user-role API key permissions: users can create integration keys on accessible services.
  - Verified end-to-end webhook delivery: Grafana v1 alert payload to `href` returns HTTP 200 and creates unacknowledged alert.
- Recorded full findings and GraphQL contract in `docs/research/integration-key-feasibility.md`.

## 2026-09-26 — Milestone v0.0.2 closure and production adoption

- Closed Milestone v0.0.2: all scoped delivery gates (V002-FOUNDATION, V002-DISC, V002-SPEC, V002-IMPL, V002-VERIFY) completed and closed on GitHub and Project 1 board.
- Successfully verified live in production cluster `downstream-consumer/deployment`:
  - Configured GoAlert API key with canonical operations GraphQL document and `Webhook.Enable: true`.
  - Automated deployment applied `goalert_escalation_policy.watch_critical` and both `goalert_service` resources without drift.
  - End-to-end integration verified: Grafana, Prometheus, GoAlert and Alertmanager reporting healthy.
- Updated `docs/milestones/v0.0.2.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.

## 2026-09-25 — Upstream version watcher workflow (#20)

- Adopted `.github/workflows/upstream-watch.yml` in accordance with the universal doctrine rule.
- Workflow periodically queries `target/goalert` releases, detects when newer versions exceed `scripts/fixture.py`, and opens an assessment issue with upstream release notes and verification steps.
- Verification: `make check-format`, `go test ./...`, and `git diff --check` passed.

## 2026-09-20 — v0.0.1 service milestone

Goal: learn and implement a generic Terraform Plugin Framework provider,
starting with GoAlert service lifecycle rather than deployment-specific logic.

- Created the private GitHub repository and a dedicated development branch.
- Investigated GoAlert v0.34.1 source and existing GraphQL tooling.
- Proved fixed-document, multi-operation API-key CRUD before defining the
  Terraform resource. User-role keys cannot create services; admin is required.
- Implemented the embedded operation contract, bounded HTTP client, explicit
  credential configuration, service CRUD, import and conservative drift refresh.
- Real Terraform acceptance passed against disposable GoAlert and PostgreSQL:
  in-place updates, empty description, import, all-field drift repair, invalid-key
  failure preserving state, external deletion/recreation, remote destroy and
  repeated clean plans. Unit tests cover transport and error classification.
- Added MPL-2.0, examples, design/development/reference/release documentation,
  pinned CI actions and draft-only GPG-signed release automation.
- A first cancellation test exposed a hanging fake HTTP handler; bounded the
  fixture handler and reran the unit suite successfully.
- Shared-doctrine assessment: the durable lesson is specific to GoAlert's API
  contract and remains in this provider's design/tests.
- Release identity is not configured. No public visibility, Registry publication,
  production adoption, signed production release or PR merge has been performed.

- Consumer validation: the first consumer's isolated HCL passed the full acceptance
  harness using private development overrides. Its regression checks also passed.
- GoReleaser configuration and six-platform snapshot build passed. Native Windows
  Git-bundled GPG could not connect to its agent; the ephemeral-signature smoke
  test runs on Linux CI. A real release signer remains unconfigured.

## CI portability and signing follow-up

- Linux CI exposed a connection reset while the disposable container was starting.
  Readiness now retries transient connection failures within the existing deadline,
  including non-ready HTTP statuses; dedicated harness regression tests cover it.
- Corrected GoReleaser Action installation mode so later steps can find the binary.
- Linux packaging CI passed: six archives, Registry manifest, SHA256 checksums and
  a detached GPG signature verified with an ephemeral test key. This validates the
  signing process but does not establish or publish a production signing identity.

## 2026-09-21 - SpecDD transition

- Reread installed and repository instructions, verified remotes and pinned
  Foundry v1.6.0; downstream governance applies to the consumer by ownership.
- Recorded the [service contract](docs/specs/goalert-service.md) and
  [dated baseline](docs/sprints/sprint-0-baseline.md), preserving existing docs
  and historical verification without inventing prior TDD or acceptance.
- Added the [roadmap](docs/roadmap.md) and [proposed sprint](docs/sprints/sprint-1.md);
  their content is authoritative there rather than duplicated in this log.
- Opened [audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2)
  and its retrospective milestone; both await maintainer review.
- Fresh `make test vet check-format` passed (Go client results were cached).
  Historical real acceptance/signing CI was re-queried, not described as a fresh run.
- No provider behavior, branch name, repository visibility or production state changed.
  Process gaps and verification limits are recorded in the baseline and roadmap.
- Doctrine assessment: no upstream rule change needed; this is repository adoption.
- Documentation verification: 62 local links resolved and `git diff --check` passed.

## 2026-09-21 - Alert delivery product discovery

- Goal: analyze source-to-chat value and propose milestones before owner selection.
- Read pinned GoAlert v0.34.1 source and official Grafana/Telegram documentation.
  [Source evidence and limitations](docs/research/alert-delivery-feasibility.md)
  are authoritative there; no new API or message-delivery PoC was run.
- Added the [proposed contract](docs/specs/alert-delivery.md) and
  [milestone options](docs/milestones/alert-delivery.md). Updated the
  [roadmap](docs/roadmap.md) and linked the still-proposed
  [Sprint 1](docs/sprints/sprint-1.md), without accepting new scope.
- Recorded PR #1's actual merge in the roadmap; preserved the dated baseline.
  No provider release, infrastructure change or Telegram message was performed.
- Doctrine assessment: protocol constraints belong in this provider's research;
  consumer-specific deployment and bot ownership remain in the consumer.
- Verification: `make check-format` and `git diff --check` passed; 52 local
  documentation links, balanced code fences, proposal labels and generic-scope
  checks passed. Documentation-only work adds no behavioral test claims.

## 2026-09-21 - Select small-increment planning and create delivery tracking

- Recorded the owner's first Grafana firing scenario, group preference and
  interest in later independent availability detection in the
  [capability proposal](docs/specs/alert-delivery.md). No group or bot was created.
- Created [v0.0.2 planning](docs/milestones/v0.0.2.md), linked issues #4-#8 and
  updated [Sprint 1](docs/sprints/sprint-1.md). Detailed behavior and normal
  implementation remain pending discovery and owner acceptance.
- Created a [private Project](https://github.com/users/healdropper/projects/1),
  verified private visibility before attaching repository data, configured Board
  layout with Status columns and verified all ten intended items.
- Tracked ingress, independent availability and Telegram delivery as later
  Backlog issues #9-#11 without assigning them to v0.0.2.
- Added bug/enhancement intake templates; full delivery readiness remains in
  issue #4. Closure workflows report enabled, but no issue was artificially
  closed or marked Done to manufacture completion evidence.
- The initial ambiguous Project creation command was rejected by automatic
  review. A safe empty-container creation followed by explicit private readback
  succeeded before adding any repository link or planning content.
- Doctrine assessment: this applies the pinned lifecycle; no upstream policy
  change is needed. No PoC, feature implementation, merge, release or deployment.
- Verification: `make check-format` and `git diff --check` passed; 81 local
  links, code fences and both intake templates passed structural checks. GitHub
  readback confirmed five open milestone issues, three later backlog issues,
  Status-grouped boards and private visibility. No issues were closed.

## 2026-09-21 - Expose workflow_dispatch in release automation

- Goal: adopt the Foundry v1.8.0 release governance contingency rule in
  the release workflow, tracked under milestone v0.0.2 and Project 1.
- Updated `.github/workflows/release.yml` to include `workflow_dispatch:`
  alongside the existing `push: tags: ["v0.0.*"]` trigger.
- Preserved the existing ref name check and ancestor validation step so manual
  runs still require an exact patch tag ancestor of `origin/main`.
- Delivery tracking: linked to [issue #12](https://github.com/healdropper/terraform-provider-goalert/issues/12),
  assigned to milestone [v0.0.2](https://github.com/healdropper/terraform-provider-goalert/milestone/2),
  and tracked on [Project 1](https://github.com/users/healdropper/projects/1).
- Doctrine assessment: follows the release automation contingency rule from
  Foundry v1.8.0 / organization governance; no downstream repository policy change needed.
- Verification: `git diff --check`, `make check-format`, `go vet ./...` and `go test ./...` passed.

## 2026-09-22 - Complete SpecDD delivery prerequisites

- Goal: reconcile remaining delivery prerequisites for milestone v0.0.2 under V002-FOUNDATION.
- Verified governance, issue templates (`bug_report.md`, `enhancement.md`), and Project 1
  private boards and status column automations.
- Reconciled release-management: confirmed workflow_dispatch deployment (PR #13 / issue #12);
  explicitly recorded that maintained signing identity remains scheduled before public release.
- Maintained linkage to audit issue #2 while keeping the baseline milestone open pending review.
- Updated `docs/milestones/v0.0.2.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.
- Delivery tracking: resolves [issue #4](https://github.com/healdropper/terraform-provider-goalert/issues/4),
  unblocking V002-DISC (issue #5).
- Doctrine assessment: follows organization-neutral SpecDD lifecycle and Foundry governance.
- Verification: `git diff --check`, `python scripts/check_format.py`, `go vet ./...` and `go test ./...` passed.

## 2026-09-22 - Prove policy and webhook API feasibility

- Goal: establish API and webhook destination feasibility under V002-DISC on disposable GoAlert v0.34.1.
- Executed comprehensive probe suite (`scripts/test_feasibility.py`) against live container:
  - Discovered `builtin-webhook` is disabled by default; setting `Webhook.Enable: true` enables it.
  - Proved `createEscalationPolicy` atomically creates inline steps with `actions: [{type: "builtin-webhook", args: {webhook_url: ...}}]`.
  - Proved step deletion and reordering is controlled via `updateEscalationPolicy(stepIDs: [...])`, where omitted steps are deleted automatically.
  - Proved referential integrity: policy deletion is rejected while attached to a managed service.
  - Proved validation constraints: `delayMinutes >= 1`, `repeat >= 0`, `webhook_url` requires URI scheme.
  - Proved multi-operation canonical GraphQL document with AST hash preservation and backwards compatibility with service operations.
- Updated `docs/research/alert-delivery-feasibility.md`, `docs/milestones/v0.0.2.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.
- Delivery tracking: resolves [issue #5](https://github.com/healdropper/terraform-provider-goalert/issues/5), unblocking V002-SPEC (issue #6).
- Doctrine assessment: conforms to organization-neutral SpecDD lifecycle and Foundry governance.
- Verification: live test script `scripts/test_feasibility.py` passed with 100% empirical assertions; `git diff --check`, `python scripts/check_format.py`, `go vet ./...`, and `go test ./...` passed.

## 2026-09-22 - Resolve v0.0.2 policy resource contract

- Goal: define canonical specification for `goalert_escalation_policy` resource under V002-SPEC.
- Created canonical specification `docs/specs/goalert-escalation-policy.md`:
  - Resolved `goalert_escalation_policy` resource schema: `id`, `name`, `description`, `repeat`, and nested `step` blocks.
  - Specified `delay_minutes` (>= 1), `step_number` (0-indexed, computed), and `webhook_action` with required `url`.
  - Specified step ownership and reconciliation lifecycle: atomic creation via `createEscalationPolicy`, step modification via `updateEscalationPolicyStep`, new step addition via `createEscalationPolicyStep`, and reordering/pruning via `updateEscalationPolicy(stepIDs: [...])`.
  - Documented deletion cascade via `deleteAll`, referential integrity enforcement when referenced by a service, and key migration requirements for GraphQL AST hash validation.
- Updated `docs/milestones/v0.0.2.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.
- Delivery tracking: resolves [issue #6](https://github.com/healdropper/terraform-provider-goalert/issues/6), unblocking V002-IMPL (issue #7) and V002-VERIFY (issue #8).
- Doctrine assessment: conforms to organization-neutral SpecDD lifecycle and Foundry governance.
- Verification: `git diff --check`, `python scripts/check_format.py`, `go vet ./...`, and `go test ./...` passed.

## 2026-09-22 - Implement goalert_escalation_policy resource and lifecycle tests

- Goal: implement `goalert_escalation_policy` resource under V002-IMPL ([issue #7](https://github.com/healdropper/terraform-provider-goalert/issues/7)) and verify complete lifecycle and key migration under V002-VERIFY ([issue #8](https://github.com/healdropper/terraform-provider-goalert/issues/8)).
- Implemented `goalert_escalation_policy` resource using Terraform Plugin Framework (`internal/provider/escalation_policy_resource.go`):
  - Defined schema with `id`, `name`, `description`, `repeat`, and ordered `step` blocks (`ListNestedBlock`).
  - Implemented nested `webhook_action` blocks within `step`. Used `UseNonNullStateForUnknown()` plan modifiers on computed step fields (`id`, `step_number`) to support dynamic step additions without plan-null mismatch errors.
  - Implemented atomic creation submitting inline steps and actions in a single `createEscalationPolicy` mutation.
  - Implemented step reconciliation: modifying existing steps via `updateEscalationPolicyStep`, creating new steps via `createEscalationPolicyStep`, and reordering/pruning via `updateEscalationPolicy(stepIDs: [...])`.
  - Implemented drift detection and remote refresh for policies, steps, and webhook actions.
  - Implemented referential integrity error diagnostics when policy deletion is blocked by referencing services (`ErrInUse`).
- Updated GraphQL client (`internal/client/client.go`):
  - Added operations to `internal/client/operations.graphql` for escalation policies, steps, and actions.
  - Implemented `CreateEscalationPolicy`, `ReadEscalationPolicy`, `UpdateEscalationPolicy`, `DeleteEscalationPolicy`, `CreateEscalationPolicyStep`, and `UpdateEscalationPolicyStep`.
  - Enhanced error handling to recognize HTTP 422, `wrong query for API key`, and missing operations (`operation X not found`) as `API key document mismatch`, and `currently in use` as `ErrInUse`.
- Added resource documentation `docs/resources/escalation_policy.md` and Terraform examples in `examples/resources/goalert_escalation_policy/`.
- Updated test fixtures and acceptance tests:
  - Enabled webhook support in `compose.yaml` (`GOALERT_WEBHOOK_ENABLE: "true"`) and `scripts/fixture.py` (`setConfig(input: [{id: "Webhook.Enable", value: "true"}])`).
  - Added comprehensive `policy_acceptance` in `scripts/acceptance.py` covering: atomic creation, nested blocks, step delays, webhooks, in-place reordering, step additions, drift repair, import, referential integrity check on delete, and key migration validation (rejecting old v0.0.1 key without state loss).
- Verification:
  - Unit tests: `go test -v ./...` passed (client and provider packages).
  - Acceptance tests: `python scripts/acceptance.py` passed 100% against real GoAlert v0.34.1 disposable container.
  - Static analysis: `go fmt ./...`, `go vet ./...`, `git diff --check`, `python scripts/check_format.py`, and `python -m unittest discover -s scripts` passed with zero errors.
- Doctrine assessment: follows SpecDD lifecycle, preserves generic provider design, ensures zero leakage of sensitive credentials.
