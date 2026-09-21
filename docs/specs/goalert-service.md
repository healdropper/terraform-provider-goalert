# Service milestone contract

Owner: healdropper.
Status: Accepted original milestone requirements below; detailed existing
behavior is Observed pending baseline review.
Recorded: 2026-09-21, after implementation; not a pre-implementation spec.
Related: [PR #1](https://github.com/healdropper/terraform-provider-goalert/pull/1),
[issue #2](https://github.com/healdropper/terraform-provider-goalert/issues/2),
[baseline](../sprints/sprint-0-baseline.md).

## Accepted requirements from the original task

| ID | Requirement | Evidence location |
| --- | --- | --- |
| SVC-01 | Prove GraphQL CRUD with the fixed-document API key before defining the resource | [PoC](../../scripts/fixture.py), historical [design evidence](../design.md) |
| SVC-02 | Generic Plugin Framework resource with id, name, description, escalation_policy_id and CRUD | [resource](../../internal/provider/service_resource.go), [acceptance](../../scripts/acceptance.py) |
| SVC-03 | Import an existing service by its identifier | acceptance import phase |
| SVC-04 | Detect and repair external changes; converge to a second plan with no changes | acceptance drift and clean-plan phases |
| SVC-05 | Test the actual provider against disposable GoAlert v0.34.1 and PostgreSQL | [compose](../../compose.yaml), acceptance harness |
| SVC-06 | Document and test private dev_overrides without introducing them to production | [development guide](../development.md), acceptance harness |
| SVC-07 | Provide documentation, examples, license, tests and signed-release automation preparation | [release contract](../releases.md), packaging CI |
| SVC-08 | Keep public visibility, Registry publication and production adoption as separate decisions | [release gates](../releases.md#three-independent-decisions) |

## Observed behavior, not newly inferred product acceptance

The existing [resource reference](../resources/service.md) defines the observed
attribute schema and import syntax. [Provider reference](../index.md) describes
endpoint and credential settings. [Design](../design.md) records absence/error
handling, request limits, no automatic mutation retries and key constraints.

These documents are preserved as supporting observed contracts. In particular,
the required existing policy UUID, default empty description, admin role and
strict read-error handling are implemented/tested details. Baseline review
accepts or identifies deltas; transition does not silently bless all code.

## Verification limits

Passing cases apply only to the environments and revisions linked in the
baseline. No claim is made for complete alert delivery, production credentials,
other GoAlert versions, or a signed distributable release.
