# RFC: detect Grafana unavailability independently

Status: Proposed Backlog; owner expressed interest on 2026-09-21.
Owner: healdropper. Not selected for v0.0.2.
Related: [alert delivery](alert-delivery.md), [roadmap](../roadmap.md).

A running Grafana can notify a workload condition. It cannot be relied on to
evaluate and deliver an alert during its own outage. An independent evaluator
or dead-man monitor should create/recover a GoAlert incident.

Discovery must distinguish process failure, datasource failure, host/network
failure and loss of the entire monitoring stack. A probe running beside the
failing Grafana may share the same failure domain; an external probe still
cannot deliver through an unavailable GoAlert or Telegram adapter.

Candidates, not selected architecture: an independently evaluated availability
probe or a GoAlert heartbeat monitor with a separately proven lifecycle.
Heartbeat semantics, ownership and false-positive controls need a new PoC.

Proposed success: a controlled non-production Grafana failure is detected and
reported through a surviving notification path, then resolves after recovery.
Thresholds, evaluator deployment, transport and fallback for GoAlert failure
remain owner decisions. No outage or deployment is authorized by this RFC.

Tracking: [issue #10](https://github.com/healdropper/terraform-provider-goalert/issues/10), outside the v0.0.2 milestone.
