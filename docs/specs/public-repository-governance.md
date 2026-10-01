# Specification: Public Repository Governance and `v1.0.0` Release Contract

Owner: healdropper. Status: Accepted (`GA-SPEC`, Issue [#80](https://github.com/healdropper/terraform-provider-goalert/issues/80)).

## 1. Purpose & Scope

This specification defines the public governance, security, branch protection, CI/CD isolation, and semantic-version release contracts for `healdropper/terraform-provider-goalert` as an open-source Terraform provider published on the HashiCorp Terraform Registry (`registry.terraform.io/providers/healdropper/goalert`).

## 2. Privacy & Sanitization Invariant

1. **Zero Private Infrastructure References**: No working-tree file, commit message, commit diff, issue, or pull request description in `healdropper/terraform-provider-goalert` may reference private repositories, internal cluster names, private bot handles, internal chat IDs, or RFC 1918 host addresses.
2. **Pre-Exposure Git History Audit**: Prior to changing repository visibility to `public`, both `git grep -i` and `git log --all -S` must return zero occurrences of any private ecosystem identifier.

## 3. Self-Hosted Runner Isolation & GitHub Actions Gating

1. **Cloud-Only Execution (`total_count: 0` Self-Hosted Runners)**:
   - No self-hosted GitHub Actions runner (personal workstation, home lab, or MiniPC) may ever be registered on `healdropper/terraform-provider-goalert` or included in any runner group accessible to it.
   - All workflow jobs in `.github/workflows/integration.yml` and `.github/workflows/release.yml` must exclusively target GitHub-hosted cloud runners (`ubuntu-latest`, `windows-latest`).
2. **Fork Pull Request Approval Gating (`all_external_contributors`)**:
   - `.github/workflows/integration.yml` triggers on standard `pull_request: branches: [main]` so maintainer-authored pull requests (`@healdropper`) execute `unit`, `acceptance`, and `packaging` automatically without self-approval deadlock.
   - Repository Actions fork PR policy is enforced as `approval_policy: all_external_contributors` (`PUT /repos/healdropper/terraform-provider-goalert/actions/permissions/fork-pr-contributor-approval`), blocking all external fork pull requests from running any GitHub Actions workflow until `@healdropper` inspects the diff and explicitly clicks **Approve and run**.
3. **Least-Privilege Workflow Permissions & Pinning**:
   - All workflows declare top-level `permissions: contents: read` and `persist-credentials: false` on `actions/checkout`.
   - Third-party GitHub Actions are pinned to full 40-character commit SHAs.
   - Only the `package` job in `.github/workflows/release.yml` (bound to the `release` environment and triggered on `main`-reachable SemVer tags) receives `permissions: contents: write`.

## 4. Default-Branch (`main`) Protection & Contribution Model

1. **Fork + Pull Request Contribution Model**:
   - External contributors fork `healdropper/terraform-provider-goalert` and open pull requests targeting `main`.
2. **Rebase-Only Linear History**:
   - Repository merge settings enforce `allow_rebase_merge: true`, `allow_merge_commit: false`, `allow_squash_merge: false`, and `delete_branch_on_merge: true`.
3. **Active `main` Branch Ruleset**:
   - Immediately upon public transition, an active repository ruleset (`main-protection`) protects `refs/heads/main` with:
     - `deletion` blocked.
     - `non_fast_forward` (force-push) blocked.
     - `required_linear_history` enabled.
     - `pull_request` required (`required_approving_review_count: 1`, `require_code_owner_review: true`, `dismiss_stale_reviews_on_push: true`, `required_review_thread_resolution: true`).
     - `required_status_checks` requiring all 5 `Integration` jobs (`unit (ubuntu-latest)`, `unit (windows-latest)`, `acceptance (1.10.5)`, `acceptance (1.16.3)`, `packaging`).
     - Maintainer repository-admin bypass enabled for `@healdropper` so single-maintainer PRs (where GitHub prohibits self-approving one's own PR) can be rebase-merged once all 5 required status checks pass.

## 5. Community Governance & Vulnerability Reporting

- `.github/CODEOWNERS`: Assigns `* @healdropper` as code owner across the repository.
- `CONTRIBUTING.md`: Documents local build, unit test, formatting, and disposable Docker Compose acceptance verification (`make check-format test vet build acceptance`), plus the fork-and-PR workflow.
- `SECURITY.md`: Documents supported versions (`1.0.x`) and directs security disclosures to GitHub Private Vulnerability Reporting (`/security/advisories/new`).

## 6. Signed `v1.0.0` Release & Terraform Registry Contract

1. **Tag & Ancestry Gate**: `.github/workflows/release.yml` triggers on `v*` tags matching `^v[0-9]+\.[0-9]+\.[0-9]+$` whose commit is an ancestor of `origin/main`.
2. **GPG Detached Signature**: The `package` job in the `release` environment imports the dedicated RSA-4096 release signing key (`GPG_FINGERPRINT = 4C110F32FCEFCBE7A0662DFE8738CFE29D8E6C12`), verifies the fingerprint, runs GoReleaser v2.18.2 to build cross-platform archives + `terraform-registry-manifest.json` + `SHA256SUMS` + `SHA256SUMS.sig`, and verifies the detached GPG signature before publishing the GitHub Release.
3. **Registry Documentation Parity**: `scripts/check_format.py` verifies that every resource and data source registered in `internal/provider/provider.go` has a matching Registry Markdown document under `docs/resources/` and `docs/data-sources/`.
