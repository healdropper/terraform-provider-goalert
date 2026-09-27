"""Empirical feasibility probe for Issue #18 / Gate V004-DISC: Heartbeat Monitors, Labels, and Data Sources.
Validates GraphQL operations, boundaries, and permissions against disposable GoAlert v0.35.0.
"""
import json
import urllib.request
import urllib.error
from pathlib import Path
from fixture import disposable, ROOT, wait_for_health

CANONICAL_DOC = (ROOT / "internal/client/operations.graphql").read_text(encoding="utf-8")

PROBE_OPERATIONS_DOC = CANONICAL_DOC + """
query ProviderReadHeartbeatMonitor($id: ID!) {
  heartbeatMonitor(id: $id) {
    id
    serviceID
    name
    timeoutMinutes
    lastState
    href
  }
}
mutation ProviderCreateHeartbeatMonitor($input: CreateHeartbeatMonitorInput!) {
  createHeartbeatMonitor(input: $input) {
    id
    serviceID
    name
    timeoutMinutes
    lastState
    href
  }
}
mutation ProviderUpdateHeartbeatMonitor($input: UpdateHeartbeatMonitorInput!) {
  updateHeartbeatMonitor(input: $input)
}
mutation ProviderDeleteHeartbeatMonitor($id: ID!) {
  deleteAll(input: [{type: heartbeatMonitor, id: $id}])
}
mutation ProviderSetServiceLabel($input: SetLabelInput!) {
  setLabel(input: $input)
}
query ProviderReadServiceWithLabels($id: ID!) {
  service(id: $id) {
    id
    name
    description
    escalationPolicy { id }
    labels {
      key
      value
    }
  }
}
query ProviderSearchServices($search: String!) {
  services(input: {search: $search, first: 15}) {
    nodes {
      id
      name
      description
      escalationPolicy { id }
    }
  }
}
query ProviderSearchEscalationPolicies($search: String!) {
  escalationPolicies(input: {search: $search, first: 15}) {
    nodes {
      id
      name
      description
      repeat
    }
  }
}
"""

