# Sprint 0: provider baseline at SpecDD adoption

Date: 2026-09-21.
Status: Observed audit evidence; maintainer review pending.
Owner: healdropper; audit performed by the coding agent for this task.
Audited code revision: `41531ae6857b1e7541d256495da5ae44e7b44186`.
Branch: `codex/goalert-service`.
[Audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2)
and [v0.1.0-baseline milestone](https://github.com/healdropper/terraform-provider-goalert/milestone/1)
remain open. The milestone is an audit marker, not a change to v0.0.x releases.

## Scope and evidence

Read the entrypoint, provider/resource/client code, GraphQL document, Makefile,
acceptance harness, existing research/reference docs and release/CI configuration.
This is an audit of the service-provider branch, not an audit of a consuming
installation or its entire infrastructure repository.

Fresh non-mutating verification on Windows/amd64, Go 1.25.0, Terraform 1.10.5:
`make test vet check-format` passed on 2026-09-21. Go reported cached client
tests; three Python readiness regression tests executed and passed. No test
history or uncached execution is implied.

Historical evidence re-queried on 2026-09-21:
[CI run 35524383830](https://github.com/healdropper/terraform-provider-goalert/actions/runs/35524383830)
at the audited SHA passed unit/vet/format jobs on Windows/Linux, real acceptance
with Terraform 1.10.5/1.16.3, and ephemeral GPG signing/package verification.
These are 2026-09-20 results, not newly rerun Docker or production tests.

## Capability classification

| Capability | Classification | Evidence / limit |
| --- | --- | --- |
| Service CRUD, UUID import, mutable-field drift, remote deletion and clean plans | Verified by passing automated acceptance | SVC-02 through SVC-05, historical real GoAlert CI |
| Key-document protocol, user-role rejection, revoked/unknown-operation rejection | Verified by PoC assertions executed in acceptance | SVC-01, scripts/fixture.py |
| HTTP failures, null/error distinction, missing results, redirects, cancellation | Verified by named client unit tests | TestFixedDocumentProtocol, TestReadFailureNeverMeansDeleted, TestMutationRequiresConfirmation, TestRedirectIsNotFollowed, TestCancellation, TestConfiguration |
| Bounded fixture startup | Verified by readiness regression tests | Three tests in scripts/test_fixture.py |
| Private local provider installation | Verified by isolated acceptance | SVC-06; not production installation |
| Six packages, manifest, checksums and test GPG signature | Verified by packaging smoke CI | Ephemeral test identity only |
| Real tag-triggered release with maintained signing identity | Implemented but unverified end to end | Workflow exists; no release observed |
| Policy/step/integration-key resources | Not implemented | Backlog; no stubs counted as working resources |
| Production routing, Registry publication and non-tested GoAlert versions | Unverified / outside accepted milestone | No production access performed |

## Transition gaps and disposition

- Original behavior was implemented before SpecDD adoption. No historical
  red-before-green evidence is claimed; do not manufacture it retroactively.
- PR #1 remains open and unmerged. GitHub reports no provider release.
- No repository delivery Project was found in the complete account inventory.
  Board view, membership and closure automation therefore remain unverified.
- Governance readback: rebase, merge commits and squash are all enabled; branch
  deletion is disabled. Rebase-only configuration is still an adoption gap.
- Release Please is absent. Reconcile Foundry lifecycle release management with
  the maintained GoReleaser signing flow before release-process completion.
- Issue templates and future delivery issues require technical/planning follow-up.
- Baseline review, next-sprint scope acceptance and release signer selection
  remain with the maintainer.

Disposition and proposed sequence are authoritative in [roadmap](../roadmap.md).
The [service spec](../specs/goalert-service.md) separates original accepted scope
from Observed details. No gap is closed by creating this document.
