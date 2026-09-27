"""Empirical feasibility probe for Issue #43 / Gate ROT-DISC: Rotations and Escalation Targets.
Validates GraphQL operations, schema types, boundaries, and step targets against disposable GoAlert v0.35.0.
"""
import json
import uuid
from fixture import disposable

INTROSPECTION_QUERY = """
query IntrospectRotationSchema {
  rotationType: __type(name: "RotationType") {
    enumValues {
      name
      description
    }
  }
  createRotationInput: __type(name: "CreateRotationInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  updateRotationInput: __type(name: "UpdateRotationInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  rotation: __type(name: "Rotation") {
    fields {
      name
      type { name kind ofType { name kind } }
    }
  }
  targetType: __type(name: "TargetType") {
    enumValues {
      name
    }
  }
  targetInput: __type(name: "TargetInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  createStepInput: __type(name: "CreateEscalationPolicyStepInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  updateStepInput: __type(name: "UpdateEscalationPolicyStepInput") {
    inputFields {
      name
      type { name kind ofType { name kind } }
    }
  }
  escalationPolicyStep: __type(name: "EscalationPolicyStep") {
    fields {
      name
      type { name kind ofType { name kind } }
    }
  }
}
"""

def main():
    print("Starting disposable GoAlert instance...")
    with disposable() as fix:
        print("GoAlert started. Running schema introspection...")
        data = fix.graphql(INTROSPECTION_QUERY, session=True)
        
        print("\n=== RotationType Enum Values ===")
        print(json.dumps(data.get("rotationType"), indent=2))
        
        print("\n=== CreateRotationInput ===")
        print(json.dumps(data.get("createRotationInput"), indent=2))
        
        print("\n=== UpdateRotationInput ===")
        print(json.dumps(data.get("updateRotationInput"), indent=2))
        
        print("\n=== Rotation Fields ===")
        print(json.dumps(data.get("rotation"), indent=2))

        print("\n=== TargetType Enum Values ===")
        print(json.dumps(data.get("targetType"), indent=2))

        print("\n=== TargetInput Fields ===")
        print(json.dumps(data.get("targetInput"), indent=2))

        print("\n=== CreateEscalationPolicyStepInput ===")
        print(json.dumps(data.get("createStepInput"), indent=2))

        print("\n=== UpdateEscalationPolicyStepInput ===")
        print(json.dumps(data.get("updateStepInput"), indent=2))

        print("\n=== EscalationPolicyStep Fields ===")
        print(json.dumps(data.get("escalationPolicyStep"), indent=2))

        # Test user creation
        print("\n=== Creating Test Users ===")
        u1_res = fix.graphql("""
        mutation CreateU1($input: CreateUserInput!) {
          createUser(input: $input) { id name }
        }
        """, {"input": {"name": "Operator Alpha", "email": "alpha@example.com", "role": "user", "username": "op_alpha", "password": "Password123!"}}, session=True)
        u1_id = u1_res["createUser"]["id"]
        print(f"Created user 1: {u1_id}")

        u2_res = fix.graphql("""
        mutation CreateU2($input: CreateUserInput!) {
          createUser(input: $input) { id name }
        }
        """, {"input": {"name": "Operator Bravo", "email": "bravo@example.com", "role": "user", "username": "op_bravo", "password": "Password123!"}}, session=True)
        u2_id = u2_res["createUser"]["id"]
        print(f"Created user 2: {u2_id}")

        # Test rotation creation
        print("\n=== Testing createRotation ===")
        create_rot_res = fix.graphql("""
        mutation CreateRot($input: CreateRotationInput!) {
          createRotation(input: $input) {
            id
            name
            description
            type
            start
            timeZone
            shiftLength
            userIDs
          }
        }
        """, {
            "input": {
                "name": "Primary SRE On-Call",
                "description": "Daily 24h primary rotation",
                "type": "daily",
                "start": "2026-10-01T09:00:00Z",
                "timeZone": "Europe/Madrid",
                "shiftLength": 1,
                "userIDs": [u1_id, u2_id]
            }
        }, session=True)
        rot = create_rot_res["createRotation"]
        rot_id = rot["id"]
        print(f"Created rotation: {json.dumps(rot, indent=2)}")

        # Test query rotation
        print("\n=== Testing query rotation ===")
        query_rot_res = fix.graphql("""
        query GetRot($id: ID!) {
          rotation(id: $id) {
            id
            name
            description
            type
            start
            timeZone
            shiftLength
            userIDs
            users { id name }
          }
        }
        """, {"id": rot_id}, session=True)
        print(f"Queried rotation: {json.dumps(query_rot_res['rotation'], indent=2)}")

        # Test updateRotation (reorder participants and update name)
        print("\n=== Testing updateRotation (reorder users) ===")
        update_rot_res = fix.graphql("""
        mutation UpdateRot($input: UpdateRotationInput!) {
          updateRotation(input: $input)
        }
        """, {
            "input": {
                "id": rot_id,
                "name": "Primary SRE On-Call Updated",
                "userIDs": [u2_id, u1_id]
            }
        }, session=True)
        print(f"Update rotation result: {json.dumps(update_rot_res, indent=2)}")

        # Verify update
        query_rot_updated = fix.graphql("""
        query GetRot($id: ID!) {
          rotation(id: $id) {
            name
            userIDs
          }
        }
        """, {"id": rot_id}, session=True)
        print(f"After update: {json.dumps(query_rot_updated['rotation'], indent=2)}")
        assert query_rot_updated["rotation"]["userIDs"] == [u2_id, u1_id]

        # Test Escalation Policy with User and Rotation Targets
        print("\n=== Testing Escalation Policy with Targets ===")
        ep_res = fix.graphql("""
        mutation CreateEP($input: CreateEscalationPolicyInput!) {
          createEscalationPolicy(input: $input) { id name }
        }
        """, {"input": {"name": "Targeted Escalation Policy", "repeat": 2}}, session=True)
        ep_id = ep_res["createEscalationPolicy"]["id"]
        print(f"Created EP: {ep_id}")

        # Add step with targets
        print("\n=== Adding step targeting user and rotation ===")
        step_res = fix.graphql("""
        mutation AddStep($input: CreateEscalationPolicyStepInput!) {
          createEscalationPolicyStep(input: $input) {
            id
            stepNumber
            delayMinutes
            actions {
              type
              args
            }
          }
        }
        """, {
            "input": {
                "escalationPolicyID": ep_id,
                "delayMinutes": 15,
                "targets": [
                    {"id": u1_id, "type": "user"},
                    {"id": rot_id, "type": "rotation"}
                ]
            }
        }, session=True)
        print(f"Created step with targets: {json.dumps(step_res, indent=2)}")

        # Query EP with steps and targets
        print("\n=== Querying EP with steps and targets ===")
        ep_query_res = fix.graphql("""
        query GetEP($id: ID!) {
          escalationPolicy(id: $id) {
            id
            steps {
              id
              delayMinutes
              actions {
                type
                args
              }
            }
          }
        }
        """, {"id": ep_id}, session=True)
        print(f"Queried EP: {json.dumps(ep_query_res['escalationPolicy'], indent=2)}")

        # Test deletion
        print("\n=== Testing deleteRotation ===")
        del_rot_res = fix.graphql("""
        mutation DelRot($id: ID!) {
          deleteAll(input: [{type: rotation, id: $id}])
        }
        """, {"id": rot_id}, session=True)
        print(f"Deleted rotation result: {json.dumps(del_rot_res, indent=2)}")

        print("\nFEASIBILITY PROBE PASSED ALL CHECKS!")

if __name__ == "__main__":
    main()
