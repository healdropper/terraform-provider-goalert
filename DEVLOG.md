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
