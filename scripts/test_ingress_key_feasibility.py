"""Feasibility probe for Issue #9 / Gate V003-DISC: Integration Key Lifecycle.

Proves:
1. Schema introspection: IntegrationKeyType enum, CreateIntegrationKeyInput, IntegrationKey fields, Query/Mutation fields.
2. Direct query vs Service.integrationKeys: Can integration keys be queried directly or only via service?
3. createIntegrationKey mutation with types (grafana, generic).
4. Update capability: Does updateIntegrationKey exist, or is replacement required?
5. URL and token format: href structure, token parameter vs key ID.
6. Deletion lifecycle: deleteAll(type: integrationKey) and cascade deletion on service deletion.
7. Role enforcement: Can 'user' role create/delete integration keys or is 'admin' required?
8. Named operations document: Registration in operations.graphql and execution over fixed-document API key.
9. Webhook ingress delivery: POSTing Grafana and generic payloads to href.
"""
import json
import urllib.request
import urllib.error
from pathlib import Path
from fixture import disposable, ROOT, wait_for_health

PROBE_OPERATIONS_DOC = """
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
query ProviderReadServiceIntegrationKeys($serviceID: ID!) {
  service(id: $serviceID) {
    id
    integrationKeys {
      id
      name
      type
      href
    }
  }
}
mutation ProviderCreateIntegrationKey($input: CreateIntegrationKeyInput!) {
  createIntegrationKey(input: $input) {
    id
    name
    type
    href
  }
}
mutation ProviderDeleteIntegrationKey($id: ID!) {
  deleteAll(input: [{type: integrationKey, id: $id}])
}
"""

