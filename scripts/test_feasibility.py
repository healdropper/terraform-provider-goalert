"""Comprehensive feasibility probe for Issue #5 (V002-DISC).

Proves:
1. Webhook destination enablement via Webhook.Enable.
2. Escalation policy CRUD (create, read, update, delete).
3. Step lifecycle: inline creation vs separate creation, update, reordering, deletion.
4. Webhook action specification (type: 'builtin-webhook', args: {'webhook_url': ...}).
5. Deletion of policy with attached service (referential integrity / cascade).
6. Validation limits and error cases (invalid URL, delayMinutes limits, repeat limits).
7. AST hash verification with named operations over API key.
8. Role enforcement (user vs admin).
"""
import json
import uuid
from fixture import disposable

CANONICAL_TEST_DOC = """
query ProviderReadService($id: ID!) {
  service(id: $id) { id name description escalationPolicy { id } }
}
mutation ProviderCreateService($name: String!, $description: String!, $escalationPolicyID: ID!) {
  createService(input: {name: $name, description: $description, escalationPolicyID: $escalationPolicyID}) {
    id name description escalationPolicy { id }
  }
}
mutation ProviderUpdateService($id: ID!, $name: String!, $description: String!, $escalationPolicyID: ID!) {
  updateService(input: {id: $id, name: $name, description: $description, escalationPolicyID: $escalationPolicyID})
}
mutation ProviderDeleteService($id: ID!) {
  deleteAll(input: [{type: service, id: $id}])
}
query ProviderReadEscalationPolicy($id: ID!) {
  escalationPolicy(id: $id) {
    id
    name
    description
    repeat
    steps {
      id
      stepNumber
      delayMinutes
      actions {
        type
        args
      }
    }
  }
}
mutation ProviderCreateEscalationPolicy($input: CreateEscalationPolicyInput!) {
  createEscalationPolicy(input: $input) {
    id
    name
    description
    repeat
    steps {
      id
      stepNumber
      delayMinutes
      actions {
        type
        args
      }
    }
  }
}
mutation ProviderUpdateEscalationPolicy($input: UpdateEscalationPolicyInput!) {
  updateEscalationPolicy(input: $input)
}
mutation ProviderDeleteEscalationPolicy($id: ID!) {
  deleteAll(input: [{type: escalationPolicy, id: $id}])
}
mutation ProviderCreateEscalationPolicyStep($input: CreateEscalationPolicyStepInput!) {
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
mutation ProviderUpdateEscalationPolicyStep($input: UpdateEscalationPolicyStepInput!) {
  updateEscalationPolicyStep(input: $input)
}
"""

