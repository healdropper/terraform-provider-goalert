# goalert_user data source specification

Authority: [Milestone User Identity and Notification Rules](../milestones/user-identity-and-notification-rules.md), Gate V005-SPEC, [Issue #39](https://github.com/healdropper/terraform-provider-goalert/issues/39).

## 1. Data source identity
- Data source type: `goalert_user`
- Managed entity: Read-only query for existing GoAlert user operators.

## 2. Schema attributes

| Attribute | Type | Requirement | Description |
| --- | --- | --- | --- |
| `id` | String (UUID) | Optional (Computed) | Direct lookup by GoAlert User UUID. Exactly one of `id`, `email`, or `name` must be specified. |
| `name` | String | Optional (Computed) | Lookup by exact user full name. |
| `email` | String | Optional (Computed) | Lookup by exact operator email address. |
| `role` | String | Computed | Role in GoAlert: `"user"` or `"admin"`. |

## 3. Query resolution rules
- If `id` is specified: performs direct O(1) query `Query.user(id: $id)`.
- If `email` or `name` is specified: queries `Query.users(input: {search: $search})` and filters for exact case-insensitive match on the requested field.
- If zero matches are found: returns a diagnostic error with clear explanation.
- If multiple matches are found: returns a diagnostic error requiring `id` disambiguation.
