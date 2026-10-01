---
page_title: "goalert Provider"
description: "Declaratively manage a self-hosted GoAlert installation through its GraphQL API."
---

# goalert Provider

The `goalert` provider (`healdropper/goalert`) manages a self-hosted [GoAlert](https://goalert.me/) installation (targeting GoAlert `v0.35.0`) using the HashiCorp Terraform Plugin Framework (Protocol v6) and a least-privilege admin GraphQL API key bound to [the canonical operations document](https://github.com/healdropper/terraform-provider-goalert/blob/main/internal/client/operations.graphql).

## Example Usage

```hcl
terraform {
  required_version = ">= 1.10.0"

  required_providers {
    goalert = {
      source  = "healdropper/goalert"
      version = "~> 1.0"
    }
  }
}

provider "goalert" {
  # Recommended: configure via GOALERT_ENDPOINT and GOALERT_API_KEY environment variables
  # endpoint = "https://alerts.example.com/api/graphql"
}
```

## Schema

### Optional

- `endpoint` (String) Full GoAlert GraphQL URL (for example, `https://alerts.example.com/api/graphql`). Defaults to `GOALERT_ENDPOINT`. Required at runtime via configuration or environment variable.
- `api_key` (String, Sensitive) Admin-role GoAlert GraphQL API key bound to the provider's canonical `operations.graphql` document. Defaults to `GOALERT_API_KEY`. Required at runtime via configuration or environment variable.
- `allow_insecure_http` (Boolean) Allow plain HTTP for non-loopback endpoints. Defaults to `false` (loopback HTTP `127.0.0.1` / `localhost` is always permitted for local port-forwards and disposable tests).

## Authentication & Canonical GraphQL Document

GoAlert system GraphQL API keys are bound at creation time to an explicit GraphQL document and only allow executing the named operations present in that document. Create an **Admin**-role System API key in GoAlert (*Admin -> API Keys -> Create API Key*) using the exact contents of [`internal/client/operations.graphql`](https://github.com/healdropper/terraform-provider-goalert/blob/main/internal/client/operations.graphql), and supply the resulting token via `GOALERT_API_KEY`.

When upgrading the provider to a minor/major version that introduces new resources or fields, reissue the GoAlert API key with the updated `operations.graphql` document.

## Supported Resources & Data Sources

### Managed Resources (14)
- `goalert_service` — Alerted services bound to an escalation policy
- `goalert_escalation_policy` — Escalation policies with ordered steps (`user_ids`, `rotation_ids`, `schedule_ids`, `webhook_urls`, `multi_ack`)
- `goalert_integration_key` — Service ingress keys (`generic`, `grafana`, `site24x7`, `prometheusAlertmanager`, `email`)
- `goalert_heartbeat_monitor` — Dead-man's-switch heartbeat monitors on a service
- `goalert_service_label` — Key/value labels on a service
- `goalert_label` — Polymorphic key/value labels on a `service`, `escalationPolicy`, `schedule`, or `rotation` (GoAlert v0.35.0+)
- `goalert_user` — GoAlert user accounts and Basic Auth bootstrap credentials
- `goalert_user_contact_method` — User notification channels (`SMS`, `VOICE`, `EMAIL`, `WEBHOOK`, `SLACK_DM`, `enable_status_updates`, `private`)
- `goalert_user_notification_rule` — User alert notification delay rules
- `goalert_rotation` — On-call rotations (`hourly`, `daily`, `weekly`, `monthly`) and ordered participant lists
- `goalert_schedule` — On-call schedules with IANA time zones
- `goalert_schedule_rule` — Time-of-day and weekday coverage rules binding rotations or users to a schedule
- `goalert_user_override` — Temporary schedule coverage overrides (`add`, `remove`, or `replace` user)
- `goalert_system_limit` — GoAlert system-wide safety and rate limits

### Data Sources (9)
- `goalert_service`
- `goalert_escalation_policy`
- `goalert_integration_key`
- `goalert_heartbeat_monitor`
- `goalert_user`
- `goalert_rotation`
- `goalert_schedule`
- `goalert_slack_channel`
- `goalert_slack_user_group`
