# Contributing to `terraform-provider-goalert`

Thank you for your interest in contributing to `terraform-provider-goalert`!

## Development Prerequisites

- **Go**: `1.25+` (using `terraform-plugin-framework` Protocol v6)
- **Terraform CLI**: `1.10+`
- **Python**: `3.10+` (used by formatting, Registry doc linting, and disposable acceptance scripts)
- **Docker Engine & Docker Compose v2**: Required for running real end-to-end acceptance tests against an ephemeral `goalert/goalert:v0.35.0` + `postgres:17-alpine` container stack.

## Contribution Workflow

1. **Fork and Branch**:
   - Fork `healdropper/terraform-provider-goalert` on GitHub and create a focused topic branch (`feat/...`, `fix/...`, `docs/...`) from `main`.
2. **Canonical GraphQL Operations**:
   - The provider authenticates using GoAlert admin GraphQL API keys bound to the canonical operations document in [`internal/client/operations.graphql`](internal/client/operations.graphql).
   - Any new or modified GraphQL query/mutation must be added to `internal/client/operations.graphql` and covered by unit tests in `internal/client/` and `internal/provider/`.
3. **Registry Documentation**:
   - Every resource (`goalert_<name>`) and data source (`goalert_<name>`) must have a corresponding Terraform Registry documentation file in `docs/resources/<name>.md` or `docs/data-sources/<name>.md` with valid YAML frontmatter (`page_title`, `description`).
4. **Local Verification**:
   Run the full verification suite before opening a pull request:
   ```console
   make check-format
   make test
   make vet
   make build
   make acceptance
   ```
5. **Pull Requests & CI**:
   - Open a pull request against `main` using a [Conventional Commits](https://www.conventionalcommits.org/) title.
   - For external fork pull requests, GitHub Actions workflows require maintainer approval (`Approve and run`) before executing.
   - All 5 required status checks (`unit` on Linux/Windows, `acceptance` on Terraform 1.10/1.16, and `packaging`) must pass before merge.
   - Pull requests are integrated via rebase merge to preserve a clean, linear commit history on `main`.
