---
page_title: "goalert Provider"
description: "Manage a self-hosted GoAlert installation through its GraphQL API."
---

# goalert Provider

Uses Terraform Plugin Framework and an admin-role GraphQL API key registered
with [the canonical document](../internal/client/operations.graphql).

| Argument | Type | Default |
| --- | --- | --- |
| `endpoint` | string, optional | `GOALERT_ENDPOINT`; full GraphQL URL required |
| `api_key` | sensitive string, optional | `GOALERT_API_KEY`; required at runtime |
| `allow_insecure_http` | bool, optional | false; loopback HTTP is always allowed |

Explicit configuration takes precedence over environment variables. Prefer
environment injection for keys. HTTPS verification cannot be disabled.
TLS uses the operating system trust store. Use a trusted local port forward when
developing against an installation without HTTPS.

Keys expire and are bound to operations, not to Terraform state ownership.
Reissue the key with the correct document when adding provider capabilities.
