# ADR 001: Framework provider and fixed-document GoAlert client

Status: Accepted explicitly requested architectural choices; Observed
implementation details identified below.
Owner: healdropper.
Related: [vision](00_VISION.md), [service contract](goalert-service.md),
[design research](../design.md), [baseline](../sprints/sprint-0-baseline.md).

## Context and alternatives

GoAlert API keys are tied to a stored GraphQL document. A generic token with
arbitrary per-call query generation cannot satisfy that constraint.
The existing [research record](../design.md) compares raw GraphQL tooling and
documents the API feasibility evidence.

Accepted choices from the original task: Terraform Plugin Framework (not SDK v2);
a canonical document containing named operations; requests select operationName
and variables while allowing the server to inject the stored query; generic
installation configuration; real disposable acceptance before production use.

## Observed implementation

The provider server in [main.go](../../main.go) exposes protocol 6.
[Provider configuration](../../internal/provider/provider.go) builds a
[bounded HTTP client](../../internal/client/client.go).
The [service resource](../../internal/provider/service_resource.go) maps
Terraform lifecycle operations to [the canonical document](../../internal/client/operations.graphql).
State and diagnostics preserve the distinction between remote absence and API
failure. The acceptance harness uses local state and process-scoped overrides.

Existing implementation diagrams and technical details are retained in
[README](../../README.md) and [design.md](../design.md), not copied here.

## Consequences and open decisions

Adding operations requires a compatible document/key migration before upgrading.
Admin-role key capability is broader than Terraform's state ownership; the fixed
document restricts operations, not which service UUIDs can be supplied.
GoAlert compatibility is verified for the tested version only.

Policy/step ownership, deletion of referenced policies, target types, release
identity and release-management reconciliation remain proposed work in the
roadmap. No ADR acceptance for those decisions is inferred from this record.
