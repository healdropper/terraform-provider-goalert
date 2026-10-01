# Development log

## 2026-10-01 — Milestone 10 planning: Registry Publication and GA (`v1.0.0`)

- Created [GitHub Milestone 10 (`Registry Publication and GA`)](https://github.com/healdropper/terraform-provider-goalert/milestone/10) and authored canonical milestone plan [`docs/milestones/registry-publication-and-ga.md`](docs/milestones/registry-publication-and-ga.md) aligned with Doctrine Foundry v1.11.0 public repository governance rules.
- Opened delivery gate issues [#79](https://github.com/healdropper/terraform-provider-goalert/issues/79) (`GA-DISC`), [#80](https://github.com/healdropper/terraform-provider-goalert/issues/80) (`GA-SPEC`), [#81](https://github.com/healdropper/terraform-provider-goalert/issues/81) (`GA-IMPL`), and [#82](https://github.com/healdropper/terraform-provider-goalert/issues/82) (`GA-VERIFY`) and linked them to Project #1.

## 2026-10-01 — Embrace Latest GoAlert Version delivery and closure (Issues #70, #71, #72, #73, #22)

- **UPST-DISC (Issue #70, PR #75):** Ran empirical probe `scripts/test_v035_feasibility.py` against disposable GoAlert v0.35.0 (`docs/research/goalert-v035-capabilities-feasibility.md`), verifying `multiAck` on escalation policy steps, owner-only SQL visibility of `private = true` contact methods (`CMStore.FindOne`), `enableStatusUpdates` / `statusUpdates`, and polymorphic `setLabel` across `service`, `escalationPolicy`, `schedule`, and `rotation`.
- **UPST-SPEC (Issue #71, PR #76):** Updated canonical contracts `docs/specs/goalert-escalation-policy.md` and `docs/specs/goalert-user-contact-method.md`, authored `docs/specs/goalert-label.md`, and published Registry docs in `docs/resources/`.
- **UPST-IMPL (Issue #72, PR #77):** Implemented `multi_ack` in `goalert_escalation_policy`, `enable_status_updates` / `private` / `status_updates` in `goalert_user_contact_method`, and the polymorphic `goalert_label` resource (`internal/provider/label_resource.go`), with unit tests in `internal/client/` and `internal/provider/`.
- **UPST-VERIFY (Issue #73, PR #78):** Added and verified `v035_acceptance` in `scripts/acceptance.py` (5/5 end-to-end checks passing against disposable GoAlert v0.35.0: creation, in-place `multi_ack` & label updates, external drift/deletion remediation, `terraform import`, and clean teardown). Closed Milestone 9 (`Embrace Latest GoAlert Version`) and upstream tracking [Issue #22](https://github.com/healdropper/terraform-provider-goalert/issues/22).

## 2026-10-01 — Upstream triage and Milestone 9 planning: Embrace Latest GoAlert Version

- Executed `issue-triage-engineer` and `delivery-planning` workflows:
  - Closed [Issue #19](https://github.com/healdropper/terraform-provider-goalert/issues/19) as completed and superseded by Milestones 5, 6, and 7.
  - Audited the GoAlert `v0.34.1...v0.35.0` upstream GraphQL schema delta (`graphql2/schema.graphql`, `graphql2/graph/escalationpolicy.graphqls`) originally surfaced by [Issue #22](https://github.com/healdropper/terraform-provider-goalert/issues/22):
    1. `multiAck` on `EscalationPolicyStep`, `CreateEscalationPolicyStepInput`, and `UpdateEscalationPolicyStepInput`.
    2. `private` and `enableStatusUpdates` on `UserContactMethod`, `CreateUserContactMethodInput`, and `UpdateUserContactMethodInput`.
    3. Polymorphic `labels` on `EscalationPolicy`, `Schedule`, and `Rotation` via `setLabel`.
- Created [GitHub Milestone 9 (`Embrace Latest GoAlert Version`)](https://github.com/healdropper/terraform-provider-goalert/milestone/9) and shifted `Registry Publication and GA` to Milestone 10.
- Authored canonical milestone plan [`docs/milestones/embrace-latest-goalert-version.md`](docs/milestones/embrace-latest-goalert-version.md) and opened delivery issues [#70](https://github.com/healdropper/terraform-provider-goalert/issues/70) (UPST-DISC), [#71](https://github.com/healdropper/terraform-provider-goalert/issues/71) (UPST-SPEC), [#72](https://github.com/healdropper/terraform-provider-goalert/issues/72) (UPST-IMPL), and [#73](https://github.com/healdropper/terraform-provider-goalert/issues/73) (UPST-VERIFY).

## 2026-09-28 — Collaboration Channels and System Limits delivery and closure (Issues #61, #62, #63, #64)

- **COL-DISC (Issue #61, PR #66):** Characterized GoAlert v0.35.0 `systemLimits`, `setSystemLimits`, `slackChannels`, and `slackUserGroups` schemas in `docs/research/collaboration-and-limits-feasibility.md` via `scripts/test_collaboration_feasibility.py`.
- **COL-SPEC (Issue #62, PR #67):** Authored specifications `docs/specs/goalert-system-limit.md`, `docs/specs/goalert-slack-channel-data-source.md`, `docs/specs/goalert-slack-user-group-data-source.md`, and Registry documentation.
- **COL-IMPL (Issue #63, PR #68):** Implemented `goalert_system_limit` resource (`internal/provider/system_limit_resource.go`), `data.goalert_slack_channel`, and `data.goalert_slack_user_group`, expanding `internal/client/operations.graphql` and `internal/client/client.go`.
- **COL-VERIFY (Issue #64, PR #69):** Implemented and verified `system_limits_acceptance` in `scripts/acceptance.py` (6/6 assertions passing against disposable GoAlert v0.35.0) and closed Milestone 8.

## 2026-09-28 — Schedules and User Overrides verification and closure (Issue #55, Gate SCHED-VERIFY)

- Implemented and executed end-to-end acceptance test suite `schedules_acceptance` in `scripts/acceptance.py`:
  1. Creation of users, rotation, schedule, schedule rule with weekday filter, user override, and multi-target escalation policy.
  2. Data source lookups for `data.goalert_schedule` by UUID and name.
  3. In-place modification of schedule attributes, rule active hours (08:00 - 18:00), and override duration.
  4. Remote drift detection, external user override deletion, and state reconciliation.
  5. State import of `goalert_schedule`, compound key import of `goalert_schedule_rule`, and import of `goalert_user_override`.
  6. AST hash key migration validation (outdated v0.0.6 API key rejected with HTTP 422).
  7. Clean remote destroy of schedules, rules, overrides, rotations, and policies.
- Verified 100% passing across unit test suite (`go test ./...`) and end-to-end acceptance tests.
- Resolves [Issue #55](https://github.com/healdropper/terraform-provider-goalert/issues/55), successfully completing Milestone 7 (`Schedules and User Overrides`).

## 2026-09-28 — Schedule, rule, override, and step destination implementation (Issue #54, Gate SCHED-IMPL)

- Expanded canonical GraphQL document `internal/client/operations.graphql` with schedule operations: `ProviderCreateSchedule`, `ProviderReadSchedule`, `ProviderSearchSchedules`, `ProviderUpdateSchedule`, `ProviderDeleteSchedule`, `ProviderUpdateScheduleTarget`, `ProviderCreateUserOverride`, `ProviderReadUserOverride`, `ProviderUpdateUserOverride`, `ProviderDeleteUserOverride`.
- Implemented type-safe client methods and data types in `internal/client/client.go` for schedules, rules, overrides, and error-safe unmarshaling.
- Implemented `goalert_schedule` resource (`internal/provider/schedule_resource.go`) with IANA timezone validation and name pattern verification.
- Implemented `goalert_schedule_rule` resource (`internal/provider/schedule_rule_resource.go`) supporting rotation/user target assignments with 24-hour clock time windows and 7-day weekday filters.
- Implemented `goalert_user_override` resource (`internal/provider/user_override_resource.go`) supporting add/remove user shift substitutions with RFC3339 timestamps.
- Implemented `goalert_schedule` data source (`internal/provider/schedule_data_source.go`) supporting lookup by UUID or exact name.
- Extended `goalert_escalation_policy` (`internal/provider/escalation_policy_resource.go`) to support `schedule_ids` in `step` blocks, mapping directly to GoAlert `builtin-schedule` actions.
- Registered all new resources and data sources in `internal/provider/provider.go`.
- Added unit tests in `internal/provider/schedule_resource_test.go`, `internal/provider/schedule_rule_resource_test.go`, `internal/provider/user_override_resource_test.go`, and updated `internal/provider/escalation_policy_resource_test.go`.
- Resolves [Issue #54](https://github.com/healdropper/terraform-provider-goalert/issues/54), unblocking SCHED-VERIFY (Issue #55).

## 2026-09-28 — Rotations and Escalation Targets verification and closure (Issue #46, Gate ROT-VERIFY)

- Diagnosed GoAlert v0.35.0 GraphQL destination behavior: when configuring escalation policy steps, the modern `actions` array supersedes legacy `targets`. Passing targets separately caused `cty.ListVal` vs `null` inconsistencies after apply.
- Updated `internal/provider/escalation_policy_resource.go` to unify all step targets (`user_ids`, `rotation_ids`, and `webhook_action`) into the typed `Actions` destination list (`builtin-user`, `builtin-rotation`, `builtin-webhook`).
- Fixed `ProviderSearchRotations` GraphQL query argument signature (`input: {search: $search, first: 50}`) in `internal/client/operations.graphql`.
- Implemented and executed end-to-end acceptance test suite `rotations_acceptance` in `scripts/acceptance.py`:
  1. Creation of users, rotation with daily shifts, and multi-target escalation policy.
  2. Data source lookups for `data.goalert_rotation` by UUID and name.
  3. In-place participant reordering and step delay updates.
  4. Remote drift detection, rotation recreation, and automatic policy step reconciliation.
  5. State import of `goalert_rotation` by UUID.
  6. AST hash key migration validation (outdated v0.0.5 API key rejected with HTTP 422).
  7. Clean remote destroy of rotation and escalation policy.
- Verified 100% passing across unit test suite (`go test ./...`) and acceptance tests.
- Resolves [Issue #46](https://github.com/healdropper/terraform-provider-goalert/issues/46), successfully completing Milestone 6 (`Rotations and Escalation Targets`).

## 2026-09-27 — Rotation resource, data source, and escalation step targets implementation (Issue #45, Gate ROT-IMPL)

- Expanded canonical GraphQL document `internal/client/operations.graphql` with rotation operations: `ProviderCreateRotation`, `ProviderReadRotation`, `ProviderSearchRotations`, `ProviderUpdateRotation`, `ProviderDeleteRotation`.
- Implemented type-safe client methods and data types in `internal/client/client.go`: `CreateRotation`, `ReadRotation`, `SearchRotations`, `UpdateRotation`, `DeleteRotation`, and extended `CreateEscalationPolicyStepInput` / `UpdateEscalationPolicyStepInput` with `Targets []TargetInput`.
- Implemented `goalert_rotation` resource (`internal/provider/rotation_resource.go`) with validation for name pattern, shift frequencies (`daily`, `weekly`, `hourly`), timezone, and participant list.
- Implemented `goalert_rotation` data source (`internal/provider/rotation_data_source.go`) supporting lookup by UUID or exact name search.
- Extended `goalert_escalation_policy` (`internal/provider/escalation_policy_resource.go`) step block schema to support `user_ids` and `rotation_ids` targets alongside `webhook_action`. Updated `stepMatches` and step creation/update logic.
- Registered `NewRotationResource` and `NewRotationDataSource` in `internal/provider/provider.go`.
- Added unit tests covering all rotation operations and step target matching in `internal/client/client_test.go`, `internal/provider/rotation_resource_test.go`, and `internal/provider/escalation_policy_resource_test.go`.
- Delivery tracking: resolves [Issue #45](https://github.com/healdropper/terraform-provider-goalert/issues/45), unblocking ROT-VERIFY (Issue #46).

## 2026-09-27 — Milestone planning: Rotations and Escalation Targets and doctrine milestone naming alignment

- Renamed past GitHub milestones and updated `docs/roadmap.md` to follow the updated Spec-Driven Lifecycle doctrine: milestones represent functional delivery themes rather than strict SemVer patch increments.
  - Milestone 2: `Escalation Policies and Webhook Routing` (was `v0.0.2`).
  - Milestone 3: `Ingress and Integration Keys` (was `v0.0.3`).
  - Milestone 4: `Heartbeat Monitors and Service Labels` (was `v0.0.4`).
  - Milestone 5: `User Identity and Notification Rules` (was `v0.0.5`).
- Created active GitHub Milestone 6: `Rotations and Escalation Targets`.
- Authored canonical milestone plan `docs/milestones/rotations-and-escalation-targets.md`.
- Opened delivery issues #43 (ROT-DISC), #44 (ROT-SPEC), #45 (ROT-IMPL), and #46 (ROT-VERIFY).

## 2026-09-27 — Milestone v0.0.4 delivery: heartbeats, labels, data sources, and v0.35.0 (Issues #18, #22, #32, #33, #34)

- **V004-UPSTREAM (Issue #22):** Upgraded upstream GoAlert container image to `goalert/goalert:v0.35.0` (digest `sha256:f090a90538d7e61446aad245c21a7a1694d36d7b4f6c56fc55e9cc7405d9c03f`). Updated `compose.yaml` and verified backwards compatibility for all provider operations.
- **V004-DISC (Issue #18):** Executed automated heartbeat and label probe (`scripts/test_heartbeat_feasibility.py`) against GoAlert v0.35.0. Documented findings in `docs/research/heartbeat-feasibility.md`:
  - Enforced minimum timeout validation (`timeoutMinutes >= 5`).
  - Verified `updateHeartbeatMonitor` mutation for in-place updates of `name` and `timeoutMinutes`, with immutable `serviceID` (`RequiresReplace`).
  - Verified ping delivery via HTTP POST and GET returning HTTP 200.
  - Characterized `setServiceLabel` with domain prefix requirement (`<domain-prefix>/<suffix>`) and deletion via empty string value.
- **V004-SPEC (Issue #32):** Authored specifications and user-facing documentation:
  - Technical specs: `docs/specs/goalert-heartbeat-monitor.md`, `docs/specs/goalert-service-label.md`, `docs/specs/goalert-data-sources.md`.
  - Provider documentation: `docs/resources/heartbeat_monitor.md`, `docs/resources/service_label.md`, and `docs/data-sources/` (`service.md`, `escalation_policy.md`, `integration_key.md`, `heartbeat_monitor.md`).
- **V004-IMPL (Issue #33):** Implemented client methods, resources, and data sources:
  - Canonical GraphQL document `internal/client/operations.graphql` expanded with queries and mutations for heartbeats, labels, and exact-match searches.
  - Client library `internal/client/client.go` implemented type-safe methods with AST hash-pinned operations.
  - Resources implemented: `goalert_heartbeat_monitor` (`internal/provider/heartbeat_monitor_resource.go`) and `goalert_service_label` (`internal/provider/service_label_resource.go`).
  - Data sources implemented: `goalert_service`, `goalert_escalation_policy`, `goalert_integration_key`, and `goalert_heartbeat_monitor`.
  - Registered all new resources and data sources in `internal/provider/provider.go`.
- **V004-VERIFY (Issue #34):**
  - Added unit test suites for client and provider components (`internal/client/client_test.go`, `internal/provider/heartbeat_monitor_resource_test.go`, `internal/provider/service_label_resource_test.go`, `internal/provider/data_sources_test.go`).
  - Implemented end-to-end acceptance test `v004_acceptance` in `scripts/acceptance.py`:
    - Full resource creation (`goalert_heartbeat_monitor`, `goalert_service_label`).
    - Webhook ping delivery to heartbeat monitor `href` verifying HTTP 200.
    - Querying and attribute resolution across all 4 data sources.
    - In-place mutation of monitor and label attributes without resource replacement.
    - Remote drift detection and recreation upon out-of-band deletion.
    - Import support (standard UUID for heartbeats, compound `<service_id>/<key>` for labels).
    - Upstream AST key migration validation (outdated v0.0.3 key rejected with zero state corruption).
    - Clean teardown of remote resources.
  - 100% of unit tests and end-to-end acceptance tests passed cleanly.


- Conducted exhaustive mapping of GoAlert upstream GraphQL schema (`target/goalert`) to Terraform provider resources and data sources.
- Established phased SpecDD roadmap from v0.0.4 through v1.0.0 (Registry publication):
  - v0.0.4: Heartbeat monitors, service labels, foundational data sources (`service`, `escalation_policy`, `integration_key`, `heartbeat_monitor`), and GoAlert v0.35.0 compatibility.
  - v0.0.5: User identity, contact methods, and individual notification rules.
  - v0.0.6: On-call shift rotations and escalation policy target expansion (users and rotations).
  - v0.0.7: Schedules, rules, shifts, overrides, and channel on-call notifications.
  - v0.0.8: Collaboration integrations (Slack channels/groups) and system limits.
  - v1.0.0: Public Terraform Registry launch (`tfplugindocs`, GPG signing, public repository transition, GA release).
- Created GitHub Milestone 4 (`v0.0.4`) and linked issues:
  - Issue #22: Upstream GoAlert v0.35.0 compatibility (Gate V004-UPSTREAM).
  - Issue #18: Heartbeat monitor lifecycle investigation (Gate V004-DISC).
  - Issue #32: Heartbeat monitor, service label, and data source contract specification (Gate V004-SPEC).
  - Issue #33: Heartbeat monitor, service label, and data source implementation (Gate V004-IMPL).
  - Issue #34: Lifecycle verification and acceptance testing (Gate V004-VERIFY).
- Updated canonical documents: `docs/roadmap.md`, `docs/milestones/heartbeat-monitors-and-service-labels.md`, and `docs/sprints/sprint-3.md`.

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
- Updated `docs/milestones/escalation-policies-and-webhook-routing.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.

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
- Created [v0.0.2 planning](docs/milestones/escalation-policies-and-webhook-routing.md), linked issues #4-#8 and
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
- Updated `docs/milestones/escalation-policies-and-webhook-routing.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.
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
- Updated `docs/research/alert-delivery-feasibility.md`, `docs/milestones/escalation-policies-and-webhook-routing.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.
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
- Updated `docs/milestones/escalation-policies-and-webhook-routing.md`, `docs/sprints/sprint-1.md`, and `docs/roadmap.md`.
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

## 2026-09-27 - Implement v0.0.5 users, contact methods, and notification rules

- Goal: implement human operator identity and user notification management in GoAlert provider for Milestone `v0.0.5` covering issues #38 (V005-DISC), #39 (V005-SPEC), #40 (V005-IMPL), and #41 (V005-VERIFY).
- Research & Feasibility (Gate V005-DISC):
  - Probed GoAlert v0.35.0 schema via `scripts/test_user_feasibility.py` (`docs/research/user-feasibility.md`).
  - Identified requirement for `username` and `password` on user creation.
  - Proved contact method immutability of `value` in GoAlert (`cannot update value` GraphQL error on update).
  - Determined fallback search behavior for `data.goalert_user` (search filter matches name; exact email matches require full user listing scan).
- Specifications & Documentation (Gate V005-SPEC):
  - Created resource specifications: `docs/specs/goalert-user.md`, `docs/specs/goalert-user-contact-method.md`, `docs/specs/goalert-user-notification-rule.md`, `docs/specs/goalert-user-data-source.md`.
  - Created user-facing documentation: `docs/resources/user.md`, `docs/resources/user_contact_method.md`, `docs/resources/user_notification_rule.md`, `docs/data-sources/user.md`.
- Implementation (Gate V005-IMPL):
  - Added GraphQL queries and mutations to `internal/client/operations.graphql`.
  - Implemented client methods in `internal/client/client.go` with full mock unit tests in `internal/client/client_test.go`.
  - Implemented `goalert_user` resource (`internal/provider/user_resource.go`) with in-place updates for name/email/role, replacement on username/password change, and compound `<user_id>/<username>` or bare `<user_id>` import.
  - Implemented `goalert_user_contact_method` resource (`internal/provider/user_contact_method_resource.go`) with `RequiresReplace()` on `value` and `type`, in-place update for `name`, and compound `<user_id>/<cm_id>` or bare `<cm_id>` import.
  - Implemented `goalert_user_notification_rule` resource (`internal/provider/user_notification_rule_resource.go`) as an immutable resource requiring replacement on any modification, and compound `<user_id>/<rule_id>` import.
  - Implemented `goalert_user` data source (`internal/provider/user_data_source.go`) supporting lookup by `id`, exact `name`, or exact `email`.
  - Registered resources and data source in `internal/provider/provider.go`.
- Verification (Gate V005-VERIFY):
  - Unit tests: `go test -v ./...` passed 100%.
  - Acceptance tests: `python scripts/acceptance.py --suite all` passed 100% against real GoAlert v0.35.0 container across all milestones (PoC, Service, Policy, Integration Key, Heartbeats/Labels/Data Sources, and Users/Contact Methods/Notification Rules).
  - Static checks: `python scripts/check_format.py` and `git diff --check` passed cleanly.
