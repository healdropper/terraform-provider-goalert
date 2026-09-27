"""Empirical feasibility probe for Issue #38 / Gate V005-DISC: Users, Contact Methods, and Notification Rules.
Validates GraphQL operations, schema types, boundaries, and permissions against disposable GoAlert v0.35.0.
"""
import json
import uuid
from fixture import disposable

INTROSPECTION_QUERY = """
query IntrospectUserSchema {
  __type(name: "Mutation") {
    fields {
      name
      description
      args {
        name
        type {
          name
          kind
          ofType { name kind }
        }
      }
      type {
        name
        kind
        ofType { name kind }
      }
    }
  }
  queryType: __type(name: "Query") {
    fields {
      name
    }
  }
  targetType: __type(name: "TargetType") {
    enumValues {
      name
      description
    }
  }
  userRole: __type(name: "UserRole") {
    enumValues {
      name
      description
    }
  }
  cmType: __type(name: "ContactMethodType") {
    enumValues {
      name
      description
    }
  }
  userInput: __type(name: "CreateUserInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  updateUserInput: __type(name: "UpdateUserInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  cmInput: __type(name: "CreateUserContactMethodInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  nrInput: __type(name: "CreateUserNotificationRuleInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  userType: __type(name: "User") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  contactMethodType: __type(name: "UserContactMethod") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  destType: __type(name: "Destination") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  updateCmInput: __type(name: "UpdateUserContactMethodInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
}
"""

