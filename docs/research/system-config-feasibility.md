# GoAlert v0.35.0 System Configuration (`config` / `setConfig`) Feasibility Probe

Status: Complete empirical evidence for Gate CFG-DISC ([Issue #89](https://github.com/healdropper/terraform-provider-goalert/issues/89)).
Executed: 2026-10-28 against disposable GoAlert v0.35.0 (`scripts/test_system_config_feasibility.py`).

## 1. GraphQL Schema & Permission Characterization

### Queries and Mutations
- **`config(all: Boolean): [ConfigValue!]!`**:
  - `id: String!`
  - `description: String!`
  - `value: String!`
  - `type: ConfigType!` (`boolean`, `integer`, `string`, `stringList`)
  - `password: Boolean!`
  - `deprecated: String!`
- **`setConfig(input: [ConfigValueInput!]): Boolean!`**:
  - `input`: List of `{ id: String!, value: String! }`.
  - Applies partial updates to the stored configuration, runs cross-field validation (`cfg.Validate()`), persists a new encrypted `config` row in PostgreSQL, and reloads the in-memory store immediately.

### Role Enforcement
- Both `config(all: true)` and `setConfig` require `Admin` (or `System`) role.
- Empirical test confirmed that a `user`-role GraphQL API key is rejected on both `ProviderReadSystemConfig` and `ProviderSetSystemConfig`, while an `admin`-role GraphQL API key registered with the canonical document succeeds.

## 2. Configuration Inventory (`v0.35.0`)

`config(all: true)` returns **68** configuration keys across 11 sections (`General`, `Services`, `Maintenance`, `Auth`, `GitHub`, `OIDC`, `Mailgun`, `Slack`, `Twilio`, `SMTP`, `Webhook`, `Feedback`):

- **9 `password: true` keys** (returned in plaintext when `all: true` is queried by an Admin key):
  1. `GitHub.ClientSecret`
  2. `OIDC.ClientSecret`
  3. `Mailgun.APIKey`
  4. `Slack.ClientSecret`
  5. `Slack.AccessToken`
  6. `Slack.SigningSecret`
  7. `Twilio.AuthToken`
  8. `Twilio.AlternateAuthToken`
  9. `SMTP.Password`
- **59 `password: false` keys**, including 6 `stringList` keys serialized with `\n` (`Services.RequiredLabels`, `Auth.RefererURLs`, `GitHub.AllowedUsers`, `GitHub.AllowedOrgs`, `Twilio.SMSFromNumberOverride`, `Webhook.AllowedURLs`).

## 3. Validation Error Format & Input Scrubbing

Empirical testing against GoAlert v0.35.0 confirmed two structured GraphQL validation error shapes in `errors[].extensions`:

1. **Multi-field validation (`extensions.isMultiFieldError: true`)**:
   - Triggered by cross-field rules such as `OIDC.Enable = "true"` without `OIDC.IssuerURL`, `OIDC.ClientID`, and `OIDC.ClientSecret`:
     ```json
     {
       "message": "Multiple fields failed validation.",
       "extensions": {
         "isMultiFieldError": true,
         "fieldErrors": [
           {"fieldName": "OIDC.Enable", "message": "requires OIDC.IssuerURL to be set "},
           {"fieldName": "OIDC.IssuerURL", "message": "required to enable OIDC"},
           {"fieldName": "OIDC.ClientSecret", "message": "required to enable OIDC"}
         ]
       }
     }
     ```
2. **Single-field validation (`extensions.isFieldError: true`)**:
   - Triggered by type/format errors such as invalid integers or invalid URLs:
     ```json
     {
       "message": "integer value invalid: strconv.ParseInt: parsing \"not-a-number\": invalid syntax",
       "extensions": {
         "fieldName": "\"Maintenance.AlertCleanupDays\".Value",
         "isFieldError": true
       }
     }
     ```
   - Note that `fieldName` may wrap the config ID in quotes with a `.Value` suffix (`"\"<ConfigID>\".Value"`), and `strconv.ParseInt` / `url.Parse` include the submitted value in quotes (`parsing "<value>"`).
   - Therefore, `internal/client/client.go` must normalize `fieldName` (stripping surrounding quotes and `.Value`) and strip any quoted or echoed input values from `message` before surfacing `<ConfigID>: <reason>`.
