# Alert delivery feasibility: source review

Status: Observed source evidence and Proposed experiments.
Owner: healdropper. Date: 2026-09-21.
GoAlert source baseline: v0.34.1,
`0918387e38650aaddd6a923d445ee992f64d6ab6`.
Related: [proposed contract](../specs/alert-delivery.md),
[experiments and milestones](../milestones/alert-delivery.md).
No new API PoC, alert injection, Telegram message or runtime test was performed.

## What the inspected sources establish

| Finding | Evidence | Implication / confidence boundary |
| --- | --- | --- |
| GoAlert has a Grafana-specific incoming endpoint and integration-key type | [HTTP routes](https://github.com/target/goalert/blob/v0.34.1/app/inithttp.go), [key types](https://github.com/target/goalert/blob/v0.34.1/integrationkey/type.go) | Use the documented Grafana payload path as the first candidate; prove compatibility with the consumer's pinned Grafana image |
| Payload version 1 maps firing/resolved to triggered/closed, using fingerprint for deduplication | [Grafana handler](https://github.com/target/goalert/blob/v0.34.1/grafana/grafana.go) | Test firing, duplicate and recovery, including multiple alerts |
| That handler truncates a batch after ten alerts | Same handler | A one-rule PoC is safe from this limit; larger adoption needs an explicit grouping/splitting strategy. Merely setting a sender's maximum can discard excess alerts |
| Policy steps expose actions; targets are deprecated | [schema](https://github.com/target/goalert/blob/v0.34.1/graphql2/schema.graphql), [step extension](https://github.com/target/goalert/blob/v0.34.1/graphql2/graph/escalationpolicy.graphqls) | Investigate typed destination actions before inventing channel resources |
| Step destination inputs have a type and argument map | [destinations](https://github.com/target/goalert/blob/v0.34.1/graphql2/graph/destinations.graphqls) | Capability discovery is possible; map availability still needs an authenticated PoC |
| Create/read key operations exist; no ordinary updateIntegrationKey was found | [mutations](https://github.com/target/goalert/blob/v0.34.1/graphql2/graph/_Mutation.graphqls), [resolver](https://github.com/target/goalert/blob/v0.34.1/graphql2/graphqlapp/integrationkey.go) | Treat mutable-vs-replacement behavior as a design question; key rotation is not an innocuous rename |
| A Grafana key href embeds the key ID as the token query parameter | Same key resolver | Both ID and href are secret-bearing. Import commands, shell history, Terraform state and outputs need explicit handling |
| Built-in webhook uses its own alert/status/bundle JSON | [sender](https://github.com/target/goalert/blob/v0.34.1/notification/webhook/sender.go), [webhook documentation](https://github.com/target/goalert/blob/v0.34.1/web/src/app/documentation/sections/Webhooks.md) | This is not Telegram sendMessage JSON; an ordinary notification requires translation |
| Webhook destination advertises enablement, required URL and status capabilities | [destination implementation](https://github.com/target/goalert/blob/v0.34.1/notification/webhook/nfydest.go) | Verify enabled/allowlisted URL and actual status delivery from a policy action |

Current [Grafana webhook documentation](https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/)
describes the version-1 payload and resolved-message controls. It also warns
that webhook URLs are readable, unlike protected authorization fields. A GoAlert
href containing a token therefore requires an exposure review. Test supported
header authentication before suggesting a safer equivalent; do not assume a
Grafana authentication option is understood by GoAlert. Latest docs do not prove
the exact consumer version behaves identically.

## Telegram and delivery options

[Telegram sendMessage](https://core.telegram.org/bots/api#sendmessage) requires
a destination and text, authenticated using a bot token. A personal chat, group
and channel have different setup/permission requirements. Confirm the target and
bot membership/posting rights before testing. The API response supplies message
metadata; it does not prove the human read the notification.
See [bot introduction](https://core.telegram.org/bots) and
[FAQ](https://core.telegram.org/bots/faq).

| Option | Product value | Cost / limit | Proposal |
| --- | --- | --- | --- |
| GoAlert webhook to a small external Telegram adapter | Preserves GoAlert incident and escalation lifecycle; supports any bot/chat configured by the operator | Another runtime component, delivery/state/retry ownership | Recommended hypothesis for the first end-to-end PoC |
| GoAlert notification plugin | Could integrate transport-specific semantics more deeply | Plugin/deployment/API investigation and lifecycle ownership | Research alternative; not proven here |
| Ordinary webhook directly to Telegram URL | Few components | GoAlert alert JSON does not supply Telegram's required message contract | Do not treat it as a working integration |
| Grafana directly to Telegram | Fast independent baseline | Bypasses the requested GoAlert chain | Optional comparison only, never counted as provider success |

The webhook destination also advertises dynamic body parameters for a separate
dynamic-action path. That does not establish templated Telegram payloads for
ordinary escalation notifications. Investigate it only if choosing that
alternative; do not conflate incoming universal-key rules with policy delivery.

## Critical reliability caveat in v0.34.1

The inspected ordinary webhook sender calls `http.DefaultClient.Do(req)` and
does not inspect HTTP response status before returning StateSent. Although it
creates a three-second context, the request is constructed without attaching
that context. Its configured client field is not used on this path.

This is source evidence, not a newly executed regression or a claim about every
GoAlert version. It means the PoC must include 401/500/timeout probes and receiver
evidence; an upstream “sent” flag alone cannot establish success. A provider
cannot repair runtime delivery behavior through Terraform state management.

An adapter could acknowledge only after a durable enqueue, respond promptly,
and handle Telegram 429/5xx/timeouts with bounded retries and explicit outcomes.
That helps after acceptance but does not guarantee the preceding GoAlert-to-
adapter hop. Before production claims, decide whether verified upstream behavior,
a reviewed fix/upgrade or an explicitly accepted limitation is necessary.
Do not claim exactly-once delivery: a timeout after Telegram accepts a message
can leave the adapter uncertain. Correlation and deduplication policies must
recognize retries without suppressing legitimate escalation reminders.

## Credential boundaries

- Management GraphQL key: fixed canonical document, named operations, admin
  capability already proven for the service. Expanded policy/key operations
  need their own PoC and migration contract.
- Incoming Grafana integration key: service-scoped alert ingress; distinct from
  the management key. Do not assume universal-key token rotation applies to the
  legacy Grafana key type.
- Adapter authentication: define a supported, restricted ingress contract;
  native webhook custom headers/signatures are not established by this review.
- Telegram bot token: belongs to the adapter's secret manager, never to the
  provider implementation. Token-bearing URLs must not enter diagnostics.

Terraform `sensitive` suppresses ordinary display; it is not state encryption
or a promise that identifiers, import commands or provider errors are hidden.
Proposed acceptance must include synthetic-secret leak checks, replacement
ordering, explicit revocation and sanitized troubleshooting.

## Runtime evidence still needed

Policy action creation through the canonical key; alert/closed/bundle event
delivery; maximum-field and ordering behavior; invalid-destination validation;
referenced-policy deletion; key read-after-delete and import; the actual Grafana
payload; adapter error cases; and a separately authorized real chat receipt.
The existing service acceptance suite proves none of these new delivery paths.
