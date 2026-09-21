# Vision

Status: Accepted original task scope, recorded at SpecDD transition on 2026-09-21.
Owner: healdropper.
Evidence: the maintainer's original provider request and explicit subsequent
decision to use Plugin Framework and a canonical GraphQL document.
This records existing authorization; it does not assert earlier formal specs.
Related: [lifecycle](spec-driven-lifecycle.md), [roadmap](../roadmap.md),
[audit issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2).

## Problem and users

Infrastructure engineers need a typed Terraform provider to manage GoAlert
configuration reproducibly while learning how Terraform provider lifecycles work.

## Three inviolable pillars

1. Installation-independent design: no consuming project's names, endpoints or
   topology in the provider's resource behavior.
2. Verified state management: prove API feasibility, then verify CRUD, import,
   drift and convergence with real disposable GoAlert and Terraform.
3. Controlled delivery: private development, small v0.0.x increments and separate
   decisions for public visibility, Registry publication and production adoption.

## Scope, exclusions and success

The accepted first milestone is specified in [goalert-service.md](goalert-service.md).
It includes documentation, examples, a license, tests, private development
integration and signed-release automation preparation.

The initial milestone does not manage a complete on-call configuration, deploy
GoAlert, select real recipients, expose the repository publicly or install a
provider in production. Later resources require concrete API investigation and
accepted specification/sprint deltas; general expansion intent does not select
their schema or operational settings.

Success is evidence against the service acceptance identifiers, a reviewable
change, and preserved publication boundaries. A passed prototype test is not a
production release. Accountable acceptance and unresolved decisions remain with
the maintainer.
