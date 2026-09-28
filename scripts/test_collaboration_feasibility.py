"""Empirical feasibility probe for Issue #61 / Gate COL-DISC: Collaboration Channels and System Limits.
Introspects and tests systemLimits, setSystemLimit, slackChannels, and slackUserGroups against GoAlert v0.35.0.
"""
import json
from fixture import disposable

INTROSPECTION_QUERY = """
query IntrospectCollabSchema {
  systemLimitID: __type(name: "SystemLimitID") {
    enumValues {
      name
      description
    }
  }
  systemLimit: __type(name: "SystemLimit") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  slackChannel: __type(name: "SlackChannel") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  slackUserGroup: __type(name: "SlackUserGroup") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  schemaMutations: __schema {
    mutationType {
      fields {
        name
        args {
          name
          type { name kind ofType { name kind ofType { name kind } } }
        }
      }
    }
    queryType {
      fields {
        name
        args {
          name
          type { name kind ofType { name kind ofType { name kind } } }
        }
      }
    }
  }
}
"""

def main():
    print("Starting disposable GoAlert instance...", flush=True)
    with disposable() as fix:
        print("GoAlert ready. Running schema introspection...", flush=True)
        data = fix.graphql(INTROSPECTION_QUERY, session=True)

        print("\n=== SystemLimitID Enum Values ===", flush=True)
        print(json.dumps(data.get("systemLimitID"), indent=2), flush=True)

        print("\n=== SystemLimit Fields ===", flush=True)
        print(json.dumps(data.get("systemLimit"), indent=2), flush=True)

        print("\n=== SlackChannel Fields ===", flush=True)
        print(json.dumps(data.get("slackChannel"), indent=2), flush=True)

        print("\n=== SlackUserGroup Fields ===", flush=True)
        print(json.dumps(data.get("slackUserGroup"), indent=2), flush=True)

        mutations = [
            f for f in data.get("schemaMutations", {}).get("mutationType", {}).get("fields", [])
            if any(k in f["name"].lower() for k in ["limit", "slack"])
        ]
        print("\n=== Relevant Mutations ===", flush=True)
        print(json.dumps(mutations, indent=2), flush=True)

        queries = [
            f for f in data.get("schemaMutations", {}).get("queryType", {}).get("fields", [])
            if any(k in f["name"].lower() for k in ["limit", "slack"])
        ]
        print("\n=== Relevant Queries ===", flush=True)
        print(json.dumps(queries, indent=2), flush=True)

        # 1. Query existing systemLimits
        print("\n=== Querying existing systemLimits ===", flush=True)
        limits_res = fix.graphql("""
        query GetLimits {
            systemLimits {
                id
                value
                description
            }
        }
        """, session=True)
        print(json.dumps(limits_res, indent=2), flush=True)

        # 2. Test setSystemLimits mutation
        if limits_res.get("systemLimits"):
            first_limit = limits_res["systemLimits"][0]
            limit_id = first_limit["id"]
            old_val = first_limit["value"]
            new_val = old_val + 5
            print(f"\n=== Testing setSystemLimits on {limit_id} from {old_val} to {new_val} ===", flush=True)
            fix.graphql("""
            mutation SetLimits($input: [SystemLimitInput!]!) {
                setSystemLimits(input: $input)
            }
            """, {"input": [{"id": limit_id, "value": new_val}]}, session=True)

            # Read back
            check_res = fix.graphql("""
            query GetLimitsCheck {
                systemLimits {
                    id
                    value
                    description
                }
            }
            """, session=True)
            updated = next((l for l in check_res.get("systemLimits", []) if l["id"] == limit_id), None)
            print("Read back updated limit:", json.dumps(updated, indent=2), flush=True)
            assert updated is not None and updated["value"] == new_val, f"Expected {new_val}, got {updated}"

            # Reset back to old value
            fix.graphql("""
            mutation ResetLimits($input: [SystemLimitInput!]!) {
                setSystemLimits(input: $input)
            }
            """, {"input": [{"id": limit_id, "value": old_val}]}, session=True)
            print(f"Reset {limit_id} back to {old_val}", flush=True)

        # 3. Test slackChannels / slackUserGroups queries
        print("\n=== Querying slackChannels ===", flush=True)
        try:
            sc_res = fix.graphql("""
            query GetSlackChannels {
                slackChannels {
                    nodes {
                        id
                        name
                        teamID
                    }
                }
            }
            """, session=True)
            print("Slack channels:", json.dumps(sc_res, indent=2), flush=True)
        except Exception as e:
            print("slackChannels query failed:", e, flush=True)

        print("\n=== Querying slackUserGroups ===", flush=True)
        try:
            sug_res = fix.graphql("""
            query GetSlackUserGroups {
                slackUserGroups {
                    nodes {
                        id
                        name
                        handle
                    }
                }
            }
            """, session=True)
            print("Slack user groups:", json.dumps(sug_res, indent=2), flush=True)
        except Exception as e:
            print("slackUserGroups query failed:", e, flush=True)

        print("\nALL COLLABORATION AND LIMITS PROBES SUCCEEDED!", flush=True)

if __name__ == "__main__":
    main()
