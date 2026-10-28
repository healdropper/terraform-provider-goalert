# Specification: `goalert_system_config` Resource Contract

Owner: healdropper. Status: Accepted (`CFG-SPEC`, Issue [#90](https://github.com/healdropper/terraform-provider-goalert/issues/90)).

## 1. Purpose & Scope

This specification defines the Terraform resource schema, state security invariants, GraphQL client boundary, partial key management semantics, zero-diff import behavior, and error diagnostics for `goalert_system_config` (`healdropper/goalert` `v1.1.0`, Issue [#88](https://github.com/healdropper/terraform-provider-goalert/issues/88)).

## 2. GoAlert `v0.35.0` Upstream API Summary

As empirically verified in [`docs/research/system-config-feasibility.md`](../research/system-config-feasibility.md) (`CFG-DISC`, Issue [#89](https://github.com/healdropper/terraform-provider-goalert/issues/89)):

1. `Query.config(all: true)` returns all **68** registered configuration items (`ConfigValue`: `id`, `description`, `value`, `type`, `password`, `deprecated`), consisting of **59** non-secret keys (`password: false`) and **9** secret keys (`password: true`).
2. `Query.config(all: true)` returns the **plaintext** `value` for every key, including `password: true` keys (`GitHub.ClientSecret`, `OIDC.ClientSecret`, `Mailgun.APIKey`, `Slack.ClientSecret`, `Slack.AccessToken`, `Slack.SigningSecret`, `Twilio.AuthToken`, `Twilio.AlternateAuthToken`, `SMTP.Password`).
3. `Mutation.setConfig(input: [ConfigValueInput!]): Boolean` applies a partial overlay of `[ConfigValueInput!]` items (`{id, value}`) and validates inter-field dependencies and scalar/list syntax atomically.

## 3. Resource Schema (`goalert_system_config`)

| Attribute | Type | Mutability | Description |
| --- | --- | --- | --- |
| `id` | `String` | `Computed` | Singleton identifier fixed to `"system"`. |
| `settings` | `Map(String)` | `Optional` | Partial map of non-secret GoAlert configuration keys (`password: false`) to string values. Setting a key to `""` clears it remotely; removing a key from `settings` unmanages it (drops it from Terraform state without modifying GoAlert). |
| `secrets_wo` | `Map(String)` | `Optional`, `WriteOnly: true` | Partial write-only map of secret GoAlert configuration keys (`password: true`) to string values (requires Terraform `>= 1.11` when populated in HCL). Never stored in `terraform.tfstate` or plan files. |
| `secret_hashes` | `Map(String)` | `Computed` | Map of managed secret keys (`password: true`) to salted `HMAC-SHA256` hex digests computed with a per-resource 256-bit random salt stored in Terraform private state (`resp.Private`). |

### 3.1 Key Partitioning & Compile-Time Validation

1. **9 Allowed `secrets_wo` Keys (`password: true`)**:
   - `GitHub.ClientSecret`
   - `OIDC.ClientSecret`
   - `Mailgun.APIKey`
   - `Slack.ClientSecret`
   - `Slack.AccessToken`
   - `Slack.SigningSecret`
   - `Twilio.AuthToken`
   - `Twilio.AlternateAuthToken`
   - `SMTP.Password`
2. **59 Allowed `settings` Keys (`password: false`)**:
   - `General.ApplicationName`, `General.PublicURL`, `General.GoogleAnalyticsID`, `General.NotificationDisclaimer`, `General.DisableMessageBundles`, `General.ShortURL`, `General.DisableSMSLinks`, `General.DisableLabelCreation`, `General.DisableCalendarSubscriptions`
   - `Services.RequiredLabels`
   - `Maintenance.AlertCleanupDays`, `Maintenance.AlertAutoCloseDays`, `Maintenance.AutoCloseAckedAlerts`, `Maintenance.APIKeyExpireDays`, `Maintenance.ScheduleCleanupDays`
   - `Auth.RefererURLs`, `Auth.DisableBasic`
   - `GitHub.Enable`, `GitHub.NewUsers`, `GitHub.ClientID`, `GitHub.AllowedUsers`, `GitHub.AllowedOrgs`, `GitHub.EnterpriseURL`
   - `OIDC.Enable`, `OIDC.NewUsers`, `OIDC.OverrideName`, `OIDC.IssuerURL`, `OIDC.ClientID`, `OIDC.Scopes`, `OIDC.UserInfoEmailPath`, `OIDC.UserInfoEmailVerifiedPath`, `OIDC.UserInfoNamePath`
   - `Mailgun.Enable`, `Mailgun.EmailDomain`
   - `Slack.Enable`, `Slack.ClientID`, `Slack.InteractiveMessages`, `Slack.DisableBroadcastThreadReplies`
   - `Twilio.Enable`, `Twilio.VoiceName`, `Twilio.VoiceLanguage`, `Twilio.AccountSID`, `Twilio.FromNumber`, `Twilio.MessagingServiceSID`, `Twilio.RCSSenderID`, `Twilio.DisableTwoWaySMS`, `Twilio.SMSCarrierLookup`, `Twilio.SMSFromNumberOverride`
   - `SMTP.Enable`, `SMTP.From`, `SMTP.Address`, `SMTP.DisableTLS`, `SMTP.SkipVerify`, `SMTP.Username`
   - `Webhook.Enable`, `Webhook.AllowedURLs`, `Webhook.BlockPrivateAddresses`
   - `Feedback.Enable`, `Feedback.OverrideURL`
3. **Diagnostic Enforcement**:
   - Placing any `password: true` key inside `settings` produces an immediate validation error instructing the practitioner to move that key to `secrets_wo`.
   - Placing any `password: false` key inside `secrets_wo` produces an immediate validation error instructing the practitioner to move that key to `settings`.
   - Placing any unknown key ID in either map produces an immediate validation error listing the invalid key.

## 4. Secret Scrubbing & Salted `HMAC-SHA256` Drift Detection

1. **Client-Boundary Plaintext Scrubbing (`internal/client/client.go`)**:
   - `ReadSystemConfig(ctx, salt)` unmarshals the `config(all: true)` response, computes `HMAC-SHA256(salt, item.Value)` (when `salt != ""`) for each item where `item.Password == true`, stores the hex digest in `item.ValueHash`, and zeroes `item.Value = ""` before returning.
   - No caller outside `ReadSystemConfig` ever receives a plaintext `password: true` value from GoAlert.
2. **Per-Resource Salt in Private State (`system_config_salt`)**:
   - A 32-byte cryptographically random hex salt (`64` hex characters) is generated via `crypto/rand` and persisted in Terraform private state (`resp.Private.SetKey(ctx, "system_config_salt", ...)`).
3. **`ModifyPlan` Drift & Rotation Detection**:
   - During `ModifyPlan`, the provider reads `secrets_wo` from `req.Config` and `system_config_salt` from `req.Private` (generating a new salt if none exists yet).
   - If `secrets_wo` is known, the provider computes the expected `secret_hashes` map (`HMAC-SHA256(salt, v)` for each key `k, v` in `secrets_wo`) and writes it to `resp.Plan` (`secret_hashes`).
   - Because `Read` refreshes `state.SecretHashes` from GoAlert's live values using the same `system_config_salt`, both local changes to `secrets_wo` in HCL and out-of-band modifications in the GoAlert Admin UI produce a deterministic plan diff on `secret_hashes[<key>]` and trigger an in-place `Update` without requiring manual version counters.

## 5. Lifecycle Semantics (`Create`, `Read`, `Update`, `Delete`, `ImportState`)

1. **`Create` / `Update`**:
   - Reads planned `settings` from `req.Plan` and write-only `secrets_wo` from `req.Config`.
   - Builds `[]client.ConfigValueInput` for all declared `settings` and `secrets_wo` keys and invokes `SetSystemConfig(ctx, inputs)` in a single atomic `setConfig` mutation.
   - Calls `ReadSystemConfig(ctx, salt)` to record confirmed `settings` and `secret_hashes` for the managed keys in state.
   - **Key Unmanagement on `Update`**: Removing a key from `settings` or `secrets_wo` drops that key from `state.Settings` / `state.SecretHashes` without sending a clearing mutation to GoAlert. Setting a key's value to `""` while keeping the key declared in HCL sends `{id: <key>, value: ""}` to GoAlert to explicitly clear the remote setting.
2. **`Read`**:
   - Calls `ReadSystemConfig(ctx, salt)` and updates only the keys currently tracked in `state.Settings` and `state.SecretHashes`. Unmanaged GoAlert config keys are ignored.
3. **`Delete`**:
   - Removes `goalert_system_config` from Terraform state without clearing or resetting remote GoAlert configuration values (`resp.State.RemoveResource(ctx)`), matching the non-destructive governance semantics of `goalert_system_limit`.
4. **Zero-Diff `ImportState`**:
   - Accepts either `"system"` or `"system:<key1>,<key2>,..."` (for example, `system:OIDC.Enable,OIDC.IssuerURL,OIDC.ClientID,OIDC.ClientSecret`).
   - Normalizes `id = "system"`, generates and stores `system_config_salt` in `resp.Private`, and seeds `state.Settings` and `state.SecretHashes` with the listed non-secret and secret keys so the immediately following `Read` populates their live remote values and `HMAC-SHA256` hashes, yielding a zero-diff `terraform plan` when HCL matches the live configuration.

## 6. Sanitized Validation Diagnostics

When `setConfig` fails validation in GoAlert:
- `extensions.isMultiFieldError == true`: Each entry in `extensions.fieldErrors` (`fieldName`, `message`) is formatted as `<ConfigID>: <reason>`.
- `extensions.isFieldError == true`: `extensions.fieldName` (normalized by stripping surrounding quotes and any `.Value` suffix, e.g. `"\"Maintenance.AlertCleanupDays\".Value"` -> `"Maintenance.AlertCleanupDays"`) and `message` (with any echoed `strconv.Parse...: parsing "...": ` substring stripped) are formatted as `<ConfigID>: <reason>`.
- Submitted secret values are never echoed in error strings.
