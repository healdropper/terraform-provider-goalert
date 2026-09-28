# Collaboration Channels and System Limits Feasibility Probe

Status: Complete empirical evidence for Gate COL-DISC (Issue #61).
Executed: 2026-09-28 against disposable GoAlert v0.35.0 (`scripts/test_collaboration_feasibility.py`).

## 1. System Limits GraphQL Schema Characterization

### Queries
- **`systemLimits: [SystemLimit!]!`**:
  - `id: SystemLimitID!`
  - `description: String!`
  - `value: Int!`

### Enum `SystemLimitID`
GoAlert v0.35.0 defines 14 system limit identifiers:
1. `CalendarSubscriptionsPerUser`: Maximum number of calendar subscriptions per user.
2. `ContactMethodsPerUser`: Maximum number of contact methods per user.
3. `EPActionsPerStep`: Maximum number of actions on a single escalation policy step.
4. `EPStepsPerPolicy`: Maximum number of steps on a single escalation policy.
5. `HeartbeatMonitorsPerService`: Maximum number of heartbeat monitors per service.
6. `IntegrationKeysPerService`: Maximum number of integration keys per service.
7. `NotificationRulesPerUser`: Maximum number of notification rules per user.
8. `ParticipantsPerRotation`: Maximum number of participants per rotation.
9. `PendingSignalsPerDestPerService`: Maximum number of pending signals per destination per service.
10. `PendingSignalsPerService`: Maximum number of pending signals per service.
11. `RulesPerSchedule`: Pertains to all rules for all assignments/targets.
12. `TargetsPerSchedule`: Maximum number of targets per schedule.
13. `UnackedAlertsPerService`: Only affects newly created alerts but not re-escalated ones.
14. `UserOverridesPerSchedule`: Only limits future overrides (i.e. end in the future).

### Mutations
- **`setSystemLimits(input: [SystemLimitInput!]!): Boolean!`**:
  - `input`: List of `SystemLimitInput` objects:
    - `id: SystemLimitID!`
    - `value: Int!`
  - Mutates system limits atomically.
  - Verified empirically: Updating limit value from 15 to 20 and reading back confirmed in-place update without restarting services.

### Lifecycle & Delete Semantics
- System limits are built-in platform configurations in GoAlert; they cannot be deleted or pruned via `deleteAll`.
- In Terraform, the lifecycle for `goalert_system_limit`:
  - `Create`: Calls `setSystemLimits` with target `id` and `value`.
  - `Read`: Queries `systemLimits`, filters by matching `id`, updates `value` and computed `description`.
  - `Update`: Calls `setSystemLimits` with updated `value`.
  - `Delete`: No-op (state removal), as GoAlert limits cannot be destroyed; optionally reset to default if tracked.

## 2. Collaboration Channels (Slack) Characterization

### Queries
- **`slackChannels(input: SlackChannelSearchOptions): SlackChannelConnection!`**:
  - `input`: `{ search: String, first: Int, after: String }`
  - Returns `nodes: [SlackChannel!]!` with fields:
    - `id: ID!`
    - `name: String!`
    - `teamID: String!`
- **`slackChannel(id: ID!): SlackChannel`**:
  - Returns single `SlackChannel` by ID.
- **`slackUserGroups(input: SlackUserGroupSearchOptions): SlackUserGroupConnection!`**:
  - `input`: `{ search: String, first: Int, after: String }`
  - Returns `nodes: [SlackUserGroup!]!` with fields:
    - `id: ID!`
    - `name: String!`
    - `handle: String!`
- **`slackUserGroup(id: ID!): SlackUserGroup`**:
  - Returns single `SlackUserGroup` by ID.

### Environment & Integration Requirement
- Testing against disposable GoAlert without Slack App credentials confirmed:
  - If Slack integration is disabled or credentials are unset, GraphQL returns `Permission Denied` or internal error.
  - Slack channels and user groups represent external Slack infrastructure synchronized via the GoAlert Slack App bot token.
  - Therefore, Slack objects are strictly read-only within GoAlert (they are provisioned in Slack and discovered via GoAlert).
  - Modeled in Terraform as read-only Data Sources (`data.goalert_slack_channel`, `data.goalert_slack_user_group`).

## 3. Terraform Provider Architecture

1. **`goalert_system_limit` Resource**:
   - `id`: Required String (matches `SystemLimitID` enum, e.g. `"RulesPerSchedule"`, RequiresReplace).
   - `value`: Required Int64 (the configurable threshold, must be >= 0).
   - `description`: Computed String (descriptive explanation returned by GoAlert).
   - Import: `id` (e.g. `terraform import goalert_system_limit.rules RulesPerSchedule`).

2. **`data.goalert_slack_channel` Data Source**:
   - `id`: Optional String (Slack channel ID, e.g. `"C01234567"`).
   - `name`: Optional String (Slack channel name, e.g. `"general"`).
   - `team_id`: Computed String (Slack team ID).
   - Requires either `id` or `name`.

3. **`data.goalert_slack_user_group` Data Source**:
   - `id`: Optional String (Slack user group ID, e.g. `"S01234567"`).
   - `name`: Optional String (Slack user group name).
   - `handle`: Computed String (Slack user group handle, e.g. `"ops-team"`).
   - Requires either `id` or `name`.