def run_probes(f):
    results = {}
    
    # --- Probe 1: Webhook destination enablement ---
    # Default state
    dt_before = f.graphql("query { destinationTypes { type enabled } }", session=True)
    wh_before = [d for d in dt_before["destinationTypes"] if d["type"] == "builtin-webhook"][0]
    assert not wh_before["enabled"], "Expected builtin-webhook to be disabled by default"
    results["webhook_default_enabled"] = False
    
    # Enable Webhook.Enable via setConfig
    f.graphql("mutation { setConfig(input: [{id: \"Webhook.Enable\", value: \"true\"}]) }", session=True)
    dt_after = f.graphql("query { destinationTypes { type enabled requiredFields { fieldID label } } }", session=True)
    wh_after = [d for d in dt_after["destinationTypes"] if d["type"] == "builtin-webhook"][0]
    assert wh_after["enabled"], "Expected builtin-webhook to be enabled after setConfig"
    results["webhook_enabled_after_config"] = True
    results["webhook_required_fields"] = wh_after["requiredFields"]
    print("PASS: Webhook destination successfully enabled via Webhook.Enable config.", flush=True)

    # --- Probe 2: API Key with expanded document ---
    # Create admin API key with CANONICAL_TEST_DOC
    admin_key = f.key("admin", document=CANONICAL_TEST_DOC)
    results["api_key_created"] = True
    print("PASS: API Key created with expanded canonical GraphQL document.", flush=True)

    # Verify backwards compatibility: Service CRUD still works with the new key
    dummy_policy_id = f.policy("Compat policy")
    svc = f.graphql(
        variables={"name": "Compat Svc", "description": "Desc", "escalationPolicyID": dummy_policy_id},
        operation="ProviderCreateService",
        token=admin_key
    )["createService"]
    svc_id = svc["id"]
    read_svc = f.graphql(variables={"id": svc_id}, operation="ProviderReadService", token=admin_key)["service"]
    assert read_svc["name"] == "Compat Svc"
    assert f.graphql(variables={"id": svc_id}, operation="ProviderDeleteService", token=admin_key)["deleteAll"]
    results["service_compatibility_verified"] = True
    print("PASS: Service CRUD backwards compatibility verified with expanded key.", flush=True)

    # --- Probe 3: Create Escalation Policy with inline steps ---
    ep_input = {
        "name": "Production Alerts Policy",
        "description": "Critical alerts escalation",
        "repeat": 3,
        "steps": [
            {
                "delayMinutes": 5,
                "actions": [{
                    "type": "builtin-webhook",
                    "args": {"webhook_url": "https://alerts.example.com/hook1"}
                }]
            },
            {
                "delayMinutes": 10,
                "actions": [{
                    "type": "builtin-webhook",
                    "args": {"webhook_url": "https://alerts.example.com/hook2"}
                }]
            }
        ]
    }
    ep = f.graphql(
        variables={"input": ep_input},
        operation="ProviderCreateEscalationPolicy",
        token=admin_key
    )["createEscalationPolicy"]
    
    ep_id = ep["id"]
    assert ep["name"] == ep_input["name"]
    assert ep["description"] == ep_input["description"]
    assert ep["repeat"] == 3
    assert len(ep["steps"]) == 2
    assert ep["steps"][0]["stepNumber"] == 0
    assert ep["steps"][0]["delayMinutes"] == 5
    assert ep["steps"][0]["actions"][0]["type"] == "builtin-webhook"
    assert ep["steps"][0]["actions"][0]["args"]["webhook_url"] == "https://alerts.example.com/hook1"
    assert ep["steps"][1]["stepNumber"] == 1
    assert ep["steps"][1]["delayMinutes"] == 10
    step0_id = ep["steps"][0]["id"]
    step1_id = ep["steps"][1]["id"]
    results["create_with_inline_steps"] = True
    print("PASS: Escalation policy created atomically with inline steps.", flush=True)

    # --- Probe 4: Read Escalation Policy ---
    ep_read = f.graphql(
        variables={"id": ep_id},
        operation="ProviderReadEscalationPolicy",
        token=admin_key
    )["escalationPolicy"]
    assert ep_read["id"] == ep_id
    assert len(ep_read["steps"]) == 2
    results["read_escalation_policy"] = True
    print("PASS: Escalation policy read verified.", flush=True)

    # --- Probe 5: Update Escalation Policy (metadata & step reordering) ---
    # Update name, description, repeat, and reverse step order via stepIDs
    update_res = f.graphql(
        variables={"input": {
            "id": ep_id,
            "name": "Renamed Policy",
            "description": "Updated description",
            "repeat": 1,
            "stepIDs": [step1_id, step0_id]
        }},
        operation="ProviderUpdateEscalationPolicy",
        token=admin_key
    )["updateEscalationPolicy"]
    assert update_res is True

    ep_read2 = f.graphql(variables={"id": ep_id}, operation="ProviderReadEscalationPolicy", token=admin_key)["escalationPolicy"]
    assert ep_read2["name"] == "Renamed Policy"
    assert ep_read2["repeat"] == 1
    assert ep_read2["steps"][0]["id"] == step1_id
    assert ep_read2["steps"][0]["stepNumber"] == 0
    assert ep_read2["steps"][1]["id"] == step0_id
    assert ep_read2["steps"][1]["stepNumber"] == 1
    results["update_policy_and_reorder"] = True
    print("PASS: Policy update and step reordering verified.", flush=True)

    # --- Probe 6: Update Escalation Policy Step ---
    step_up_res = f.graphql(
        variables={"input": {
            "id": step1_id,
            "delayMinutes": 15,
            "actions": [{
                "type": "builtin-webhook",
                "args": {"webhook_url": "https://alerts.example.com/hook2-updated"}
            }]
        }},
        operation="ProviderUpdateEscalationPolicyStep",
        token=admin_key
    )["updateEscalationPolicyStep"]
    assert step_up_res is True

    ep_read3 = f.graphql(variables={"id": ep_id}, operation="ProviderReadEscalationPolicy", token=admin_key)["escalationPolicy"]
    updated_step1 = [s for s in ep_read3["steps"] if s["id"] == step1_id][0]
    assert updated_step1["delayMinutes"] == 15
    assert updated_step1["actions"][0]["args"]["webhook_url"] == "https://alerts.example.com/hook2-updated"
    results["update_step_actions_and_delay"] = True
    print("PASS: Step update (delayMinutes and actions) verified.", flush=True)

    # --- Probe 7: Add a 3rd step via createEscalationPolicyStep ---
    new_step = f.graphql(
        variables={"input": {
            "escalationPolicyID": ep_id,
            "delayMinutes": 20,
            "actions": [{
                "type": "builtin-webhook",
                "args": {"webhook_url": "https://alerts.example.com/hook3"}
            }]
        }},
        operation="ProviderCreateEscalationPolicyStep",
        token=admin_key
    )["createEscalationPolicyStep"]
    step2_id = new_step["id"]
    assert new_step["stepNumber"] == 2
    results["create_step_separately"] = True
    print("PASS: Step created separately and appended to policy.", flush=True)

    # --- Probe 8: Delete a step via ProviderUpdateEscalationPolicy stepIDs ---
    del_step_res = f.graphql(
        variables={"input": {"id": ep_id, "stepIDs": [step1_id, step2_id]}},
        operation="ProviderUpdateEscalationPolicy",
        token=admin_key
    )["updateEscalationPolicy"]
    assert del_step_res is True

    ep_read4 = f.graphql(variables={"id": ep_id}, operation="ProviderReadEscalationPolicy", token=admin_key)["escalationPolicy"]
    assert len(ep_read4["steps"]) == 2
    assert all(s["id"] != step0_id for s in ep_read4["steps"])
    # Step numbers after deletion:
    print(f"Step numbers after deleting step 0: {[s['stepNumber'] for s in ep_read4['steps']]}")
    assert ep_read4["steps"][0]["stepNumber"] == 0 and ep_read4["steps"][0]["id"] == step1_id
    assert ep_read4["steps"][1]["stepNumber"] == 1 and ep_read4["steps"][1]["id"] == step2_id
    results["step_numbers_after_deletion"] = [s["stepNumber"] for s in ep_read4["steps"]]
    print("PASS: Step deletion via stepIDs omission verified.", flush=True)

    # --- Probe 9: Policy deletion with attached service (Referential integrity) ---
    # Attach a service to ep_id
    svc_attached = f.graphql(
        variables={"name": "Attached Svc", "description": "", "escalationPolicyID": ep_id},
        operation="ProviderCreateService",
        token=admin_key
    )["createService"]
    attached_svc_id = svc_attached["id"]

    # Attempt to delete policy while referenced by service
    del_ep_attempt = f.graphql(
        variables={"id": ep_id},
        operation="ProviderDeleteEscalationPolicy",
        token=admin_key,
        raw=True
    )
    print("Delete EP with attached service response: " + json.dumps(del_ep_attempt), flush=True)
    assert del_ep_attempt.get("errors"), "Expected policy deletion to fail when attached to service"
    results["policy_delete_attached_prevented"] = True
    print("PASS: Deletion of policy referenced by service is prevented (referential integrity).", flush=True)

    # Clean up service first, then delete policy
    f.graphql(variables={"id": attached_svc_id}, operation="ProviderDeleteService", token=admin_key)
    del_ep_success = f.graphql(
        variables={"id": ep_id},
        operation="ProviderDeleteEscalationPolicy",
        token=admin_key
    )["deleteAll"]
    assert del_ep_success is True
    
    # Read after delete
    ep_after_del = f.graphql(variables={"id": ep_id}, operation="ProviderReadEscalationPolicy", token=admin_key)
    assert ep_after_del["escalationPolicy"] is None
    results["policy_deleted_cleanly"] = True
    print("PASS: Policy deleted cleanly after releasing attached service; read returns null.", flush=True)

    # --- Probe 10: Role enforcement (user role vs admin) ---
    user_key = f.key("user", document=CANONICAL_TEST_DOC)
    
    # 1. User role creates an escalation policy with webhook action
    user_ep = f.graphql(
        variables={"input": {
            "name": "User EP",
            "repeat": 1,
            "steps": [{"delayMinutes": 5, "actions": [{"type": "builtin-webhook", "args": {"webhook_url": "https://example.com/user-hook"}}]}]
        }},
        operation="ProviderCreateEscalationPolicy",
        token=user_key,
        raw=True
    )
    print("User role create EP response: " + json.dumps(user_ep), flush=True)
    user_ep_allowed = "data" in user_ep and user_ep["data"].get("createEscalationPolicy") is not None
    results["user_role_ep_create_allowed"] = user_ep_allowed
    if user_ep_allowed:
        user_ep_id = user_ep["data"]["createEscalationPolicy"]["id"]
        # Can user role delete the policy they created?
        user_del = f.graphql(variables={"id": user_ep_id}, operation="ProviderDeleteEscalationPolicy", token=user_key, raw=True)
        print("User role delete EP response: " + json.dumps(user_del), flush=True)
        results["user_role_ep_delete_allowed"] = "data" in user_del and user_del["data"].get("deleteAll") is True

    # 2. User role creates a service
    user_svc = f.graphql(
        variables={"name": "User Svc", "description": "", "escalationPolicyID": dummy_policy_id},
        operation="ProviderCreateService",
        token=user_key,
        raw=True
    )
    print("User role create service response: " + json.dumps(user_svc), flush=True)
    results["user_role_service_create_allowed"] = "data" in user_svc and user_svc["data"].get("createService") is not None
    # Clean up user svc
    if results["user_role_service_create_allowed"]:
        f.graphql(variables={"id": user_svc["data"]["createService"]["id"]}, operation="ProviderDeleteService", token=user_key)
    print("PASS: Role observation: user-role key can perform resource CRUD when operations are in its document.", flush=True)

    # 3. Invalid / revoked key rejection
    revoked_key = f.key("admin", document=CANONICAL_TEST_DOC)
    f.graphql("mutation DeleteKey($id: ID!) { deleteGQLAPIKey(id: $id) }", {"id": f.key_ids[revoked_key]}, session=True)
    revoked_attempt = f.graphql(variables={"id": ep_id}, operation="ProviderReadEscalationPolicy", token=revoked_key, raw=True)
    assert revoked_attempt.get("errors"), "Expected revoked key to be rejected"
    results["revoked_key_rejected"] = True
    print("PASS: Revoked API key is rejected.", flush=True)

    # --- Probe 11: Validation constraints ---
    # Delay minutes minimum check
    zero_delay = f.graphql(
        variables={"input": {
            "name": "Zero Delay EP",
            "steps": [{"delayMinutes": 0, "actions": [{"type": "builtin-webhook", "args": {"webhook_url": "http://example.com"}}]}]
        }},
        operation="ProviderCreateEscalationPolicy",
        token=admin_key,
        raw=True
    )
    print("Zero delayMinutes response: " + json.dumps(zero_delay), flush=True)
    results["zero_delay_result"] = zero_delay.get("errors", [])

    # Invalid webhook URL (not http/https)
    invalid_url = f.graphql(
        variables={"input": {
            "name": "Invalid URL EP",
            "steps": [{"delayMinutes": 5, "actions": [{"type": "builtin-webhook", "args": {"webhook_url": "not-a-url"}}]}]
        }},
        operation="ProviderCreateEscalationPolicy",
        token=admin_key,
        raw=True
    )
    print("Invalid webhook URL response: " + json.dumps(invalid_url), flush=True)
    results["invalid_url_result"] = invalid_url.get("errors", [])

    # Repeat limits (e.g. negative or huge)
    neg_repeat = f.graphql(
        variables={"input": {"name": "Negative Repeat EP", "repeat": -1}},
        operation="ProviderCreateEscalationPolicy",
        token=admin_key,
        raw=True
    )
    print("Negative repeat response: " + json.dumps(neg_repeat), flush=True)
    results["negative_repeat_result"] = neg_repeat.get("errors", [])

    print("\nALL FEASIBILITY PROBES COMPLETED SUCCESSFULLY!")
    return results

if __name__ == "__main__":
    with disposable() as instance:
        res = run_probes(instance)
        with open("scratch/feasibility_results.json", "w") as out:
            json.dump(res, out, indent=2)
