---
page_title: "goalert_system_config Resource"
description: "Manages global system configuration settings and write-only secrets in GoAlert (Admin > Config)."
---

# goalert_system_config (Resource)

Manages GoAlert global system configuration (`Admin > Config`) as a partial overlay. Only the keys declared in `settings` and `secrets_wo` are managed by Terraform; all other GoAlert configuration keys remain untouched.

## Example Usage

```terraform
variable "oidc_client_secret" {
  type      = string
  sensitive = true
}

resource "goalert_system_config" "main" {
  settings = {
    "General.PublicURL"       = "https://goalert.example.com"
    "General.ApplicationName" = "Example On-Call"
    "OIDC.Enable"             = "true"
    "OIDC.IssuerURL"          = "https://idp.example.com/application/o/goalert/"
    "OIDC.ClientID"           = "goalert-oidc-client"
    "OIDC.NewUsers"           = "true"
  }

  secrets_wo = {
    "OIDC.ClientSecret" = var.oidc_client_secret
  }
}
```

> **Note on Ephemeral Secret Sources (Terraform 1.11+):** Because `secrets_wo` is a write-only attribute (`WriteOnly: true`), you can pass values directly from Terraform `ephemeral` resources (such as Vault, AWS Secrets Manager, or Kubernetes secrets) or `ephemeral = true` variables so plaintext secrets never enter `terraform.tfstate` or plan files.

> **Security Note:** GoAlert's `config(all: true)` GraphQL query returns `password: true` keys over the wire to authenticated Admin callers. The provider immediately scrubs all plaintext secret values at the GraphQL client boundary and stores only a per-resource salted `HMAC-SHA256` digest in `secret_hashes`. Always use high-entropy secrets for `secrets_wo` entries.

## Schema

### Optional

- `settings` (Map of String) Non-secret GoAlert configuration key/value pairs (`password: false`, such as `General.PublicURL`, `OIDC.Enable`, `Services.RequiredLabels`). Multi-item `stringList` settings use newline-separated values (`\n`). Setting a key to `""` clears the setting in GoAlert; removing a key from `settings` unmanages it in Terraform without modifying GoAlert.
- `secrets_wo` (Map of String, Write-Only) Secret GoAlert configuration key/value pairs (`password: true`: `GitHub.ClientSecret`, `OIDC.ClientSecret`, `Mailgun.APIKey`, `Slack.ClientSecret`, `Slack.AccessToken`, `Slack.SigningSecret`, `Twilio.AuthToken`, `Twilio.AlternateAuthToken`, `SMTP.Password`). Requires Terraform `>= 1.11` when populated in HCL. Plaintext values are never persisted in Terraform state or plan artifacts.

### Read-Only

- `id` (String) Singleton identifier for the system configuration (`"system"`).
- `secret_hashes` (Map of String) Salted `HMAC-SHA256` hex digests for the secret keys managed in `secrets_wo`. Enables automatic drift detection if a secret is changed in HCL or modified out-of-band in the GoAlert Admin UI.

## Lifecycle & Deletion Behavior

- **Unmanaging a Single Key**: Removing a key from `settings` or `secrets_wo` removes it from Terraform state on the next apply without resetting its value in GoAlert. To clear a remote value in GoAlert, keep the key in the map and set its value to `""`.
- **Resource Destruction (`terraform destroy`)**: Destroying `goalert_system_config` is a state-only no-op and leaves the active GoAlert installation configuration intact.

## Import

Import is supported using `"system"` or `"system:<comma-separated-keys>"` for zero-diff adoption of existing settings and secrets:

```bash
terraform import goalert_system_config.main "system:General.PublicURL,General.ApplicationName,OIDC.Enable,OIDC.IssuerURL,OIDC.ClientID,OIDC.NewUsers,OIDC.ClientSecret"
```
