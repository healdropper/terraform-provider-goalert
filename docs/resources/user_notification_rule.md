# goalert_user_notification_rule (Resource)

Manages an individual notification rule for an operator in GoAlert.

## Example Usage

```terraform
# Immediate alert via SMS
resource "goalert_user_notification_rule" "immediate_sms" {
  user_id           = goalert_user.lead.id
  contact_method_id = goalert_user_contact_method.oncall_phone.id
  delay_minutes     = 0
}

# Follow-up alert via Email after 10 minutes
resource "goalert_user_notification_rule" "backup_email" {
  user_id           = goalert_user.lead.id
  contact_method_id = goalert_user_contact_method.work_email.id
  delay_minutes     = 10
}
```

## Schema

### Required

- `contact_method_id` (String) ID of the contact method to alert.
- `user_id` (String) ID of the user owning this rule.

### Optional

- `delay_minutes` (Number) Minutes to wait after alert trigger before notifying. Defaults to `0`.

### Read-Only

- `id` (String) The unique identifier of the notification rule.

## Import

Notification rules can be imported using `<user_id>/<rule_id>`:

```shell
terraform import goalert_user_notification_rule.immediate_sms 8541866e-eba1-4d6a-b905-3f2fe2da2c02/0fac4581-8ca1-48de-8daf-85f7ae9e2284
```