def probe():
    with disposable() as f:
        print("=== 1. Introspection of User, Contact Method, and Notification Rule Schema ===", flush=True)
        schema_data = f.graphql(query=INTROSPECTION_QUERY, session=True)

        mutations = {field["name"]: field for field in schema_data["__type"]["fields"]}
        user_mutations = [name for name in mutations if any(k in name.lower() for k in ["user", "contactmethod", "notificationrule"])]
        print(f"Relevant mutations found: {user_mutations}", flush=True)

        query_fields = [f["name"] for f in (schema_data.get("queryType") or {}).get("fields", [])]
        user_queries = [q for q in query_fields if any(k in q.lower() for k in ["user", "contact", "rule", "method"])]
        print(f"Relevant queries found: {user_queries}", flush=True)

        target_types = [ev["name"] for ev in (schema_data.get("targetType") or {}).get("enumValues", [])]
        print(f"TargetType enum values: {target_types}", flush=True)

        user_roles = [ev["name"] for ev in (schema_data.get("userRole") or {}).get("enumValues", [])]
        print(f"UserRole enum values: {user_roles}", flush=True)

        cm_types = [ev["name"] for ev in (schema_data.get("cmType") or {}).get("enumValues", [])]
        print(f"ContactMethodType enum values: {cm_types}", flush=True)

        print("\nUpdateUserInput fields:", flush=True)
        print(json.dumps(schema_data.get("updateUserInput"), indent=2), flush=True)

        print("\nCreateUserContactMethodInput fields:", flush=True)
        print(json.dumps(schema_data.get("cmInput"), indent=2), flush=True)

        print("\nCreateUserNotificationRuleInput fields:", flush=True)
        print(json.dumps(schema_data.get("nrInput"), indent=2), flush=True)

        print("\nUser fields:", flush=True)
        user_fields = [fld["name"] for fld in (schema_data.get("userType") or {}).get("fields", [])]
        print(user_fields, flush=True)

        print("\nUserContactMethod fields:", flush=True)
        cm_fields = [fld["name"] for fld in (schema_data.get("contactMethodType") or {}).get("fields", [])]
        print(cm_fields, flush=True)

        print("\nUserContactMethod field details:", flush=True)
        for fld in (schema_data.get("contactMethodType") or {}).get("fields", []):
            print(f"  {fld['name']}: {json.dumps(fld['type'])}", flush=True)

        print("\nUserNotificationRule fields:", flush=True)
        nr_fields = [fld["name"] for fld in (schema_data.get("notificationRuleType") or {}).get("fields", [])]
        print(nr_fields, flush=True)

        print("\n=== 2. Testing Live Operations with Session Admin ===", flush=True)
        
        # Test createUser with username and password
        create_user_q = """
        mutation CreateUser($input: CreateUserInput!) {
          createUser(input: $input) {
            id
            name
            email
            role
          }
        }
        """
        user_res = f.graphql(query=create_user_q, variables={
            "input": {
                "name": "Alex DevOps",
                "email": "alex.devops@example.com",
                "role": "user",
                "username": "alexdevops",
                "password": "Password123!"
            }
        }, session=True)
        user = user_res["createUser"]
        user_id = user["id"]
        print(f"Created user: id={user_id}, name={user['name']}, email={user['email']}, role={user['role']}", flush=True)

        # Test updateUser
        update_user_q = """
        mutation UpdateUser($input: UpdateUserInput!) {
          updateUser(input: $input)
        }
        """
        f.graphql(query=update_user_q, variables={
            "input": {
                "id": user_id,
                "name": "Alex Senior DevOps",
                "email": "alex.senior@example.com",
                "role": "admin"
            }
        }, session=True)
        print("Updated user successfully", flush=True)

        # Read back user
        read_user_q = """
        query ReadUser($id: ID!) {
          user(id: $id) {
            id
            name
            email
            role
          }
        }
        """
        read_user = f.graphql(query=read_user_q, variables={"id": user_id}, session=True)["user"]
        assert read_user["name"] == "Alex Senior DevOps"
        assert read_user["email"] == "alex.senior@example.com"
        assert read_user["role"] == "admin"
        print("Read updated user confirmed", flush=True)

        # Test search users
        search_user_q = """
        query SearchUsers($search: String!) {
          users(input: {search: $search, first: 10}) {
            nodes {
              id
              name
              email
              role
            }
          }
        }
        """
        search_res = f.graphql(query=search_user_q, variables={"search": "Alex Senior DevOps"}, session=True)["users"]["nodes"]
        assert any(u["id"] == user_id for u in search_res)
        print("Search user by name confirmed", flush=True)

        print("\nDestination fields:", flush=True)
        dest_fields = [fld["name"] for fld in (schema_data.get("destType") or {}).get("fields", [])]
        print(dest_fields, flush=True)
        for fld in (schema_data.get("destType") or {}).get("fields", []):
            print(f"  {fld['name']}: {json.dumps(fld['type'])}", flush=True)

        print("\nUpdateUserContactMethodInput fields:", flush=True)
        print(json.dumps(schema_data.get("updateCmInput"), indent=2), flush=True)

        # Test create contact method (WEBHOOK is enabled by default in GoAlert)
        create_cm_q = """
        mutation CreateContactMethod($input: CreateUserContactMethodInput!) {
          createUserContactMethod(input: $input) {
            id
            name
            disabled
            dest {
              type
              args
            }
          }
        }
        """
        webhook_cm = f.graphql(query=create_cm_q, variables={
            "input": {
                "userID": user_id,
                "name": "Ops Webhook",
                "type": "WEBHOOK",
                "value": "https://example.com/alerts"
            }
        }, session=True)["createUserContactMethod"]
        webhook_cm_id = webhook_cm["id"]
        print(f"Created WEBHOOK contact method: id={webhook_cm_id}, disabled={webhook_cm['disabled']}, dest={webhook_cm['dest']}", flush=True)

        # Check contact method read on User
        user_with_cms_q = """
        query UserWithCMs($id: ID!) {
          user(id: $id) {
            id
            contactMethods {
              id
              name
              disabled
              dest {
                type
                args
              }
            }
          }
        }
        """
        user_cms = f.graphql(query=user_with_cms_q, variables={"id": user_id}, session=True)["user"]["contactMethods"]
        print(f"User contact methods: {user_cms}", flush=True)
        assert len(user_cms) == 1

        # Test notification rule creation
        create_nr_q = """
        mutation CreateNR($input: CreateUserNotificationRuleInput!) {
          createUserNotificationRule(input: $input) {
            id
            delayMinutes
            contactMethod {
              id
              name
              dest {
                type
                args
              }
            }
          }
        }
        """
        nr1 = f.graphql(query=create_nr_q, variables={
            "input": {
                "userID": user_id,
                "contactMethodID": webhook_cm_id,
                "delayMinutes": 0
            }
        }, session=True)["createUserNotificationRule"]
        nr1_id = nr1["id"]
        print(f"Created immediate notification rule: id={nr1_id}, delay={nr1['delayMinutes']}", flush=True)

        nr2 = f.graphql(query=create_nr_q, variables={
            "input": {
                "userID": user_id,
                "contactMethodID": webhook_cm_id,
                "delayMinutes": 10
            }
        }, session=True)["createUserNotificationRule"]
        nr2_id = nr2["id"]
        print(f"Created delayed notification rule: id={nr2_id}, delay={nr2['delayMinutes']}", flush=True)

        # Check user notification rules query
        user_with_nrs_q = """
        query UserWithNRs($id: ID!) {
          user(id: $id) {
            id
            notificationRules {
              id
              delayMinutes
              contactMethod {
                id
              }
            }
          }
        }
        """
        user_nrs = f.graphql(query=user_with_nrs_q, variables={"id": user_id}, session=True)["user"]["notificationRules"]
        print(f"User notification rules: {user_nrs}", flush=True)
        assert len(user_nrs) == 2

        # Test deletions via deleteAll
        delete_nr_q = """
        mutation DeleteNR($id: ID!) {
          deleteAll(input: [{type: notificationRule, id: $id}])
        }
        """
        del_nr_res = f.graphql(query=delete_nr_q, variables={"id": nr1_id}, session=True)
        print(f"Deleted notification rule: {del_nr_res}", flush=True)

        delete_cm_q = """
        mutation DeleteCM($id: ID!) {
          deleteAll(input: [{type: contactMethod, id: $id}])
        }
        """
        del_cm_res = f.graphql(query=delete_cm_q, variables={"id": webhook_cm_id}, session=True)
        print(f"Deleted contact method: {del_cm_res}", flush=True)

        delete_user_q = """
        mutation DeleteUser($id: ID!) {
          deleteAll(input: [{type: user, id: $id}])
        }
        """
        del_user_res = f.graphql(query=delete_user_q, variables={"id": user_id}, session=True)
        print(f"Deleted user: {del_user_res}", flush=True)

        # Verify read after deletion returns null
        after = f.graphql(query=read_user_q, variables={"id": user_id}, session=True)["user"]
        assert after is None
        print("PASS: user read after deletion returned null.", flush=True)

        print("\nALL USER FEASIBILITY PROBES PASSED!", flush=True)


if __name__ == "__main__":
    probe()