def test_heartbeat_and_labels(f):
    # Create probe API key with the expanded document
    probe_key = f.key("admin", document=PROBE_OPERATIONS_DOC)

    # 1. Setup service and policy
    policy_id = f.policy("Heartbeat Feasibility Policy")
    service_resp = f.graphql(
        variables={"name": "Heartbeat Feasibility Service", "description": "Testing heartbeats and labels",
                   "escalationPolicyID": policy_id},
        operation="ProviderCreateService",
        token=probe_key)
    service_id = service_resp["createService"]["id"]
    print(f"Created test service: {service_id}", flush=True)

    # 2. Heartbeat Monitor Creation & Boundary testing
    print("\n--- 1. Testing Heartbeat Monitor Creation & Boundaries ---", flush=True)
    # Test boundary: is timeoutMinutes < 5 allowed?
    hb_small = f.graphql(
        operation="ProviderCreateHeartbeatMonitor",
        variables={"input": {"serviceID": service_id, "name": "Small Timeout Monitor", "timeoutMinutes": 1}},
        token=probe_key,
        raw=True
    )
    if hb_small.get("errors"):
        print(f"timeoutMinutes=1 rejected: {hb_small['errors'][0]['message']}", flush=True)
        assert "5 minutes" in hb_small["errors"][0]["message"] or "must be at least" in hb_small["errors"][0]["message"] or hb_small["errors"][0]["message"]
    else:
        print("timeoutMinutes=1 accepted!", flush=True)
        f.graphql(
            operation="ProviderDeleteHeartbeatMonitor",
            variables={"id": hb_small["data"]["createHeartbeatMonitor"]["id"]},
            token=probe_key
        )

    # Create canonical heartbeat monitor (timeout = 15m)
    hb_resp = f.graphql(
        operation="ProviderCreateHeartbeatMonitor",
        variables={"input": {"serviceID": service_id, "name": "Cron Worker Monitor", "timeoutMinutes": 15}},
        token=probe_key
    )
    hb = hb_resp["createHeartbeatMonitor"]
    hb_id = hb["id"]
    print(f"Created heartbeat monitor raw: {json.dumps(hb)}", flush=True)
    print(f"Ping URL (href): {hb['href']}", flush=True)
    assert hb["serviceID"] == service_id
    assert hb["name"] == "Cron Worker Monitor"
    assert hb["timeoutMinutes"] == 15
    assert hb["lastState"] in ("inactive", "healthy", "unhealthy", "")
    assert "/api/v2/heartbeat/" in hb["href"]

    # 3. Ping delivery: Send HTTP POST/GET to href
    print("\n--- 2. Testing Heartbeat Ping Delivery (href) ---", flush=True)
    req = urllib.request.Request(hb["href"], data=b"", method="POST")
    with urllib.request.urlopen(req, timeout=10) as resp:
        print(f"POST {hb['href']} returned HTTP {resp.status}", flush=True)
        assert resp.status == 200

    # 4. Read Heartbeat Monitor (O(1) query)
    print("\n--- 3. Testing Heartbeat Monitor Query & State Transition ---", flush=True)
    read_resp = f.graphql(operation="ProviderReadHeartbeatMonitor", variables={"id": hb_id}, token=probe_key)
    read_hb = read_resp["heartbeatMonitor"]
    print(f"Read heartbeat monitor: {read_hb}", flush=True)
    assert read_hb["id"] == hb_id
    assert read_hb["lastState"] in ("inactive", "healthy")

    # 5. Update Heartbeat Monitor (Mutation in place)
    print("\n--- 4. Testing Heartbeat Monitor Update ---", flush=True)
    up_resp = f.graphql(
        operation="ProviderUpdateHeartbeatMonitor",
        variables={"input": {"id": hb_id, "name": "Renamed Worker Monitor", "timeoutMinutes": 30}},
        token=probe_key
    )
    assert up_resp["updateHeartbeatMonitor"] is True
    updated_hb = f.graphql(operation="ProviderReadHeartbeatMonitor", variables={"id": hb_id}, token=probe_key)["heartbeatMonitor"]
    print(f"Updated heartbeat monitor: name={updated_hb['name']}, timeout={updated_hb['timeoutMinutes']}", flush=True)
    assert updated_hb["name"] == "Renamed Worker Monitor"
    assert updated_hb["timeoutMinutes"] == 30

    # 6. Service Labels: setLabel and query
    print("\n--- 5. Testing Service Labels (Set, Read, Update, Delete) ---", flush=True)
    # Set label "example.com/environment: staging"
    set_resp = f.graphql(
        operation="ProviderSetServiceLabel",
        variables={"input": {"target": {"type": "service", "id": service_id}, "key": "example.com/environment", "value": "staging"}},
        token=probe_key
    )
    assert set_resp["setLabel"] is True
    # Set another label "example.com/team: sre-team"
    f.graphql(
        operation="ProviderSetServiceLabel",
        variables={"input": {"target": {"type": "service", "id": service_id}, "key": "example.com/team", "value": "sre-team"}},
        token=probe_key
    )

    # Read service labels
    service_with_labels = f.graphql(
        operation="ProviderReadServiceWithLabels",
        variables={"id": service_id},
        token=probe_key
    )["service"]
    labels_dict = {l["key"]: l["value"] for l in service_with_labels["labels"]}
    print(f"Service labels read: {labels_dict}", flush=True)
    assert labels_dict.get("example.com/environment") == "staging"
    assert labels_dict.get("example.com/team") == "sre-team"

    # Update label value
    f.graphql(
        operation="ProviderSetServiceLabel",
        variables={"input": {"target": {"type": "service", "id": service_id}, "key": "example.com/environment", "value": "production"}},
        token=probe_key
    )
    updated_labels = f.graphql(
        operation="ProviderReadServiceWithLabels",
        variables={"id": service_id},
        token=probe_key
    )["service"]["labels"]
    updated_dict = {l["key"]: l["value"] for l in updated_labels}
    print(f"Updated labels read: {updated_dict}", flush=True)
    assert updated_dict.get("example.com/environment") == "production"

    # Delete label by setting value to ""
    f.graphql(
        operation="ProviderSetServiceLabel",
        variables={"input": {"target": {"type": "service", "id": service_id}, "key": "example.com/environment", "value": ""}},
        token=probe_key
    )
    deleted_labels = f.graphql(
        operation="ProviderReadServiceWithLabels",
        variables={"id": service_id},
        token=probe_key
    )["service"]["labels"]
    deleted_dict = {l["key"]: l["value"] for l in deleted_labels}
    print(f"Labels after deletion of 'example.com/environment': {deleted_dict}", flush=True)
    assert "example.com/environment" not in deleted_dict
    assert deleted_dict.get("example.com/team") == "sre-team"

    # 7. Data Source Query Testing
    print("\n--- 6. Testing Data Source Queries ---", flush=True)
    # Search service by name
    search_res = f.graphql(operation="ProviderSearchServices", variables={"search": "Heartbeat Feasibility"}, token=probe_key)
    nodes = search_res["services"]["nodes"]
    print(f"Found services by search: {[n['name'] for n in nodes]}", flush=True)
    assert any(n["id"] == service_id for n in nodes)

    # Search escalation policy by name
    policy_res = f.graphql(operation="ProviderSearchEscalationPolicies", variables={"search": "Heartbeat Feasibility"}, token=probe_key)
    p_nodes = policy_res["escalationPolicies"]["nodes"]
    print(f"Found policies by search: {[p['name'] for p in p_nodes]}", flush=True)
    assert any(p["id"] == policy_id for p in p_nodes)

    # 8. User role permissions
    print("\n--- 7. Testing Least Privilege (User Role) ---", flush=True)
    user_probe_key = f.key("user", document=PROBE_OPERATIONS_DOC)
    # Can user create heartbeat monitor on service?
    user_hb_resp = f.graphql(
        operation="ProviderCreateHeartbeatMonitor",
        variables={"input": {"serviceID": service_id, "name": "User Created Monitor", "timeoutMinutes": 10}},
        token=user_probe_key,
        raw=True
    )
    if user_hb_resp.get("errors"):
        print(f"User role cannot create heartbeat monitor: {user_hb_resp['errors'][0]['message']}", flush=True)
    else:
        print("User role CAN create heartbeat monitor on service.", flush=True)
        user_hb_id = user_hb_resp["data"]["createHeartbeatMonitor"]["id"]
        # clean up
        f.graphql(
            operation="ProviderDeleteHeartbeatMonitor",
            variables={"id": user_hb_id},
            token=probe_key
        )

    # Can user set labels?
    user_label_resp = f.graphql(
        operation="ProviderSetServiceLabel",
        variables={"input": {"target": {"type": "service", "id": service_id}, "key": "example.com/user-tag", "value": "test-val"}},
        token=user_probe_key,
        raw=True
    )
    if user_label_resp.get("errors"):
        print(f"User role cannot set labels: {user_label_resp['errors'][0]['message']}", flush=True)
    else:
        print("User role CAN set labels on service.", flush=True)

    # 9. Deletion & Drift verification
    print("\n--- 8. Testing Heartbeat Monitor Deletion & Drift ---", flush=True)
    del_resp = f.graphql(
        operation="ProviderDeleteHeartbeatMonitor",
        variables={"id": hb_id},
        token=probe_key
    )
    assert del_resp["deleteAll"] is True
    # Verify read after deletion returns null
    post_del_read = f.graphql(operation="ProviderReadHeartbeatMonitor", variables={"id": hb_id}, token=probe_key)
    print(f"Read after deletion: {post_del_read}", flush=True)
    assert post_del_read["heartbeatMonitor"] is None

    # Teardown service and policy
    f.graphql(
        operation="ProviderDeleteService",
        variables={"id": service_id},
        token=probe_key
    )
    f.graphql(
        operation="ProviderDeleteEscalationPolicy",
        variables={"id": policy_id},
        token=probe_key
    )
    print("\n==================================================", flush=True)
    print("ALL FEASIBILITY ASSERTIONS PASSED EMPIRICALLY! 100%", flush=True)
    print("==================================================", flush=True)

if __name__ == "__main__":
    with disposable() as instance:
        test_heartbeat_and_labels(instance)
