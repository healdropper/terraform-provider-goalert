# Sprint 2: ingress and integration keys

Owner: healdropper. Recorded: 2026-09-26.
Status: Completed delivery sprint. [Milestone Ingress and Integration Keys](../milestones/ingress-and-integration-keys.md) is closed;
[private board](https://github.com/users/healdropper/projects/1) tracks delivery.

## Goal and user story

As a GoAlert operator, declare integration keys on managed services to allow Grafana
and other alert sources to send alerts into GoAlert via dedicated webhook endpoints.
Connected with `@HornOfCenariusBot` and channel `Cenarion Watch` to complete the alert routing path.

## Scope and sequence

1. **Discovery (V003-DISC, Issue #9):**
   - Probe suite executed against disposable GoAlert v0.34.1 container (`scripts/test_ingress_key_feasibility.py`).
   - Verified `createIntegrationKey` with types (`grafana`, `generic`).
   - Proved key URL generation, token format (`token=<id>`), and sensitive `href`.
   - Proved user role permissions and deletion lifecycle via `deleteAll`.
   - Documented findings in `docs/research/integration-key-feasibility.md`. Closed in PR #28.

2. **Specification (V003-SPEC, Issue #24):**
   - Resolved canonical specification in `docs/specs/goalert-integration-key.md` and reference in `docs/resources/integration_key.md`.
   - Defined schema: `id`, `service_id` (RequiresReplace), `name` (RequiresReplace), `type` (default `"grafana"`, RequiresReplace), `href` (sensitive).
   - Documented token migration and compound import syntax (`<service_id>/<key_id>`). Closed in PR #29.

3. **Implementation (V003-IMPL, Issue #25):**
   - Implemented `goalert_integration_key` resource using Terraform Plugin Framework (`internal/provider/integration_key_resource.go`).
   - Updated GraphQL client in `internal/client/` with integration key operations and registered resource in provider. Closed in PR #30.

4. **Verification (V003-VERIFY, Issue #26):**
   - Added unit tests in `client_test.go` and `integration_key_resource_test.go`.
   - Added `integration_key_acceptance` in `scripts/acceptance.py` verifying CRUD, Grafana payload ingestion, replacement on attribute changes, drift recreation, compound import, and old API key rejection without state loss. Closed in PR #30.

## Exclusions

Consumer Grafana rule/contact point deployment and Telegram relay service deployment in `cenarion-watch`
follow provider readiness in a separate consumer PR. Schedules, rotations, and heartbeat monitors remain Backlog.

## DoD and evidence

All V003-DISC/SPEC/IMPL/VERIFY gates satisfied with verifiable evidence.
Passing unit tests, acceptance tests across Terraform v1.10.5 and v1.16.3, and multi-architecture packaging.
Milestone v0.0.3 closed on GitHub.