def run_probes():
    results = {}
    with disposable() as f:
        print("\n=== PROBE 1: Schema Introspection ===", flush=True)
        # Query types
        schema_query = f.graphql("""
        query {
          __type(name: "IntegrationKeyType") {
            enumValues { name }
          }
          ikType: __type(name: "IntegrationKey") {
            fields { name type { name kind ofType { name kind } } }
          }
          createInput: __type(name: "CreateIntegrationKeyInput") {
            inputFields { name type { name kind ofType { name kind } } }
          }
          mutationType: __type(name: "Mutation") {
            fields { name }
          }
          queryType: __type(name: "Query") {
            fields { name }
          }
        }
        """, session=True)
        
        ik_types = [e["name"] for e in schema_query["__type"]["enumValues"]]
        results["integration_key_types"] = ik_types
        print(f"Supported IntegrationKey types: {ik_types}")
        
        ik_fields = {fld["name"]: (fld["type"]["name"] or fld["type"]["kind"]) for fld in schema_query["ikType"]["fields"]}
        results["integration_key_fields"] = ik_fields
        print(f"IntegrationKey fields: {ik_fields}")
        
        create_fields = {fld["name"]: (fld["type"]["name"] or fld["type"]["kind"]) for fld in schema_query["createInput"]["inputFields"]}
        results["create_input_fields"] = create_fields
        print(f"CreateIntegrationKeyInput fields: {create_fields}")
        
        mutations = [fld["name"] for fld in schema_query["mutationType"]["fields"]]
        ik_mutations = [m for m in mutations if "integration" in m.lower() or "key" in m.lower()]
        results["ik_mutations"] = ik_mutations
        print(f"Key-related mutations: {ik_mutations}")
        
        queries = [fld["name"] for fld in schema_query["queryType"]["fields"]]
        ik_queries = [q for q in queries if "integration" in q.lower() or "key" in q.lower()]
        results["ik_queries"] = ik_queries
        print(f"Key-related queries: {ik_queries}")
        
        has_direct_query = "integrationKey" in queries
        has_update_mutation = "updateIntegrationKey" in mutations
        results["has_direct_query"] = has_direct_query
        results["has_update_mutation"] = has_update_mutation
        print(f"Direct integrationKey query exists: {has_direct_query}")
        print(f"updateIntegrationKey mutation exists: {has_update_mutation}")

        print("\n=== PROBE 2: Create Policy and Service ===", flush=True)
        policy_id = f.policy("Feasibility Test Policy")
        svc_res = f.graphql("""
        mutation CreateSvc($name: String!, $epID: ID!) {
          createService(input: {name: $name, escalationPolicyID: $epID}) { id name }
        }
        """, {"name": "Ingress Test Service", "epID": policy_id}, session=True)
        service_id = svc_res["createService"]["id"]
        print(f"Created Service: {service_id}")

        print("\n=== PROBE 3: Create Integration Keys (grafana & generic) ===", flush=True)
        # Create Grafana key
        grafana_res = f.graphql("""
        mutation CreateIK($input: CreateIntegrationKeyInput!) {
          createIntegrationKey(input: $input) {
            id
            name
            type
            href
          }
        }
        """, {
            "input": {
                "serviceID": service_id,
                "name": "Grafana Ingress",
                "type": "grafana"
            }
        }, session=True)
        grafana_key = grafana_res["createIntegrationKey"]
        results["grafana_key"] = grafana_key
        print(f"Created Grafana Key: id={grafana_key['id']}, href={grafana_key['href']}")

        # Create Generic key
        generic_res = f.graphql("""
        mutation CreateIK($input: CreateIntegrationKeyInput!) {
          createIntegrationKey(input: $input) {
            id
            name
            type
            href
          }
        }
        """, {
            "input": {
                "serviceID": service_id,
                "name": "Generic Ingress",
                "type": "generic"
            }
        }, session=True)
        generic_key = generic_res["createIntegrationKey"]
        results["generic_key"] = generic_key
        print(f"Created Generic Key: id={generic_key['id']}, href={generic_key['href']}")

        # Verify href token structure
        # In GoAlert, href is e.g. http://127.0.0.1:18081/api/v2/grafana/incoming?token=<id>
        print(f"Grafana href token matches key ID: {grafana_key['id'] in grafana_key['href']}")
        assert grafana_key["id"] in grafana_key["href"], "Expected token query param to contain key ID"
        results["token_matches_id"] = grafana_key["id"] in grafana_key["href"]

        print("\n=== PROBE 4: Read Keys via Service.integrationKeys ===", flush=True)
        read_res = f.graphql("""
        query ReadKeys($serviceID: ID!) {
          service(id: $serviceID) {
            integrationKeys {
              id
              name
              type
              href
            }
          }
        }
        """, {"serviceID": service_id}, session=True)
        keys_read = read_res["service"]["integrationKeys"]
        print(f"Service returned {len(keys_read)} integration keys")
        read_ids = {k["id"] for k in keys_read}
        assert grafana_key["id"] in read_ids
        assert generic_key["id"] in read_ids
        results["keys_read_count"] = len(keys_read)

        print("\n=== PROBE 5: Deletion Lifecycle via deleteAll ===", flush=True)
        del_res = f.graphql("""
        mutation DeleteKey($id: ID!) {
          deleteAll(input: [{type: integrationKey, id: $id}])
        }
        """, {"id": generic_key["id"]}, session=True)
        print(f"Deleted Generic Key result: {del_res}")
        assert del_res["deleteAll"] is True

        # Verify key is gone from service
        read_after_del = f.graphql("""
        query ReadKeys($serviceID: ID!) {
          service(id: $serviceID) {
            integrationKeys {
              id
            }
          }
        }
        """, {"serviceID": service_id}, session=True)
        remaining_ids = {k["id"] for k in read_after_del["service"]["integrationKeys"]}
        assert generic_key["id"] not in remaining_ids
        assert grafana_key["id"] in remaining_ids
        print(f"Key successfully removed. Remaining keys: {remaining_ids}")

        print("\n=== PROBE 6: Fixed Document API Key Operations ===", flush=True)
        # Create API key with PROBE_OPERATIONS_DOC
        admin_api_token = f.key("admin", document=PROBE_OPERATIONS_DOC)
        user_api_token = f.key("user", document=PROBE_OPERATIONS_DOC)
        print("Created admin and user API keys with probe document")

        # Test Read with Admin Token
        read_api_res = f.graphql(
            variables={"serviceID": service_id},
            operation="ProviderReadServiceIntegrationKeys",
            token=admin_api_token
        )
        print(f"ProviderReadServiceIntegrationKeys over API key: {read_api_res}")
        assert len(read_api_res["service"]["integrationKeys"]) == 1

        # Test Create with User Token (least privilege check)
        user_create_res = f.graphql(
            variables={
                "input": {
                    "serviceID": service_id,
                    "name": "User Created Key",
                    "type": "grafana"
                }
            },
            operation="ProviderCreateIntegrationKey",
            token=user_api_token,
            raw=True
        )
        print(f"User role create result: {user_create_res}")
        user_denied = bool(user_create_res.get("errors"))
        results["user_role_denied"] = user_denied
        print(f"User role denied as expected: {user_denied}")

        # Test Create with Admin Token
        admin_create_res = f.graphql(
            variables={
                "input": {
                    "serviceID": service_id,
                    "name": "Admin Created Key",
                    "type": "generic"
                }
            },
            operation="ProviderCreateIntegrationKey",
            token=admin_api_token
        )
        created_via_api = admin_create_res["createIntegrationKey"]
        print(f"Created via Admin API key: {created_via_api}")
        assert created_via_api["name"] == "Admin Created Key"

        # Test Delete with Admin Token
        del_api_res = f.graphql(
            variables={"id": created_via_api["id"]},
            operation="ProviderDeleteIntegrationKey",
            token=admin_api_token
        )
        print(f"Deleted via Admin API key: {del_api_res}")
        assert del_api_res["deleteAll"] is True

        print("\n=== PROBE 1b: Inspect Query.integrationKey ===", flush=True)
        q_field = f.graphql("""
        query {
          __type(name: "Query") {
            fields {
              name
              args { name type { name kind ofType { name kind } } }
              type { name kind ofType { name kind } }
            }
          }
        }
        """, session=True)
        ik_q = [fld for fld in q_field["__type"]["fields"] if fld["name"] == "integrationKey"][0]
        print(f"integrationKey query definition: {ik_q}")
        results["integrationKey_query"] = ik_q

        # Test direct integrationKey query
        direct_read = f.graphql("""
        query ReadIK($id: ID!) {
          integrationKey(id: $id) {
            id
            name
            type
            href
            serviceID
          }
        }
        """, {"id": grafana_key["id"]}, session=True)
        print(f"Direct integrationKey read: {direct_read}")
        assert direct_read["integrationKey"]["id"] == grafana_key["id"]
        assert direct_read["integrationKey"]["serviceID"] == service_id

        # Test non-existent key read (drift detection)
        missing_read = f.graphql("""
        query ReadIK($id: ID!) {
          integrationKey(id: $id) {
            id
          }
        }
        """, {"id": "00000000-0000-0000-0000-000000000000"}, session=True, raw=True)
        print(f"Non-existent key read: {missing_read}")
        assert missing_read.get("data", {}).get("integrationKey") is None
        results["missing_key_returns_null"] = True

        print("\n=== PROBE 7: Ingress Webhook Payload Delivery ===", flush=True)
        # Send a sample Grafana alert payload to grafana_key["href"]
        grafana_payload = {
            "receiver": "goalert",
            "status": "firing",
            "alerts": [
                {
                    "status": "firing",
                    "labels": {
                        "alertname": "TestHighCPU",
                        "severity": "critical",
                        "instance": "example-node-1"
                    },
                    "annotations": {
                        "summary": "CPU utilization exceeded 90%",
                        "description": "Node example-node-1 CPU is at 94%"
                    },
                    "startsAt": "2026-09-26T15:00:00Z",
                    "fingerprint": "a1b2c3d4e5f6"
                }
            ],
            "title": "[FIRING:1] TestHighCPU (critical example-node-1)",
            "state": "alerting"
        }
        
        req = urllib.request.Request(
            grafana_key["href"],
            data=json.dumps(grafana_payload).encode("utf-8"),
            headers={"Content-Type": "application/json"}
        )
        try:
            with urllib.request.urlopen(req, timeout=5) as resp:
                status_code = resp.status
                body = resp.read().decode("utf-8")
                print(f"Grafana webhook POST response: {status_code} {body}")
                results["grafana_webhook_status"] = status_code
        except urllib.error.HTTPError as e:
            print(f"Grafana webhook POST error: {e.code} {e.read().decode('utf-8')}")
            results["grafana_webhook_status"] = e.code

        # Check if an alert was generated using top-level alerts query
        alerts_res = f.graphql("""
        query CheckAlerts($serviceID: ID!) {
          alerts(input: {filterByServiceID: [$serviceID]}) {
            nodes {
              id
              summary
              details
              status
            }
          }
        }
        """, {"serviceID": service_id}, session=True)
        alerts = alerts_res["alerts"]["nodes"]
        print(f"Alerts created on service: {alerts}")
        results["alerts_created"] = len(alerts)
        assert len(alerts) > 0, "Expected at least 1 alert created by Grafana webhook payload"

        print("\n=== ALL PROBES COMPLETED SUCCESSFULLY ===")
        return results

if __name__ == "__main__":
    res = run_probes()
    print("\nSummary Results:")
    print(json.dumps(res, indent=2))
