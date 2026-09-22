# Development log

## 2026-09-20 — v0.0.1 service milestone

Goal: learn and implement a generic Terraform Plugin Framework provider,
starting with GoAlert service lifecycle rather than deployment-specific logic.

- Created the private GitHub repository and a dedicated development branch.
- Investigated GoAlert v0.34.1 source and existing GraphQL tooling.
- Proved fixed-document, multi-operation API-key CRUD before defining the
  Terraform resource. User-role keys cannot create services; admin is required.
- Implemented the embedded operation contract, bounded HTTP client, explicit
  credential configuration, service CRUD, import and conservative drift refresh.
- Real Terraform acceptance passed against disposable GoAlert and PostgreSQL:
  in-place updates, empty description, import, all-field drift repair, invalid-key
  failure preserving state, external deletion/recreation, remote destroy and
  repeated clean plans. Unit tests cover transport and error classification.
- Added MPL-2.0, examples, design/development/reference/release documentation,
  pinned CI actions and draft-only GPG-signed release automation.
- A first cancellation test exposed a hanging fake HTTP handler; bounded the
  fixture handler and reran the unit suite successfully.
- Shared-doctrine assessment: the durable lesson is specific to GoAlert's API
  contract and remains in this provider's design/tests.
- Release identity is not configured. No public visibility, Registry publication,
  production adoption, signed production release or PR merge has been performed.

- Consumer validation: the first consumer's isolated HCL passed the full acceptance
  harness using private development overrides. Its regression checks also passed.
- GoReleaser configuration and six-platform snapshot build passed. Native Windows
  Git-bundled GPG could not connect to its agent; the ephemeral-signature smoke
  test runs on Linux CI. A real release signer remains unconfigured.

## CI portability and signing follow-up

- Linux CI exposed a connection reset while the disposable container was starting.
  Readiness now retries transient connection failures within the existing deadline,
  including non-ready HTTP statuses; dedicated harness regression tests cover it.
- Corrected GoReleaser Action installation mode so later steps can find the binary.
- Linux packaging CI passed: six archives, Registry manifest, SHA256 checksums and
  a detached GPG signature verified with an ephemeral test key. This validates the
  signing process but does not establish or publish a production signing identity.

## 2026-09-21 - SpecDD transition

- Reread installed and repository instructions, verified remotes and pinned
  Foundry v1.6.0; downstream governance applies to the consumer by ownership.
- Recorded the [service contract](docs/specs/goalert-service.md) and
  [dated baseline](docs/sprints/sprint-0-baseline.md), preserving existing docs
  and historical verification without inventing prior TDD or acceptance.
- Added the [roadmap](docs/roadmap.md) and [proposed sprint](docs/sprints/sprint-1.md);
  their content is authoritative there rather than duplicated in this log.
- Opened [audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2)
  and its retrospective milestone; both await maintainer review.
- Fresh `make test vet check-format` passed (Go client results were cached).
  Historical real acceptance/signing CI was re-queried, not described as a fresh run.
- No provider behavior, branch name, repository visibility or production state changed.
  Process gaps and verification limits are recorded in the baseline and roadmap.
- Doctrine assessment: no upstream rule change needed; this is repository adoption.
- Documentation verification: 62 local links resolved and `git diff --check` passed.

## 2026-09-21 - Alert delivery product discovery

- Goal: analyze source-to-chat value and propose milestones before owner selection.
- Read pinned GoAlert v0.34.1 source and official Grafana/Telegram documentation.
  [Source evidence and limitations](docs/research/alert-delivery-feasibility.md)
  are authoritative there; no new API or message-delivery PoC was run.
- Added the [proposed contract](docs/specs/alert-delivery.md) and
  [milestone options](docs/milestones/alert-delivery.md). Updated the
  [roadmap](docs/roadmap.md) and linked the still-proposed
  [Sprint 1](docs/sprints/sprint-1.md), without accepting new scope.
- Recorded PR #1's actual merge in the roadmap; preserved the dated baseline.
  No provider release, infrastructure change or Telegram message was performed.
- Doctrine assessment: protocol constraints belong in this provider's research;
  consumer-specific deployment and bot ownership remain in the consumer.
- Verification: `make check-format` and `git diff --check` passed; 52 local
  documentation links, balanced code fences, proposal labels and generic-scope
  checks passed. Documentation-only work adds no behavioral test claims.

## 2026-09-21 - Select small-increment planning and create delivery tracking

- Recorded the owner's first Grafana firing scenario, group preference and
  interest in later independent availability detection in the
  [capability proposal](docs/specs/alert-delivery.md). No group or bot was created.
- Created [v0.0.2 planning](docs/milestones/v0.0.2.md), linked issues #4-#8 and
  updated [Sprint 1](docs/sprints/sprint-1.md). Detailed behavior and normal
  implementation remain pending discovery and owner acceptance.
- Created a [private Project](https://github.com/users/healdropper/projects/1),
  verified private visibility before attaching repository data, configured Board
  layout with Status columns and verified all ten intended items.
- Tracked ingress, independent availability and Telegram delivery as later
  Backlog issues #9-#11 without assigning them to v0.0.2.
- Added bug/enhancement intake templates; full delivery readiness remains in
  issue #4. Closure workflows report enabled, but no issue was artificially
  closed or marked Done to manufacture completion evidence.
- The initial ambiguous Project creation command was rejected by automatic
  review. A safe empty-container creation followed by explicit private readback
  succeeded before adding any repository link or planning content.
- Doctrine assessment: this applies the pinned lifecycle; no upstream policy
  change is needed. No PoC, feature implementation, merge, release or deployment.
- Verification: `make check-format` and `git diff --check` passed; 81 local
  links, code fences and both intake templates passed structural checks. GitHub
  readback confirmed five open milestone issues, three later backlog issues,
  Status-grouped boards and private visibility. No issues were closed.

## 2026-09-21 - Expose workflow_dispatch in release automation

- Goal: adopt the Foundry v1.8.0 release governance contingency rule in
  the release workflow, tracked under milestone v0.0.2 and Project 1.
- Updated `.github/workflows/release.yml` to include `workflow_dispatch:`
  alongside the existing `push: tags: ["v0.0.*"]` trigger.
- Preserved the existing ref name check and ancestor validation step so manual
  runs still require an exact patch tag ancestor of `origin/main`.
- Delivery tracking: linked to [issue #12](https://github.com/healdropper/terraform-provider-goalert/issues/12),
  assigned to milestone [v0.0.2](https://github.com/healdropper/terraform-provider-goalert/milestone/2),
  and tracked on [Project 1](https://github.com/users/healdropper/projects/1).
- Doctrine assessment: follows the release automation contingency rule from
  Foundry v1.8.0 / organization governance; no downstream repository policy change needed.
- Verification: `git diff --check`, `make check-format`, `go vet ./...` and `go test ./...` passed.
