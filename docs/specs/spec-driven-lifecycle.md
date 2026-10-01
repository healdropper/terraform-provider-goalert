# SpecDD adoption and transition

Status: Accepted process adoption by the maintainer's explicit 2026-09-21 request.
Owner: healdropper.
Tracking: [audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2).

This repository adopts the organization-neutral
[Foundry lifecycle v1.6.0](https://github.com/healdropper/doctrine-foundry/blob/b6d62d3b1d85f94cac13ddc68e7851d0ad54fc8e/rules/spec-driven-lifecycle.md).
Pinned commit: `b6d62d3b1d85f94cac13ddc68e7851d0ad54fc8e`.
The installed canonical file was read on 2026-09-21; SHA256:
`68a5bb3c38f4b0892def62e657367b238c2f5c23b71fd80d849186bf0f598613`.
Do not silently follow a later installed skill or lifecycle version.

The remote owner is healdropper. External organization-specific naming,
deployment topology and Release Please policy do not apply by ownership.
Foundry lifecycle governance prerequisites and release-management expectations
still require reconciliation with this repository's existing GoReleaser flow;
[the roadmap](../roadmap.md) records that open transition work.

## Existing work

Preserve the published `codex/goalert-service` branch and
[PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1).
Record specs/evidence before closure; do not rename the branch retroactively,
manufacture a preceding spec commit, or claim historical red-first TDD.
Existing explicit task acceptance is recorded as such; observed implementation
details remain Observed until reviewed.

[Baseline](../sprints/sprint-0-baseline.md) is a dated audit.
[Roadmap](../roadmap.md) holds strategy and proposed work.
[Sprint 1](../sprints/sprint-1.md) is a proposal, not implementation authorization.
DEVLOG records chronological work and links these artifacts.

Future behavior changes require an accepted, committed specification delta,
an issue linked to accepted sprint criteria, an observed failing behavior test
before implementation, relevant Make verification, and a reviewable PR.
Future branches follow the pinned lifecycle; the existing branch is grandfathered.
Documentation-only adoption uses document/link validation without invented
failing product tests.

Visibility change, Registry publication, production adoption, merges and
production signing-key selection retain their existing explicit decision
boundaries. A baseline milestone is not a release, tag or version reset.
