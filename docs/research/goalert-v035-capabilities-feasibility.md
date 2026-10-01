# GoAlert v0.35.0 GraphQL Schema Capabilities Feasibility Probe

Status: Complete empirical evidence for Gate UPST-DISC ([Issue #70](https://github.com/healdropper/terraform-provider-goalert/issues/70)).
Executed: 2026-10-01 against disposable GoAlert v0.35.0 (`scripts/test_v035_feasibility.py`).

## 1. Upstream GraphQL Delta (`v0.34.1...v0.35.0`)

Inspection of `target/goalert` release `v0.35.0` (`graphql2/schema.graphql`, `graphql2/graph/escalationpolicy.graphqls`, and `graphql2/graph/signals.graphqls`) identified four GraphQL schema additions beyond `v0.34.1`:

1. **`multiAck` on Escalation Policy Steps**:
   - `EscalationPolicyStep.multiAck: Boolean!`
   - `CreateEscalationPolicyStepInput.multiAck: Boolean`
   - `UpdateEscalationPolicyStepInput.multiAck: Boolean`
2. **`private` and `enableStatusUpdates` on User Contact Methods**:
   - `UserContactMethod.private: Boolean!`
   - `UserContactMethod.statusUpdates: StatusUpdateState!` (`ENABLED`, `DISABLED`, `ENABLED_FORCED`, `DISABLED_FORCED`)
   - `CreateUserContactMethodInput.private: Boolean` and `enableStatusUpdates: Boolean`
   - `UpdateUserContactMethodInput.private: Boolean` and `enableStatusUpdates: Boolean`
3. **Polymorphic Labels on Escalation Policies, Schedules, and Rotations**:
   - `EscalationPolicy.labels: [Label!]!`
   - `Schedule.labels: [Label!]!`
   - `Rotation.labels: [Label!]!`
   - `setLabel(input: SetLabelInput!): Boolean!` where `input.target.type` accepts `service`, `escalationPolicy`, `schedule`, and `rotation`.
4. **`sendSignal` Mutation**:
   - `sendSignal(input: SendSignalInput!): Boolean!`
   - Gated behind experimental flag `UnivKeys` (`expflag.UnivKeys`) and represents an imperative runtime signal event rather than declarative infrastructure state. Excluded from Terraform resource state; its associated system limits (`PendingSignalsPerDestPerService`, `PendingSignalsPerService`) are already supported by `goalert_system_limit`.

## 2. Empirical Probe Findings

### 2.1 Multi-Ack Escalation Policy Steps (`multiAck`)
- Creating an escalation policy with `steps: [{ delayMinutes: 10, multiAck: true, actions: [...] }]` persists `multiAck: true` on the step.
- Updating the step in-place via `updateEscalationPolicyStep(input: { id: $stepID, delayMinutes: 15, multiAck: false })` transitions `multiAck` to `false` without recreating the step or policy.
- **Terraform Contract**: Add `multi_ack` (`Optional`, `Computed`, default `false`) to the `step` nested block of `goalert_escalation_policy`.

### 2.2 Private Contact Methods and Status Updates (`private`, `enableStatusUpdates`)
- Creating or updating a contact method with `private: true` succeeds on mutation, **however** GoAlert v0.35.0 enforces strict SQL-level visibility filtering in `CMStore.FindOne` and `CMStore.FindAll`:
  - Any contact method with `private = true` is hidden (`userContactMethod(id)` returns `null` and `user.contactMethods` omits it) unless the caller's session `UserID` matches the contact method's `userID`.
  - Even administrator sessions and system-level GraphQL API keys receive `null` when reading another user's private contact method.
- Consequently, setting `private = true` via a system/admin Terraform provider API key causes subsequent `Read` operations to see `null` (`ErrNotFound`), which would trigger infinite recreate loops in Terraform if configured for arbitrary users.
- By contrast, `private: false` (default) is readable by admin/system API keys and returns `"private": false` and `"statusUpdates": "ENABLED" | "DISABLED" | "ENABLED_FORCED" | "DISABLED_FORCED"`.
- **Terraform Contract**:
  - Expose `enable_status_updates` (`Optional`, `Computed`, default `false`) on `goalert_user_contact_method` for configuring status update delivery where supported by the destination type, plus `status_updates` (`Computed` String) and `private` (`Optional`, `Computed`, default `false`) with explicit documentation of GoAlert's owner-only read visibility constraint when `private = true`.

### 2.3 Polymorphic Labels (`setLabel` across `service`, `escalationPolicy`, `schedule`, `rotation`)
- `setLabel` accepts `target: { type: "escalationPolicy" | "schedule" | "rotation" | "service", id: "<uuid>" }` with `key` (`<domain>/<name>`) and `value`.
- Reading `escalationPolicy(id) { labels { key value } }`, `schedule(id) { labels { key value } }`, and `rotation(id) { labels { key value } }` returns the exact assigned labels.
- Setting `value: ""` deletes the label from the target cleanly.
- **Terraform Contract**:
  - Introduce polymorphic resource `goalert_label` with attributes:
    - `id` (Computed String: `<target_type>:<target_id>/<key>`)
    - `target_type` (Required String, `RequiresReplace`: `service`, `escalation_policy`, `schedule`, `rotation`)
    - `target_id` (Required String UUID, `RequiresReplace`)
    - `key` (Required String, `RequiresReplace`)
    - `value` (Required String, updatable in-place)
  - Retain `goalert_service_label` for backwards compatibility.
