---
page_title: "goalert_slack_user_group Data Source"
description: "Fetches a Slack user group integrated with GoAlert by its user group ID or group name."
---

# goalert_slack_user_group (Data Source)

Fetches a Slack user group integrated with GoAlert by its user group ID or group name.

## Example Usage

```terraform
# Look up Slack user group by name
data "goalert_slack_user_group" "devops" {
  name = "devops-oncall"
}

# Look up Slack user group by ID
data "goalert_slack_user_group" "by_id" {
  id = "S0123456789"
}
```

## Schema

### Optional

- `id` (String) Slack user group ID. Exactly one of `id` or `name` must be set.
- `name` (String) Slack user group name. Exactly one of `id` or `name` must be set.

### Read-Only

- `handle` (String) Slack user group handle (e.g. mention handle).
