# Specification: goalert_slack_channel Data Source Contract

Status: Accepted contract for Milestone `Collaboration Channels and System Limits` (Issue #62 / Gate COL-SPEC).
Owner: healdropper.
Evidence: [Feasibility Research](../research/collaboration-and-limits-feasibility.md).

## 1. Scope and Criteria

The `data.goalert_slack_channel` data source resolves an existing Slack channel integrated with GoAlert by ID or name.

## 2. Schema Specification

### Query Arguments
- Exactly one of `id` or `name` must be specified:
  - `id` (String, Optional, Computed): Slack Channel ID (e.g. `"C01234567"`).
  - `name` (String, Optional, Computed): Slack Channel name (e.g. `"general"`).

### Exported Attributes
- `id` (String): Slack Channel ID.
- `name` (String): Slack Channel name.
- `team_id` (String): Slack Workspace Team ID.

## 3. Resolution Logic

1. When `id` is specified:
   - Execute query `slackChannel(id: $id)`.
   - If not found or if Slack is not integrated/authorized, return diagnostic error.
2. When `name` is specified:
   - Execute query `slackChannels(input: { search: $name, first: 50 })`.
   - Filter results for exact string match on `name`.
   - Return error if 0 or >1 matches are found.
