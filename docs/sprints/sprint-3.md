# Sprint 3: heartbeat monitors, labels, and data sources

Owner: healdropper. Recorded: 2026-09-27.
Status: Completed delivery sprint. [Milestone v0.0.4](../milestones/v0.0.4.md) is implemented and verified;
[GitHub milestone 4](https://github.com/healdropper/terraform-provider-goalert/milestone/4) tracks delivery.

## Goal and user story

As a GoAlert operator, declare dead-man switch heartbeat monitors on managed services to detect silent daemon/cron failures, assign labels to categorize services, and query existing GoAlert infrastructure via Terraform data sources without monolithic state coupling. Verify full compatibility with upstream GoAlert v0.35.0.

## Scope and sequence

1. **Upstream Assessment (V004-UPSTREAM, Issue #22):**
   - Review GoAlert v0.35.0 GraphQL schema diff and release changes.
   - Update `compose.yaml` and `scripts/fixture.py` to `goalert/goalert:v0.35.0`.
   - Verify AST hash compatibility for existing queries in `internal/client/operations.graphql`.

2. **Discovery (V004-DISC, Issue #18):**
   - Probe suite against disposable GoAlert v0.35.0 container.
   - Prove `createHeartbeatMonitor`, `updateHeartbeatMonitor`, `deleteAll` (type: `heartbeatMonitor`).
   - Prove ping URL generation, format, and sensitivity.
   - Prove `setLabel` mutation and service label query capabilities.
   - Prove single-resource and search query behavior for `Service`, `EscalationPolicy`, `IntegrationKey`, and `HeartbeatMonitor`.

3. **Specification (V004-SPEC, Issue #32):**
   - Resolve canonical specifications for `goalert_heartbeat_monitor` and `goalert_service_label`.
   - Resolve schemas for data sources: `goalert_service`, `goalert_escalation_policy`, `goalert_integration_key`, and `goalert_heartbeat_monitor`.
   - Define immutability rules, compound import syntax, and sensitive attribute redactions.

4. **Implementation (V004-IMPL, Issue #33):**
   - Implement resources and data sources using Terraform Plugin Framework in `internal/provider/`.
   - Update GraphQL client in `internal/client/` with heartbeat and label operations.

5. **Verification (V004-VERIFY, Issue #34):**
   - Add unit tests in `internal/provider/` and `internal/client/`.
   - Add acceptance tests in `scripts/acceptance.py` covering full CRUD, drift repair, import, and data source resolution against live GoAlert container.

## Exclusions

User identity management, rotations, on-call schedules, and Slack channel bindings remain scheduled for subsequent milestones (v0.0.5 through v0.0.8).

## DoD and evidence

All V004-UPSTREAM/DISC/SPEC/IMPL/VERIFY gates satisfied with verifiable evidence.
Passing unit tests, acceptance tests across Terraform v1.10.5 and v1.16.3, and multi-architecture packaging.
Milestone v0.0.4 closed on GitHub.
