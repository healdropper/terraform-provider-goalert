# User, contact method, and notification rule API feasibility

Recorded: 2026-09-27.
Authority: [Milestone v0.0.5](../milestones/v0.0.5.md), Gate V005-DISC, [Issue #38](https://github.com/healdropper/terraform-provider-goalert/issues/38).
Upstream version probed: GoAlert v0.35.0 (`sha256:f090a90538d7e61446aad245c21a7a1694d36d7b4f6c56fc55e9cc7405d9c03f`).

## Executive summary

Automated feasibility testing via `scripts/test_user_feasibility.py` against a disposable GoAlert v0.35.0 instance confirmed full GraphQL schema compatibility, lifecycle behaviors, mutations, and query interfaces for human operator identity management, contact channels, and user-level notification escalation rules.

---

## 1. User identity (`goalert_user`)

### Mutations
- `createUser(input: CreateUserInput!) -> User`:
  - `username: String!` (Required for basic auth credentialing in GoAlert).
  - `password: String!` (Required; must be sensitive in Terraform).
  - `name: String` (Operator full name).
  - `email: String` (Operator email).
  - `role: UserRole` (Enum: `user` or `admin`, defaults to `user`).
  - `favorite: Boolean` (Optional).
- `updateUser(input: UpdateUserInput!) -> Boolean`:
  - In-place mutation of `name`, `email`, and `role`.
  - In-place mutation of `statusUpdateContactMethodID`.
  - `username` and `password` cannot be modified via `updateUser` (changing `username` requires replacement).
- `deleteAll(input: [{type: user, id: ID!}]) -> Boolean`:
  - Atomically deletes the user and automatically cascade-deletes all associated contact methods, notification rules, and user sessions.

### Queries
- `user(id: ID!) -> User`:
  - Returns direct `User` object or `null` if deleted (O(1) drift detection).
  - Fields: `id`, `name`, `email`, `role`, `contactMethods`, `notificationRules`.
- `users(input: {search: String, first: Int}) -> UserConnection`:
  - Used for data source lookups by exact name or email.

---

## 2. Contact methods (`goalert_user_contact_method`)

### Mutations
- `createUserContactMethod(input: CreateUserContactMethodInput!) -> UserContactMethod`:
  - `userID: ID!` (Required; immutable, `RequiresReplace: true`).
  - `name: String!` (Channel label, e.g. "Work Email", "Ops Webhook", "Cell SMS").
  - `type: ContactMethodType` (Enum: `SMS`, `VOICE`, `EMAIL`, `WEBHOOK`, `SLACK_DM`).
  - `value: String` (Phone number in E.164, email address, or webhook URL).
  - `enableStatusUpdates: Boolean` (Optional).
- `updateUserContactMethod(input: UpdateUserContactMethodInput!) -> Boolean`:
  - Supports in-place updates of `name`, `value`, `enableStatusUpdates`.
  - `type` cannot be modified in-place (`RequiresReplace: true`).
- `deleteAll(input: [{type: contactMethod, id: ID!}]) -> Boolean`:
  - Note: TargetType enum value is `'contactMethod'` (not `'userContactMethod'`).

### Queries
- `userContactMethod(id: ID!) -> UserContactMethod`:
  - Direct top-level query on `Query` for O(1) state refresh and drift detection.
  - Returns `null` on deleted contact methods.
  - Fields: `id`, `name`, `disabled`, `pending`, `dest { type args }`.

---

## 3. Notification rules (`goalert_user_notification_rule`)

### Mutations
- `createUserNotificationRule(input: CreateUserNotificationRuleInput!) -> UserNotificationRule`:
  - `userID: ID` (Target user).
  - `contactMethodID: ID` (Associated contact method).
  - `delayMinutes: Int!` (Non-negative delay in minutes, e.g. 0 for immediate alert, 5, 10, 15).
- In-place mutation:
  - GoAlert does not provide an `updateUserNotificationRule` mutation; notification rules are immutable and require recreation upon attribute modifications (`RequiresReplace: true`).
- `deleteAll(input: [{type: notificationRule, id: ID!}]) -> Boolean`:
  - Note: TargetType enum value is `'notificationRule'` (not `'userNotificationRule'`).

### Queries
- Queried via parent user: `user(id).notificationRules { id, delayMinutes, contactMethod { id } }`.

---

## 4. Acceptance verdict and implementation strategy

All capabilities required for Milestone v0.0.5 are natively supported upstream in GoAlert v0.35.0. No architectural blockers exist.
Proceed with Gate V005-SPEC to finalize schema contracts.
