# Collaboration Channels and System Limits

Owner: healdropper. Recorded: 2026-09-28.
Status: Planned and tracked under [GitHub Milestone 8](https://github.com/healdropper/terraform-provider-goalert/milestone/8).

## Recorded owner decisions

- Implement system administrative limits configuration via `goalert_system_limit` (`id`, `value`, `description`).
- Provide data sources for Slack/chat integration channels (`data.goalert_slack_channel`) and Slack user groups (`data.goalert_slack_user_group`).
- Characterize and verify system limit boundaries, default values, and in-place updates.

## Planning envelope

Proposed outcomes:
1. Manage GoAlert instance system limits (e.g. max rules per schedule, max user overrides, max contact methods per user) declaratively.
2. Query Slack channels and user groups configured within the GoAlert instance.
3. Verify complete lifecycle via disposable GoAlert v0.35.0 acceptance tests.

| ID | Planning/acceptance gate | Evidence required before closing |
| --- | --- | --- |
| COL-DISC | Establish system limit and collaboration channel API feasibility | Probe suite against disposable GoAlert v0.35.0 characterizing `systemLimits`, `setSystemLimit`, `slackChannels`, and `slackUserGroups` |
| COL-SPEC | Resolve and accept resource and data source contracts | Canonical specifications in `docs/specs/` defining schemas, valid limit IDs, validations, and immutability rules; matching documentation in `docs/resources/` and `docs/data-sources/` |
| COL-IMPL | Implement accepted contracts | Provider resource `goalert_system_limit`, data sources `data.goalert_slack_channel` and `data.goalert_slack_user_group`, and typed client methods |
| COL-VERIFY | Establish provider lifecycle and acceptance confidence | Unit tests and real disposable GoAlert acceptance tests in `scripts/acceptance.py` verifying full CRUD, limit mutations, drift repair, and imports |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| COL-DISC | [#61: feat: investigate system limits and collaboration channel API feasibility](https://github.com/healdropper/terraform-provider-goalert/issues/61) | In progress |
| COL-SPEC | [#62: docs: resolve system limit and collaboration channel contracts](https://github.com/healdropper/terraform-provider-goalert/issues/62) | Pending DISC |
| COL-IMPL | [#63: feat: implement system limit resource and collaboration data sources](https://github.com/healdropper/terraform-provider-goalert/issues/63) | Pending SPEC |
| COL-VERIFY | [#64: test: verify system limit and collaboration channel lifecycles](https://github.com/healdropper/terraform-provider-goalert/issues/64) | Pending IMPL |
