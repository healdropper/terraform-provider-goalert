# Specification: Escalation Policy Step Targets Contract

Status: Accepted contract for Milestone `Rotations and Escalation Targets` (Issue #44 / Gate ROT-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/rotation-feasibility.md).

## 1. Scope and Design

Expands the `step` nested block of the `goalert_escalation_policy` resource and data source to support human operators (`user_ids`) and on-call rotations (`rotation_ids`) alongside existing webhook actions (`webhook_action`).

## 2. Updated Step Block Schema

Within `resource "goalert_escalation_policy"`:

| Field | Type | Requirement | Description |
| --- | --- | --- | --- |
| `delay_minutes` | Int64 | Optional, Computed | Delay before escalating to the next step. Default `15`. |
| `webhook_action` | Block List | Optional | Webhook notification URLs (`url`). |
| `user_ids` | List(String) | Optional | Set/list of operator user IDs to notify in this step. |
| `rotation_ids` | List(String) | Optional | Set/list of rotation IDs to notify in this step. |

### Step Validation Rule
Each step must specify at least one target or action:
`len(webhook_action) > 0 || len(user_ids) > 0 || len(rotation_ids) > 0`.

## 3. Mapping to GraphQL Operations

1. **Step Creation and Update**:
   - `CreateEscalationPolicyStepInput` / `UpdateEscalationPolicyStepInput`:
     - `targets`:
       - Append `{ id: user_id, type: "user" }` for each `user_ids` entry.
       - Append `{ id: rotation_id, type: "rotation" }` for each `rotation_ids` entry.
     - `actions`:
       - Append `{ type: "webhook", args: { "url": url } }` for each `webhook_action` entry.
2. **Step Read**:
   - Parse `step.actions` from GraphQL response:
     - If `action.type == "builtin-user"`: extract `action.args.user_id` into `user_ids`.
     - If `action.type == "builtin-rotation"`: extract `action.args.rotation_id` into `rotation_ids`.
     - If `action.type == "webhook"`: extract `action.args.url` into `webhook_action`.
