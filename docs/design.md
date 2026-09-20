# Design and feasibility

## Research (2026-09-20)

GitHub repository search for `terraform-provider-goalert` found only this newly
created repository. Searches of the Terraform Registry and GoAlert upstream
did not identify a mature GoAlert-specific provider. This is a dated search
result, not proof that no private or future implementation exists.

The generic [sullivtr/graphql provider](https://github.com/sullivtr/terraform-provider-graphql)
is an alternative for raw GraphQL lifecycle operations. This project instead
provides a typed GoAlert schema, import semantics, drift handling and a tested
credential contract using the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).
It does not use Terraform Plugin SDK v2.

Primary GoAlert sources reviewed at v0.34.1, commit
`0918387e38650aaddd6a923d445ee992f64d6ab6`:

- [API key middleware](https://github.com/target/goalert/blob/v0.34.1/apikey/middleware.go)
- [Key architecture ADR](https://github.com/target/goalert/blob/v0.34.1/docs/adrs/000005-api-key-architecture.md)
- [GraphQL schema](https://github.com/target/goalert/blob/v0.34.1/graphql2/schema.graphql)
- [Service resolvers](https://github.com/target/goalert/blob/v0.34.1/graphql2/graphqlapp/service.go)
- [Upstream API key smoke test](https://github.com/target/goalert/blob/v0.34.1/test/smoke/gqlapikeys_test.go)
- [API documentation issue](https://github.com/target/goalert/issues/4467)

## Fixed-document credential contract

A GoAlert API key is bound to a stored GraphQL document and expires.
Requests with different query text are rejected, including formatting differences.
Our embedded canonical document has four named operations with scalar variables.
The client sends `query: ""`, `operationName` and `variables`; GoAlert injects
the document registered for the key. There is no runtime query generation,
username/password login, browser-cookie fallback or direct database access in
the provider.

Create the key from the document shipped with the provider version you install.
Keep operation names, variables, selected fields and semantics intact. Although
the client omits query text, it still depends on this contract; using an unrelated
or modified stored document is unsupported. Adding resources/operations requires
an expanded document and a newly issued key before upgrading. A superset can
support older binaries only if the existing operations remain unchanged.
Rotate keys before expiration and revoke old ones after validation.

The real PoC established:

| Probe | v0.34.1 result |
| --- | --- |
| One document with query and mutations | Accepted |
| Empty query plus named operation and variables | Accepted |
| Different query text | Rejected with `wrong query for API key` |
| User-role key creating a service | Unauthorized |
| Admin-role key creating/updating a service | Accepted |
| Empty description on update | Clears description |
| Read after deletion | `data.service = null`, no errors |
| Unknown operation / revoked key | Rejected |
| Repeated deletion | `deleteAll = true` |

Admin is necessary for these API-key mutations; the key does not represent a
normal user identity. Its document limits allowed operations but **not ownership
of particular service UUIDs**. The delete operation hardcodes the `service`
target type, so it cannot delete arbitrary object types. It can delete any service
whose ID is supplied. Use a dedicated key per automation trust boundary.

## Framework lifecycle

- Provider `Configure` resolves explicit settings or environment variables and
  builds an HTTP client. It makes no network calls during configuration.
- Resource `Schema` declares required inputs, a computed UUID and an empty-string
  default for description. `UseStateForUnknown` preserves the UUID during updates.
- `Create` sends variables and saves the API's result into Terraform state.
- `Read` refreshes every managed field. Only an explicit null service with no
  GraphQL errors removes the resource from state. HTTP 404, authorization errors,
  partial GraphQL responses and malformed results are diagnostic failures.
- `Update` changes fields in place, then reads the canonical server state.
- `Delete` uses the idempotent service deletion mutation.
- `ImportState` validates and copies the UUID; the following `Read` fills fields.

The API requires an escalation policy UUID even though the input schema marks it
optional. Making it required in Terraform avoids implicit policy creation and
ambiguous ownership. This resource never manages policy steps or notification
channels. API-side validation remains authoritative for name/text restrictions.

The HTTP client uses context cancellation, a 30-second deadline, bounded
responses and verified TLS. It refuses redirects and does not log tokens, request
bodies or raw server error bodies. Mutations are not retried: a connection loss
after a server commit requires inspecting and importing the object before retry.
HTTP beyond loopback requires an explicit provider setting.

## Acceptance and limitations

`make acceptance` uses the real Terraform CLI and Framework protocol, rather
than a simulated Terraform state machine. It checks create, update without
replacement, empty description, import, drift of all mutable fields, invalid-key
state preservation, external deletion/recreation, destroy and repeated no-change
plans (detailed exit code 0 and JSON actions all `no-op`).

The test fixture bootstraps an admin account through GoAlert's own CLI and creates
keys through an authenticated admin session. That bootstrap exists only in the
disposable harness. The provider never bootstraps users or keys.

Compatibility is established for GoAlert v0.34.1, not every future API release.
The provider accepts arbitrary installation URLs and has no cluster-specific
assumptions. Future compatibility requires rerunning the suite against those
versions. Unit tests cover malformed responses, authorization failures, missing
fields, redirects, cancellation and absence of automatic mutation retries.

Durable lesson assessed: API-specific credential and null/error semantics belong
in this provider's design and tests. No shared organization doctrine change is
needed.
