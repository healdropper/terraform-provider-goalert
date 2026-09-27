# goalert_user (Data Source)

Fetches details of an existing GoAlert user operator account. Exactly one of `id`, `email`, or `name` must be specified.

## Example Usage

```terraform
# Lookup by exact ID
data "goalert_user" "by_id" {
  id = "8541866e-eba1-4d6a-b905-3f2fe2da2c02"
}

# Lookup by email
data "goalert_user" "by_email" {
  email = "jane@example.com"
}

# Lookup by full name
data "goalert_user" "by_name" {
  name = "Jane Operator"
}
```

## Schema

### Optional

- `email` (String) Email address of the user to look up.
- `id` (String) ID of the user to look up.
- `name` (String) Full name of the user to look up.

### Read-Only

- `role` (String) Permission role of the user (`user` or `admin`).
