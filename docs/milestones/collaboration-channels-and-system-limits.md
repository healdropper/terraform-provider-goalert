# Collaboration Channels and System Limits

Owner: healdropper. Recorded: 2026-09-28.
Status: Completed and ready to close under [GitHub Milestone 8](https://github.com/healdropper/terraform-provider-goalert/milestone/8).

## Recorded owner decisions

- Implement system administrative limits configuration via `goalert_system_limit` (`id`, `value`, `description`).
- Provide data sources for Slack/chat integration channels (`data.goalert_slack_channel`) and Slack user groups (`data.goalert_slack_user_group`).
- Characterize and verify system limit boundaries, default values, and in-place updates.

## Planning envelope

Proposed outcomes:
1. Manage GoAlert instance system limits (e.g. max rules per schedule, max user overrides, max contact methods per user) declaratively.
2. Query Slack channels and user groups configured within the GoAlert instance.
3. Verify complete lifecycle via disposable GoAlert v0.35.0 acceptance tests.

| ID | Planning/acceptance gate | Evidence required before closing | Status |
| --- | --- | --- | --- |
| COL-DISC | Establish system limit and collaboration channel API feasibility | Probe suite against disposable GoAlert v0.35.0 characterizing `systemLimits`, `setSystemLimit`, `slackChannels`, and `slackUserGroups` | Closed via PR #66 |
| COL-SPEC | Resolve and accept resource and data source contracts | Canonical specifications in `docs/specs/` defining schemas, valid limit IDs, validations, and immutability rules; matching documentation in `docs/resources/` and `docs/data-sources/` | Closed via PR #67 |
| COL-IMPL | Implement accepted contracts | Provider resource `goalert_system_limit`, data sources `data.goalert_slack_channel` and `data.goalert_slack_user_group`, and typed client methods | Closed via PR #68 |
| COL-VERIFY | Establish provider lifecycle and acceptance confidence | Unit tests and real disposable GoAlert acceptance tests in `scripts/acceptance.py` verifying full CRUD, limit mutations, drift repair, and imports | Closed via PR #69 |

## Delivery issues

| Gate | Issue | Current readiness |
| --- | --- | --- |
| COL-DISC | [#61: feat: investigate system limits and collaboration channel API feasibility](https://github.com/healdropper/terraform-provider-goalert/issues/61) | Closed |
| COL-SPEC | [#62: docs: resolve system limit and collaboration channel contracts](https://github.com/healdropper/terraform-provider-goalert/issues/62) | Closed |
| COL-IMPL | [#63: feat: implement system limit resource and collaboration data sources](https://github.com/healdropper/terraform-provider-goalert/issues/63) | Closed |
| COL-VERIFY | [#64: test: verify system limit and collaboration channel lifecycles](https://github.com/healdropper/terraform-provider-goalert/issues/64) | Closed |

## Empirical Acceptance Verification Results

Executed against disposable GoAlert v0.35.0 with PostgreSQL backing:
- `PASS: initial creation and configuration of system limits verified.`
- `PASS: in-place update of system limits verified.`
- `PASS: external system limit drift detected and remediated.`
- `PASS: system limit imported successfully.`
- `PASS: old v0.0.7 API key rejected by GoAlert AST hash validation for system limits.`
- `PASS: clean teardown removes system limits from Terraform state without error.`
