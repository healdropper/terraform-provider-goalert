# Sprint 4: human identity, contact methods, and notification rules

Owner: healdropper. Recorded: 2026-09-27.
Status: Sprint completed. All V005-DISC/SPEC/IMPL/VERIFY gates satisfied with verifiable evidence;
[GitHub milestone 5](https://github.com/healdropper/terraform-provider-goalert/milestone/5) ready for closure.

## Goal and user story

As a GoAlert administrator and infrastructure engineer, declare operator user accounts (`goalert_user`), configure communication channels (`goalert_user_contact_method` for voice, SMS, email, webhook), and define personal notification rules (`goalert_user_notification_rule` with delay minutes) through Terraform. Enable decoupled lookups via `data.goalert_user`.

## Scope and sequence

1. **Discovery (V005-DISC, Issue #38):**
   - Probe suite against disposable GoAlert v0.35.0 container.
   - Prove `createUser`, `updateUser`, `deleteAll` (type: `user`).
   - Determine role enum and permissions (`user`, `admin`).
   - Prove `createContactMethod`, `updateContactMethod`, `deleteAll` (type: `userContactMethod`).
   - Probe supported contact method types: `VOICE`, `SMS`, `EMAIL`, `WEBHOOK`.
   - Probe contact method verification status handling in GoAlert.
   - Prove `createUserNotificationRule`, `updateUserNotificationRule`, `deleteAll` (type: `userNotificationRule`).
   - Probe delay constraints (e.g. 0 to 9000 minutes) and contact method references.
   - Prove `Query.user(id: ID!)` and `Query.users(search: String!)` exact lookup behavior.

2. **Specification (V005-SPEC, Issue #39):**
   - Resolve canonical specifications for:
     - `goalert_user`
     - `goalert_user_contact_method`
     - `goalert_user_notification_rule`
     - `data.goalert_user`
   - Define immutability rules, compound import formats (`<user_id>/<contact_method_id>`, `<user_id>/<rule_id>`), and validation constraints.

3. **Implementation (V005-IMPL, Issue #40):**
   - Expand canonical operations document `internal/client/operations.graphql`.
   - Implement type-safe client methods in `internal/client/client.go`.
   - Implement resources and data source using Terraform Plugin Framework in `internal/provider/`.
   - Register components in `internal/provider/provider.go`.

4. **Verification (V005-VERIFY, Issue #41):**
   - Add unit tests in `internal/client/` and `internal/provider/`.
   - Add end-to-end acceptance test `v005_acceptance` in `scripts/acceptance.py` verifying full CRUD, in-place updates, drift repair, compound imports, and AST key migration against disposable GoAlert v0.35.0.

## Exclusions

Escalation policy user targets and rotation definitions remain scheduled for Milestone v0.0.6. Schedule rules and channel notifications remain in v0.0.7 / v0.0.8.

## DoD and evidence

All V005-DISC/SPEC/IMPL/VERIFY gates satisfied with verifiable evidence.
Passing unit tests, acceptance tests across Terraform v1.10.5 and v1.16.3, and multi-architecture packaging.
Milestone v0.0.5 closed on GitHub.
