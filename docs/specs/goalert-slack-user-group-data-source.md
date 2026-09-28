# Specification: goalert_slack_user_group Data Source Contract

Status: Accepted contract for Milestone `Collaboration Channels and System Limits` (Issue #62 / Gate COL-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/collaboration-and-limits-feasibility.md).

## 1. Scope and Criteria

The `data.goalert_slack_user_group` data source resolves an existing Slack user group integrated with GoAlert by ID or name.

## 2. Schema Specification

### Query Arguments
- Exactly one of `id` or `name` must be specified:
  - `id` (String, Optional, Computed): Slack User Group ID (e.g. `"S01234567"`).
  - `name` (String, Optional, Computed): Slack User Group name.

### Exported Attributes
- `id` (String): Slack User Group ID.
- `name` (String): Slack User Group name.
- `handle` (String): Slack User Group handle / mention string (e.g. `"oncall-team"`).

## 3. Resolution Logic

1. When `id` is specified:
   - Execute query `slackUserGroup(id: $id)`.
   - If not found or if Slack is not integrated/authorized, return diagnostic error.
2. When `name` is specified:
   - Execute query `slackUserGroups(input: { search: $name, first: 50 })`.
   - Filter results for exact string match on `name`.
   - Return error if 0 or >1 matches are found.
