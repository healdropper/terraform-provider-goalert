# Rotation and Escalation Target Feasibility Probe

Status: Complete empirical evidence for Gate ROT-DISC (Issue #43).
Executed: 2026-09-27 against disposable GoAlert v0.35.0 (`scripts/test_rotation_feasibility.py`).

## 1. Rotation GraphQL Schema Characterization

### Types and Enumerations
- **`RotationType`**: `daily`, `weekly`, `hourly`.
- **`TargetType`**: `user`, `rotation`, `schedule`, `chanWebhook`, `slackChannel`, `slackUserGroup`, `escalationPolicy`, etc.

### Rotation Schema Fields
- `id: ID!`
- `name: String!` (validated by GoAlert: can only contain letters, digits, hyphens, underscores, apostrophes, and spaces: `^[a-zA-Z0-9\-_' ]+$`).
- `description: String!`
- `type: RotationType!` (`daily`, `weekly`, `hourly`).
- `start: ISOTimestamp!` (ISO-8601 formatted timestamp).
- `timeZone: String!` (valid IANA timezone name, e.g. `Europe/Madrid`, `UTC`, `America/New_York`).
- `shiftLength: Int!` (duration multiplier for the rotation type, e.g. 1 day, 7 days, 12 hours).
- `userIDs: [ID!]!` (ordered sequence of operator IDs participating in the rotation).
- `users: [User!]!` (resolved user objects for each participant).
- `activeUserIndex: Int!` (zero-indexed position of the currently active participant).
- `nextHandoffTimes: [ISOTimestamp!]!`

### Mutations
- `createRotation(input: CreateRotationInput!): Rotation!`
- `updateRotation(input: UpdateRotationInput!): Boolean!`
  - All attributes (`name`, `description`, `timeZone`, `start`, `type`, `shiftLength`, `userIDs`, `activeUserIndex`) are mutable in-place.
  - Participant order can be rearranged in-place by passing an updated `userIDs` list.
- `deleteAll(input: [{type: rotation, id: $id}]): Boolean!`

## 2. Escalation Policy Step Targets Characterization

### Input Mapping
In `CreateEscalationPolicyStepInput` and `UpdateEscalationPolicyStepInput`:
- `targets: [TargetInput!]` where `TargetInput = { id: ID!, type: TargetType! }`.
  - For user targets: `{ id: <user_id>, type: "user" }`
  - For rotation targets: `{ id: <rotation_id>, type: "rotation" }`
  - For schedule targets: `{ id: <schedule_id>, type: "schedule" }`
- `actions: [ActionInput!]`
  - For webhook actions: `{ type: "webhook", args: { "url": <url> } }`

### Output Representation on EscalationPolicyStep
`EscalationPolicyStep.actions` returns a unified list of actions and targets:
- Builtin User: `{ type: "builtin-user", args: { "user_id": "<UUID>" } }`
- Builtin Rotation: `{ type: "builtin-rotation", args: { "rotation_id": "<UUID>" } }`
- Builtin Schedule: `{ type: "builtin-schedule", args: { "schedule_id": "<UUID>" } }`
- Webhook: `{ type: "webhook", args: { "url": "<URL>" } }`

## 3. Terraform Resource Modeling Implications

1. **`goalert_rotation` Resource**:
   - `id` (Computed, String)
   - `name` (Required, String, validated with `^[a-zA-Z0-9\-_' ]+$`)
   - `description` (Optional, Computed, String, default `""`)
   - `type` (Required, String, validated: `daily`, `weekly`, `hourly`)
   - `start_time` (Required, String, ISO-8601)
   - `time_zone` (Required, String, IANA timezone)
   - `shift_length` (Optional, Computed, Int64, default 1)
   - `user_ids` (Required, List of String, ordered)
   - `active_user_index` (Computed, Int64)
   - Import syntax: `<rotation_id>`

2. **`data.goalert_rotation` Data Source**:
   - Query by `id` or exact `name`.
   - Exports all rotation attributes including `user_ids`.

3. **`goalert_escalation_policy` Step Target Expansion**:
   - In `step` block, expand beyond `webhook_action` to support:
     - `target` blocks or explicit target attributes:
       - `user_ids` (List of String, Optional)
       - `rotation_ids` (List of String, Optional)
       - `webhook_action` (Block list, Optional)
     - Preserves 100% backwards compatibility with existing webhook configurations.
