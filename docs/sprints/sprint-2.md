# Sprint 2: ingress and integration keys

Owner: healdropper. Recorded: 2026-09-26.
Status: Active delivery sprint. [Milestone v0.0.3](../milestones/v0.0.3.md) is the scope authority;
[private board](https://github.com/users/healdropper/projects/1) tracks delivery.

## Goal and user story

As a GoAlert operator, declare integration keys on managed services to allow Grafana
and other alert sources to send alerts into GoAlert via dedicated webhook endpoints.
Connected with `@HornOfCenariusBot` and channel `Cenarion Watch` to complete the alert routing path.

## Scope and sequence

1. **Discovery (V003-DISC, Issue #9):**
   Execute probe suite against disposable GoAlert v0.34.1 container:
   - Prove `createIntegrationKey` mutation with types (`grafana`, `generic`).
   - Probe key URL generation, token format, and sensitivity.
   - Prove role requirements (admin vs user) and deletion lifecycle.
   - Prove canonical operations GraphQL document hash alignment.

2. **Specification (V003-SPEC, Issue #24):**
   Resolve canonical specification in `docs/specs/goalert-integration-key.md`:
   - Schema: `id`, `service_id` (Required, string), `name` (Required, string), `type` (Optional, string, default `"grafana"`), `key` (Computed, sensitive string), `href` / webhook URL (Computed, string).
   - Lifecycle: immutable fields forcing replacement, import by `service_id/key_id`.

3. **Implementation (V003-IMPL, Issue #25):**
   Implement `goalert_integration_key` resource using Terraform Plugin Framework.
   Update GraphQL client in `internal/client/` with integration key operations.

4. **Verification (V003-VERIFY, Issue #26):**
   Unit tests and end-to-end acceptance suite against disposable GoAlert container.
   Verify CRUD, drift detection, replacement on attribute changes, and import.

## Exclusions

Consumer Grafana rule/contact point deployment and Telegram relay service deployment in `cenarion-watch`
follow provider readiness in a separate consumer PR. Schedules, rotations, and heartbeat monitors remain Backlog.

## DoD and evidence

All V003-DISC/SPEC/IMPL/VERIFY gates satisfied with verifiable evidence.
Passing unit tests, acceptance tests across Terraform v1.10.5 and v1.16.3, and multi-architecture packaging.
