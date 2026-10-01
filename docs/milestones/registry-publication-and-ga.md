# Registry Publication and GA (`v1.0.0`)

Owner: healdropper. Recorded: 2026-10-01.
Status: Planned and tracked under [GitHub Milestone 10](https://github.com/healdropper/terraform-provider-goalert/milestone/10).

## Context and Scope

With Milestones 1 through 9 complete, `terraform-provider-goalert` provides 100% declarative resource and data-source coverage for GoAlert v0.35.0:
- **14 Managed Resources**: `goalert_service`, `goalert_escalation_policy` (with `multi_ack` and user/rotation/schedule/webhook actions), `goalert_integration_key`, `goalert_heartbeat_monitor`, `goalert_service_label`, `goalert_label` (polymorphic across services, policies, schedules, and rotations), `goalert_user`, `goalert_user_contact_method` (with `enable_status_updates` and `private`), `goalert_user_notification_rule`, `goalert_rotation`, `goalert_schedule`, `goalert_schedule_rule`, `goalert_user_override`, and `goalert_system_limit`.
- **9 Data Sources**: `goalert_service`, `goalert_escalation_policy`, `goalert_integration_key`, `goalert_heartbeat_monitor`, `goalert_user`, `goalert_rotation`, `goalert_schedule`, `goalert_slack_channel`, and `goalert_slack_user_group`.

Milestone 10 prepares and executes the public open-source exposure and **v1.0.0 General Availability (GA)** release to the public HashiCorp Terraform Registry under the strict public repository governance rules established in Doctrine Foundry v1.11.0:
1. **Privacy & Git History Sanitization**: Audit and purge all references to private infrastructure, internal clusters, or private notification channels from the working tree, commit history, and GitHub pull request/issue metadata prior to public exposure.
2. **Self-Hosted Runner Isolation & Fork PR Gating**: Verify zero self-hosted runners (`total_count: 0`), cloud-only workflow targets, and enforce `all_external_contributors` fork PR workflow approval upon public transition so `Integration` runs automatically on maintainer PRs while blocking external fork PRs until explicitly approved (`Approve and run`).
3. **Default-Branch (`main`) Protection & Community Governance**: Enforce rebase-only merge settings, add `.github/CODEOWNERS`, `CONTRIBUTING.md`, and `SECURITY.md`, and immediately apply a `main` branch protection ruleset and Private Vulnerability Reporting upon public transition.
4. **Signed `v1.0.0` Release, Registry Publication, and Consumer Adoption**: Validate Terraform Registry documentation structure, publish a GPG-signed `v1.0.0` release, register `healdropper/goalert` in the Terraform Registry, and upgrade the production consumer deployment to GoAlert v0.35.0 with the published `v1.0.0` provider.

## Planning envelope

| ID | Planning/acceptance gate | Evidence required before closing |
| --- | --- | --- |
| GA-DISC | Audit privacy, git history, Registry docs, and `v1.0.0` release contract | Complete audit report in `docs/research/public-release-and-registry-audit.md` enumerating all files/commits/PRs to sanitize, Registry doc verification, and backlog issue cleanup |
| GA-SPEC | Specify public repository governance, community files, and `v1.0.0` release contract | Accepted `docs/specs/public-repository-governance.md`, `CONTRIBUTING.md`, `SECURITY.md`, `.github/CODEOWNERS`, and updated `docs/releases.md`, `docs/index.md`, and `README.md` |
| GA-IMPL | Sanitize private references and history, enforce repo merge settings, and update release workflow for `v1.0.0` | Zero private references in `git grep` or `git log`, updated `.github/workflows/release.yml` for `v*` SemVer tags, Registry doc lint in `scripts/check_format.py`, and rebase-only repo settings verified via GitHub API |
| GA-VERIFY | Transition to public with `main` ruleset & fork PR gating, publish signed `v1.0.0` to Terraform Registry, and adopt in consumer | Public visibility + active `main` ruleset + `all_external_contributors` verified via API, signed `v1.0.0` release published on Terraform Registry, and GoAlert v0.35.0 + `v1.0.0` provider verified in production consumer |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| GA-DISC | [#79: chore(ga): audit repository privacy, git history, Registry docs, and v1.0.0 release contract](https://github.com/healdropper/terraform-provider-goalert/issues/79) | Completed (`docs/research/public-release-and-registry-audit.md`) |
| GA-SPEC | [#80: docs(ga): specify public repository governance, community files, and v1.0.0 release process](https://github.com/healdropper/terraform-provider-goalert/issues/80) | Ready |
| GA-IMPL | [#81: feat(ga): sanitize private references, add governance files, and enable v1.0.0 release workflow](https://github.com/healdropper/terraform-provider-goalert/issues/81) | Pending SPEC |
| GA-VERIFY | [#82: release(ga): transition to public with main ruleset, publish signed v1.0.0 to Registry, and adopt in consumer](https://github.com/healdropper/terraform-provider-goalert/issues/82) | Pending IMPL |
