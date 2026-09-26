# Sprint 1: plan and prove a small routing increment

Owner: healdropper. Completed: 2026-09-26.
Status: Completed and closed. All scoped delivery gates (V002-FOUNDATION,
V002-DISC, V002-SPEC, V002-IMPL, V002-VERIFY) verified. Production adoption
completed in `the-moonglade/cenarion-watch` via deploy workflow run #36226368007.
[Milestone v0.0.2](../milestones/v0.0.2.md) is the scope/criteria authority;
[private board](https://github.com/users/healdropper/projects/1) tracks delivery.

## Goal and user story

As a GoAlert operator, prepare a proven contract for declaring a policy with
ordered steps and a webhook destination, connected to an existing service.
The first later end-to-end scenario is a Grafana firing alert to a Telegram
group. This small routing increment does not include the entire delivery chain.

## Scope and sequence

1. Review/schedule remaining adoption prerequisites (V002-FOUNDATION) — Completed:
   audit issue #2 linked, intake templates verified, Project 1 closure workflows verified,
   and release governance contingency deployed.
2. Execute the scoped disposable policy/webhook discovery (V002-DISC) — Completed:
   disposable probes established webhook enablement, atomic inline steps, step reordering/deletion,
   referential integrity, and multi-operation API key compatibility.
3. Resolve and accept the canonical resource delta (V002-SPEC) — Completed:
   canonical specification resolved in docs/specs/goalert-escalation-policy.md.
4. Hand the issue to spec-driven-development (V002-IMPL) and verification (V002-VERIFY) — Completed:
   `goalert_escalation_policy` resource implemented via Terraform Plugin Framework, client updated with full GraphQL CRUD and error handling, unit tests and end-to-end acceptance tests passed cleanly against real GoAlert v0.34.1 container.

Issue references are maintained in the milestone's delivery table rather than
copied as a second source of acceptance criteria.

## Exclusions

No Grafana integration-key resource, real Telegram transport, bot/group creation,
production Grafana rule, outage experiment, schedules or rotations. Public
visibility, Registry publication and production adoption remain separate gates.
The [availability RFC](../specs/grafana-availability.md) is Backlog.

## Capacity and defect buffer

One routing capability at a time; no deadline or effort estimate has been
accepted. The previous 20% defect-buffer suggestion remains proposed. No active
implementation is displaced; planning refines the earlier unaccepted proposal.
Owner acceptance of the detailed scope, DoD and capacity remains an entry gate
before routine feature delivery.

## DoD and evidence

Reference V002-DISC/SPEC/IMPL/VERIFY/FOUNDATION in the milestone. Accepted specs
must precede implementation commits, followed by red-first TDD, relevant Make
checks and real disposable acceptance. Verify import, drift, auth-error state
preservation and a second plan without changes. Review merge, issue closure,
Project status, release and production adoption separately.

All sprint outcomes completed: `goalert_escalation_policy` resource implemented
and verified with full lifecycle acceptance against real GoAlert v0.34.1 container.
Deployed and verified live in `cenarion-watch` applications-configuration environment.
Milestone v0.0.2 closed.
