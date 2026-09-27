# Rotations and Escalation Targets

Owner: healdropper. Recorded: 2026-09-27.
Status: Planned and tracked under [GitHub Milestone 6](https://github.com/healdropper/terraform-provider-goalert/milestone/6).

## Recorded owner decisions

- Implement on-call shift rotation management via `goalert_rotation` (name, description, type, start_time, time_zone, shift_length, user_ids).
- Provide data source `data.goalert_rotation` for querying rotations by ID or exact name.
- Expand `goalert_escalation_policy` steps to support human and rotation targets (`user_id`, `rotation_id`), resolving the limitation where only webhook steps were supported.

## Planning envelope

Proposed outcomes:
1. Declare rotations with rotation types (`daily`, `weekly`, `hourly`), configurable shift length, timezone, and ordered participant list (`user_ids`).
2. Query existing rotation records via data source.
3. Target rotations and users directly in `goalert_escalation_policy` step blocks.
4. Verify complete lifecycle via disposable GoAlert acceptance tests.

| ID | Planning/acceptance gate | Evidence required before closing |
| --- | --- | --- |
| ROT-DISC | Establish rotation and escalation step target API feasibility | Probe suite against disposable GoAlert v0.35.0 verifying `createRotation`, `updateRotation`, `deleteRotation`, and `escalationPolicy` step mutations with user and rotation targets |
| ROT-SPEC | Resolve and accept resource and data source contracts | Canonical specification in `docs/specs/` defining schemas, validations, compound import syntax, and immutability rules |
| ROT-IMPL | Implement accepted contracts | `goalert_rotation` resource, `data.goalert_rotation` data source, and expanded `goalert_escalation_policy` step targets in `internal/provider/` with expanded client methods |
| ROT-VERIFY | Establish provider lifecycle and acceptance confidence | Unit tests and real disposable GoAlert acceptance tests in `scripts/acceptance.py` verifying full CRUD, participant list ordering, step targets, import, and drift repair |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| ROT-DISC | [#43: feat: investigate rotation and escalation step target API feasibility](https://github.com/healdropper/terraform-provider-goalert/issues/43) | In progress |
| ROT-SPEC | [#44: docs: resolve rotation and escalation step target contracts](https://github.com/healdropper/terraform-provider-goalert/issues/44) | Pending DISC |
| ROT-IMPL | [#45: feat: implement rotation resource, data source, and escalation step targets](https://github.com/healdropper/terraform-provider-goalert/issues/45) | Pending SPEC |
| ROT-VERIFY | [#46: test: verify rotation and escalation step target lifecycles](https://github.com/healdropper/terraform-provider-goalert/issues/46) | Pending IMPL |
