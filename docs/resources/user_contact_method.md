---
page_title: "goalert_user_contact_method Resource"
description: "Manages a notification contact method channel on a GoAlert user account."
---

# goalert_user_contact_method (Resource)

Manages a notification contact method channel on a GoAlert user account.

## Example Usage

```terraform
resource "goalert_user_contact_method" "work_email" {
  user_id = goalert_user.lead.id
  name    = "Primary Email"
  type    = "EMAIL"
  value   = "jane@example.com"
}

resource "goalert_user_contact_method" "oncall_phone" {
  user_id = goalert_user.lead.id
  name    = "On-call Cell"
  type    = "SMS"
  value   = "+15555550199"
}
```

## Schema

### Required

- `name` (String) Friendly name of the contact method.
- `type` (String) Channel type (`SMS`, `VOICE`, `EMAIL`, `WEBHOOK`, `SLACK_DM`). Cannot be modified after creation.
- `user_id` (String) ID of the user owning this contact method.
- `value` (String) Contact destination value (phone number, email address, or webhook URL).

### Optional

- `enable_status_updates` (Boolean) Whether to send alert status updates to this contact method (for destination types that support toggling status updates). Defaults to `false`.
- `private` (Boolean) Whether to mark this contact method as private (GoAlert v0.35.0+). Defaults to `false`. Note: GoAlert hides `private = true` contact methods from non-owner sessions and system API keys on subsequent reads.

### Read-Only

- `disabled` (Boolean) Whether the contact method is currently disabled.
- `id` (String) The unique identifier of the contact method.
- `status_updates` (String) Effective status update delivery mode reported by GoAlert (`ENABLED`, `DISABLED`, `ENABLED_FORCED`, `DISABLED_FORCED`).

## Import

Contact methods can be imported using `<user_id>/<contact_method_id>` or standalone `<contact_method_id>`:

```shell
terraform import goalert_user_contact_method.work_email 8541866e-eba1-4d6a-b905-3f2fe2da2c02/ca7f2aff-8ab4-40f5-bfe4-45b291fde9ad
```
