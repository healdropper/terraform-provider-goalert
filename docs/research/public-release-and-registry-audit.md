# Public Release, Privacy, and Terraform Registry Audit (`v1.0.0`)

Owner: healdropper. Recorded: 2026-10-01.
Status: Completed for Milestone 10 (`GA-DISC`, Issue [#79](https://github.com/healdropper/terraform-provider-goalert/issues/79)).

## 1. Executive Summary

Before transitioning `healdropper/terraform-provider-goalert` from `private` to `public` and publishing `v1.0.0` on the public HashiCorp Terraform Registry, this audit inspects four mandatory surfaces required by the Public Repository Governance doctrine (`doctrine-foundry v1.11.0`):

1. **Working-Tree, Git History, and GitHub Metadata Privacy**: Enumerates every reference to private repositories, internal consumer environments, private bot handles, chat IDs, and local network addresses so they can be completely scrubbed in `GA-IMPL` (#81) prior to public exposure.
2. **Self-Hosted Runner & GitHub Actions Security**: Confirms zero self-hosted runners are registered on the repository, workflows run exclusively on GitHub-hosted cloud runners (`ubuntu-latest`, `windows-latest`), and external fork PR approval gating (`all_external_contributors`) is ready for activation upon public transition.
3. **Terraform Registry Documentation & Manifest Compliance**: Confirms 1:1 Registry documentation coverage across all 14 managed resources and 9 data sources plus `docs/index.md` and `terraform-registry-manifest.json` (Terraform Plugin Protocol v6.0).
4. **Release Workflow & Signing Readiness (`v1.0.0`)**: Identifies the `v0.0.*` tag filter in `.github/workflows/release.yml` that must be generalized to standard SemVer (`v[0-9]+\.[0-9]+\.[0-9]+`) for `v1.0.0`, and verifies the dedicated RSA-4096 GPG release signing identity in the `release` GitHub environment.

---

## 2. Privacy & Sanitization Inventory

### 2.1 Working-Tree Files Requiring Sanitization (`GA-IMPL`)

| File | Private / Internal Reference Kind | Remediation in `GA-IMPL` (#81) |
| :--- | :--- | :--- |
| `AGENTS.md` | References private organization owner in doctrine routing | Replace with repository-scoped contributor guidance (`healdropper/terraform-provider-goalert`) |
| `DEVLOG.md` | Mentions internal consumer repository and doctrine names | Generalize to neutral "downstream consumer deployment" and "repository governance doctrine" |
| `docs/roadmap.md` | Mentions internal consumer repository in Milestone 2 description | Replace with generic downstream consumer wording |
| `docs/milestones/escalation-policies-and-webhook-routing.md` | Mentions internal consumer repository | Replace with generic downstream consumer wording |
| `docs/milestones/ingress-and-integration-keys.md` | Mentions internal consumer repository, private Telegram bot handle, and negative group `chat_id` | Replace with RFC 2606 example values (`@ExampleAlertBot`, `-1001234567890`, and generic consumer wording) |
| `docs/specs/spec-driven-lifecycle.md` | Mentions internal doctrine codename | Replace with neutral "Spec-Driven Lifecycle" terminology |
| `docs/sprints/sprint-1.md` | Mentions internal consumer repository | Replace with generic downstream consumer wording |
| `docs/sprints/sprint-2.md` | Mentions internal consumer repository and private bot handle | Replace with generic downstream consumer and `@ExampleAlertBot` |
| `scripts/test_ingress_key_feasibility.py` | Uses internal node hostname prefix in sample alert summary | Replace with `example-node-1` |
| `scripts/fixture.py` | Previously contained a local RFC 1918 WSL Docker IP (`172.18.90.210:2375`) | Already removed in PR #83; historical commits will be scrubbed during git history rewrite |

### 2.2 Git Commit History Requiring Sanitization (`GA-IMPL`)

Historical commits on `main` (`3a8a549`, `1ab8257`, `92d4544`, `f371949`, `6f5b052`, `9642a58`) contain diffs touching the files listed in Section 2.1. Because GitHub exposes full commit history once a repository is made public, `GA-IMPL` (#81) will rewrite the `main` branch history after landing the working-tree sanitization so that `git log --all -S` returns **0 matches** for any private organization, repository, bot handle, chat ID, or local RFC 1918 IP.

### 2.3 GitHub Pull Request & Issue Metadata (`GA-DISC` & `GA-IMPL`)

- **Pull Requests #1 and #23**: Contain internal consumer repository references in their PR descriptions. Both will be sanitized via `gh pr edit` in `GA-IMPL` (#81).
- **Open Backlog Issues #10 and #11**:
  - Issue [#10](https://github.com/healdropper/terraform-provider-goalert/issues/10) (*Evaluate custom email delivery relay for GoAlert SMTP notifications*) and Issue [#11](https://github.com/healdropper/terraform-provider-goalert/issues/11) (*Add Telegram Bot integration for GoAlert via webhook bridge*) describe downstream consumer runtime infrastructure rather than Terraform provider resources.
  - Both issues are closed in `GA-DISC` (#79) with neutral resolution notes so the public issue tracker only contains provider-relevant items.

---

## 3. Runner Isolation & Workflow Security Audit

| Control | Audited State | Verification Method |
| :--- | :--- | :--- |
| **Repository Self-Hosted Runners** | `{"total_count": 0, "runners": []}` (No MiniPC or other self-hosted runner attached) | `gh api repos/healdropper/terraform-provider-goalert/actions/runners` |
| **Workflow `runs-on` Targets** | `.github/workflows/integration.yml`: `ubuntu-latest`, `windows-latest`<br>`.github/workflows/release.yml`: `ubuntu-latest` | Static inspection of `.github/workflows/*.yml` |
| **Workflow Action Pinning & Permissions** | All third-party actions pinned to full 40-char commit SHAs; top-level `permissions: contents: read`; `persist-credentials: false` | Static inspection of `.github/workflows/*.yml` |
| **Maintainer vs. External Fork PR Execution** | `.github/workflows/integration.yml` triggers on `pull_request: branches: [main]` so maintainer PRs (`@healdropper`) run automatically without self-approval blockage; upon public transition in `GA-VERIFY` (#82), `approval_policy: all_external_contributors` blocks all external fork PR workflows until explicitly approved | `PUT /repos/healdropper/terraform-provider-goalert/actions/permissions/fork-pr-contributor-approval` |

---

## 4. Terraform Registry Documentation & Manifest Audit

HashiCorp Terraform Registry requires `terraform-registry-manifest.json` at the repository root and Markdown documentation under `docs/index.md`, `docs/resources/<name>.md`, and `docs/data-sources/<name>.md` (without the `goalert_` prefix).

### 4.1 Root Manifest & Provider Index
- `terraform-registry-manifest.json`: Present; declares `version: 1` and `metadata.protocol_versions: ["6.0"]` matching `terraform-plugin-framework` Protocol v6 in `.goreleaser.yml`.
- `docs/index.md`: Present; documents provider configuration (`url`, `api_key`, `GOALERT_URL`, `GOALERT_API_KEY`) and least-privilege AST-pinned API key model. Will be updated in `GA-SPEC` (#80) to reference `version = "~> 1.0"`.

### 4.2 Managed Resources (`docs/resources/` — 14/14 Present)
1. `docs/resources/escalation_policy.md` (`goalert_escalation_policy`)
2. `docs/resources/heartbeat_monitor.md` (`goalert_heartbeat_monitor`)
3. `docs/resources/integration_key.md` (`goalert_integration_key`)
4. `docs/resources/label.md` (`goalert_label`)
5. `docs/resources/rotation.md` (`goalert_rotation`)
6. `docs/resources/schedule.md` (`goalert_schedule`)
7. `docs/resources/schedule_rule.md` (`goalert_schedule_rule`)
8. `docs/resources/service.md` (`goalert_service`)
9. `docs/resources/service_label.md` (`goalert_service_label`)
10. `docs/resources/system_limit.md` (`goalert_system_limit`)
11. `docs/resources/user.md` (`goalert_user`)
12. `docs/resources/user_contact_method.md` (`goalert_user_contact_method`)
13. `docs/resources/user_notification_rule.md` (`goalert_user_notification_rule`)
14. `docs/resources/user_override.md` (`goalert_user_override`)

### 4.3 Data Sources (`docs/data-sources/` — 9/9 Present)
1. `docs/data-sources/escalation_policy.md` (`goalert_escalation_policy`)
2. `docs/data-sources/heartbeat_monitor.md` (`goalert_heartbeat_monitor`)
3. `docs/data-sources/integration_key.md` (`goalert_integration_key`)
4. `docs/data-sources/rotation.md` (`goalert_rotation`)
5. `docs/data-sources/schedule.md` (`goalert_schedule`)
6. `docs/data-sources/service.md` (`goalert_service`)
7. `docs/data-sources/slack_channel.md` (`goalert_slack_channel`)
8. `docs/data-sources/slack_user_group.md` (`goalert_slack_user_group`)
9. `docs/data-sources/user.md` (`goalert_user`)

To prevent future drift between registered provider types and Registry documentation, `GA-IMPL` (#81) will add an automated check in `scripts/check_format.py` asserting that every resource and data source registered in `internal/provider/provider.go` has a corresponding Markdown file with valid YAML frontmatter in `docs/resources/` and `docs/data-sources/`.

---

## 5. Release Pipeline Audit for `v1.0.0`

- **Current `.github/workflows/release.yml` Limitation**:
  - Line 4 filters push triggers to `tags: ["v0.0.*"]`.
  - Line 19 validates `[[ "$GITHUB_REF_NAME" =~ ^v0\.0\.[0-9]+$ ]]`.
- **Required Change in `GA-IMPL` (#81)**:
  - Update push tag trigger to `tags: ["v*"]`.
  - Update tag validation regex to `[[ "$GITHUB_REF_NAME" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]` so `v1.0.0` and all subsequent standard SemVer releases pass verification and publish signed GoReleaser artifacts.
