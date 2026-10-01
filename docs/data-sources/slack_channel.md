---
page_title: "goalert_slack_channel Data Source"
description: "Fetches a Slack channel integrated with GoAlert by its channel ID or channel name."
---

# goalert_slack_channel (Data Source)

Fetches a Slack channel integrated with GoAlert by its channel ID or channel name.

## Example Usage

```terraform
# Look up Slack channel by name
data "goalert_slack_channel" "alerts" {
  name = "alerts-critical"
}

# Look up Slack channel by ID
data "goalert_slack_channel" "by_id" {
  id = "C0123456789"
}
```

## Schema

### Optional

- `id` (String) Slack channel ID. Exactly one of `id` or `name` must be set.
- `name` (String) Slack channel name. Exactly one of `id` or `name` must be set.

### Read-Only

- `team_id` (String) Slack workspace team ID associated with the channel.
