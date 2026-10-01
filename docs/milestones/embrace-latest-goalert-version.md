# Embrace Latest GoAlert Version

Owner: healdropper. Recorded: 2026-10-01.
Status: Planned and tracked under [GitHub Milestone 9](https://github.com/healdropper/terraform-provider-goalert/milestone/9).

## Context and Upstream Delta (`v0.34.1...v0.35.0`)

While [Issue #22](https://github.com/healdropper/terraform-provider-goalert/issues/22) upgraded the disposable container image to `goalert/goalert:v0.35.0` and verified backwards compatibility for pre-existing operations, GoAlert v0.35.0 introduced new GraphQL schema capabilities (`graphql2/schema.graphql`, `graphql2/graph/escalationpolicy.graphqls`) that must be exposed in the provider to achieve 100% v0.35.0 API coverage:

1. **Multi-Ack Escalation Policy Steps (`multiAck`)**:
   - `EscalationPolicyStep.multiAck: Boolean!`
   - `CreateEscalationPolicyStepInput.multiAck: Boolean`
   - `UpdateEscalationPolicyStepInput.multiAck: Boolean`
   - Provider mapping: `multi_ack` attribute inside `goalert_escalation_policy` `step` block (`Optional`, `Computed`, default `false`).
2. **Private Contact Methods and Status Updates (`private`, `enableStatusUpdates`)**:
   - `UserContactMethod.private: Boolean!` and `statusUpdates: StatusUpdateState!`
   - `CreateUserContactMethodInput.private: Boolean` and `enableStatusUpdates: Boolean`
   - `UpdateUserContactMethodInput.private: Boolean` and `enableStatusUpdates: Boolean`
   - Provider mapping: `enable_status_updates` attribute (and `private` where compatible with non-owner API keys) on `goalert_user_contact_method`.
3. **Polymorphic Labels for Escalation Policies, Schedules, and Rotations (`labels` / `setLabel`)**:
   - `EscalationPolicy.labels: [Label!]!`, `Schedule.labels: [Label!]!`, `Rotation.labels: [Label!]!`
   - `setLabel(input: SetLabelInput!)` now accepts `target: { type: service | escalationPolicy | schedule | rotation, id: <uuid> }`.
   - Provider mapping: polymorphic `goalert_label` resource (`target_type`, `target_id`, `key`, `value`) while preserving `goalert_service_label` compatibility.

## Planning envelope

| ID | Planning/acceptance gate | Evidence required before closing |
| --- | --- | --- |
| UPST-DISC | Establish GoAlert v0.35.0 schema additions feasibility | Empirical probe against disposable GoAlert v0.35.0 characterizing `multiAck`, `private` contact method visibility under admin API keys, `enableStatusUpdates`, and `setLabel` on policies/schedules/rotations |
| UPST-SPEC | Resolve and accept v0.35.0 resource contracts | Updated canonical specifications in `docs/specs/` and user-facing Registry docs in `docs/resources/` |
| UPST-IMPL | Implement accepted v0.35.0 contracts | Updated `internal/client/operations.graphql`, `internal/client/client.go`, `goalert_escalation_policy` (`multi_ack`), `goalert_user_contact_method`, and `goalert_label` resource with unit tests |
| UPST-VERIFY | Establish provider lifecycle and acceptance confidence | End-to-end acceptance suite `v035_acceptance` in `scripts/acceptance.py` verifying CRUD, in-place updates, drift repair, and imports against disposable GoAlert v0.35.0 |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| UPST-DISC | [#70: feat: investigate GoAlert v0.35.0 GraphQL schema additions and API behavior](https://github.com/healdropper/terraform-provider-goalert/issues/70) | In progress |
| UPST-SPEC | [#71: docs: resolve GoAlert v0.35.0 resource and schema contracts](https://github.com/healdropper/terraform-provider-goalert/issues/71) | Pending DISC |
| UPST-IMPL | [#72: feat: implement GoAlert v0.35.0 schema capabilities in provider](https://github.com/healdropper/terraform-provider-goalert/issues/72) | Pending SPEC |
| UPST-VERIFY | [#73: test: verify GoAlert v0.35.0 capabilities end-to-end](https://github.com/healdropper/terraform-provider-goalert/issues/73) | Pending IMPL |
