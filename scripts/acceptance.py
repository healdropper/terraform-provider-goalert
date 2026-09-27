"""End-to-end tests of the actual Framework binary, Terraform CLI and GoAlert."""
import argparse
import json
import os
from pathlib import Path
import tempfile
import urllib.request
from fixture import ROOT, disposable, poc, run

V001_DOCUMENT = """query ProviderReadService($id: ID!) {
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
"""

V002_DOCUMENT = """query ProviderReadService($id: ID!) {
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

V003_DOCUMENT = V002_DOCUMENT + """query ProviderReadIntegrationKey($id: ID!) {
  integrationKey(id: $id) {
    id
    name
    type
    href
    serviceID
  }
}
mutation ProviderCreateIntegrationKey($input: CreateIntegrationKeyInput!) {
  createIntegrationKey(input: $input) {
    id
    name
    type
    href
    serviceID
  }
}
mutation ProviderDeleteIntegrationKey($id: ID!) {
  deleteAll(input: [{type: integrationKey, id: $id}])
}
"""

V004_DOCUMENT = V003_DOCUMENT + """query ProviderReadHeartbeatMonitor($id: ID!) {
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
query ProviderReadServiceLabels($id: ID!) {
  service(id: $id) {
    id
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


def service_acceptance(f, provider_dir, template=None):
    with tempfile.TemporaryDirectory(prefix="goalert-provider-acceptance-") as tmp:
        directory = Path(tmp)
        cli = directory / "development.tfrc"
        cli.write_text('provider_installation {\n  dev_overrides {\n    "registry.terraform.io/healdropper/goalert" = '
                       + json.dumps(str(provider_dir.resolve()).replace("\\", "/"))
                       + '\n  }\n  direct { exclude = ["registry.terraform.io/healdropper/goalert"] }\n}\n')
        env = dict(os.environ, TF_CLI_CONFIG_FILE=str(cli), GOALERT_ENDPOINT=f.url+"/api/graphql",
                   GOALERT_API_KEY=f.token, TF_IN_AUTOMATION="1", CHECKPOINT_DISABLE="1")
        # Do not inherit user's Terraform flags, debug logs, remote state or data directory.
        for key in list(env):
            if key.startswith("TF_CLI_ARGS") or key.startswith("TF_LOG") or key in ("TF_DATA_DIR","TF_WORKSPACE"):
                env.pop(key)
        policies = [f.policy("Acceptance primary"), f.policy("Acceptance secondary")]
        config = template.read_text() if template else """
terraform {
  required_providers {
    goalert = { source = "healdropper/goalert" }
  }
}
provider "goalert" {}
variable "service_name" { type = string }
variable "description" { type = string }
variable "policy_id" { type = string }
resource "goalert_service" "test" {
  name = var.service_name
  description = var.description
  escalation_policy_id = var.policy_id
}
"""
        (directory/"main.tf").write_text(config)
        def variables(name="Acceptance service", description="Managed by Terraform", policy=None):
            (directory/"terraform.tfvars.json").write_text(json.dumps({
                "service_name": name, "description": description, "policy_id": policy or policies[0]}))
        def tf(*args, accepted=(0,), override=None):
            result=run(["terraform", *args],cwd=directory,env=override or env,accepted=accepted)
            return result
        def plan(expected):
            result=tf("plan","-input=false","-no-color","-detailed-exitcode","-out=plan.bin",accepted=(expected,))
            value=json.loads(tf("show","-json","plan.bin").stdout)
            return value
        def clean():
            value=plan(0)
            assert all(r["change"]["actions"]==["no-op"] for r in value.get("resource_changes",[]))
            print("PASS: second plan 0 to add, 0 to change, 0 to destroy.",flush=True)
        def state():
            return json.loads(tf("show","-json").stdout)["values"]["root_module"]["resources"][0]["values"]
        variables()
        tf("validate","-no-color")
        plan(2)
        tf("apply","-input=false","-auto-approve","-no-color","plan.bin")
        first=state()
        assert first["name"]=="Acceptance service" and first["escalation_policy_id"]==policies[0]
        clean()
        variables("Acceptance renamed","",policies[1])
        update=plan(2)
        assert update["resource_changes"][0]["change"]["actions"]==["update"]
        tf("apply","-input=false","-auto-approve","-no-color","plan.bin")
        assert state()["id"]==first["id"] and state()["description"]==""
        clean()
        tf("state","rm","goalert_service.test")
        tf("import","-input=false","-no-color","goalert_service.test",first["id"])
        assert state()["id"]==first["id"]
        clean()
        print("PASS: CRUD update in place and UUID import.",flush=True)
        f.graphql(variables={"id":first["id"],"name":"External rename","description":"External drift",
                            "escalationPolicyID":policies[0]},operation="ProviderUpdateService")
        drift=plan(2)
        change=drift["resource_changes"][0]["change"]
        assert change["actions"]==["update"] and change["before"]["name"]=="External rename"
        assert change["before"]["description"]=="External drift"
        assert change["before"]["escalation_policy_id"]==policies[0]
        tf("apply","-input=false","-auto-approve","-no-color","plan.bin")
        clean()
        print("PASS: drift of all managed fields detected and repaired.",flush=True)
        # An invalid key must fail refresh without removing the object from local state.
        before=(directory/"terraform.tfstate").read_bytes()
        bad=dict(env,GOALERT_API_KEY="invalid-test-key")
        tf("plan","-input=false","-no-color",accepted=(1,),override=bad)
        assert (directory/"terraform.tfstate").read_bytes()==before
        assert state()["id"]==first["id"]
        print("PASS: authentication failure preserves resource state.",flush=True)
        f.graphql(variables={"id":first["id"]},operation="ProviderDeleteService")
        missing=plan(2)
        assert missing["resource_changes"][0]["change"]["actions"]==["create"]
        tf("apply","-input=false","-auto-approve","-no-color","plan.bin")
        recreated=state()["id"]
        assert recreated!=first["id"]
        clean()
        tf("destroy","-input=false","-auto-approve","-no-color")
        assert f.graphql(variables={"id":recreated},operation="ProviderReadService")["service"] is None
        destroyed=plan(2)
        assert destroyed["resource_changes"][0]["change"]["actions"]==["create"]
        assert "values" not in json.loads(tf("show","-json").stdout) or not json.loads(tf("show","-json").stdout).get("values")
        print("PASS: external deletion recreates; Terraform destroy deletes the remote service.",flush=True)


def policy_acceptance(f, provider_dir):
    with tempfile.TemporaryDirectory(prefix="goalert-policy-acceptance-") as tmp:
        directory = Path(tmp)
        cli = directory / "development.tfrc"
        cli.write_text('provider_installation {\n  dev_overrides {\n    "registry.terraform.io/healdropper/goalert" = '
                       + json.dumps(str(provider_dir.resolve()).replace("\\", "/"))
                       + '\n  }\n  direct { exclude = ["registry.terraform.io/healdropper/goalert"] }\n}\n')
        env = dict(os.environ, TF_CLI_CONFIG_FILE=str(cli), GOALERT_ENDPOINT=f.url+"/api/graphql",
                   GOALERT_API_KEY=f.token, TF_IN_AUTOMATION="1", CHECKPOINT_DISABLE="1")
        for key in list(env):
            if key.startswith("TF_CLI_ARGS") or key.startswith("TF_LOG") or key in ("TF_DATA_DIR","TF_WORKSPACE"):
                env.pop(key)

        config = """
terraform {
  required_providers {
    goalert = { source = "healdropper/goalert" }
  }
}
provider "goalert" {}

variable "policy_name" { type = string }
variable "description" { type = string }
variable "repeat" { type = number }
variable "steps" {
  type = list(object({
    delay_minutes = number
    webhook_urls  = list(string)
  }))
  default = []
}
variable "with_service" {
  type    = bool
  default = false
}

resource "goalert_escalation_policy" "test" {
  name        = var.policy_name
  description = var.description
  repeat      = var.repeat

  dynamic "step" {
    for_each = var.steps
    content {
      delay_minutes = step.value.delay_minutes
      dynamic "webhook_action" {
        for_each = step.value.webhook_urls
        content {
          url = webhook_action.value
        }
      }
    }
  }
}

resource "goalert_service" "svc" {
  count                = var.with_service ? 1 : 0
  name                 = "Service Attached To Policy"
  description          = "Testing referential integrity"
  escalation_policy_id = goalert_escalation_policy.test.id
}
"""
        (directory/"main.tf").write_text(config)

        def variables(name="Acceptance Policy", description="Managed by Terraform", repeat=2, steps=None, with_service=False):
            if steps is None:
                steps = [
                    {"delay_minutes": 5, "webhook_urls": ["http://fake-receiver.invalid/hook1"]},
                    {"delay_minutes": 10, "webhook_urls": ["http://fake-receiver.invalid/hook2"]},
                ]
            (directory/"terraform.tfvars.json").write_text(json.dumps({
                "policy_name": name,
                "description": description,
                "repeat": repeat,
                "steps": steps,
                "with_service": with_service,
            }))

        def tf(*args, accepted=(0,), override=None):
            return run(["terraform", *args], cwd=directory, env=override or env, accepted=accepted)

        def plan(expected):
            tf("plan", "-input=false", "-no-color", "-detailed-exitcode", "-out=plan.bin", accepted=(expected,))
            return json.loads(tf("show", "-json", "plan.bin").stdout)

        def clean():
            value = plan(0)
            assert all(r["change"]["actions"] == ["no-op"] for r in value.get("resource_changes", []))
            print("PASS: second plan 0 to add, 0 to change, 0 to destroy (policy).", flush=True)

        def get_resource_values(addr):
            data = json.loads(tf("show", "-json").stdout)
            for res in data["values"]["root_module"]["resources"]:
                if res["address"] == addr:
                    return res["values"]
            raise KeyError(f"Resource {addr} not found in state")

        # 1. Validation & initial plan/apply
        variables()
        tf("validate", "-no-color")
        plan(2)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        first = get_resource_values("goalert_escalation_policy.test")
        assert first["name"] == "Acceptance Policy"
        assert first["repeat"] == 2
        assert len(first["step"]) == 2
        assert first["step"][0]["delay_minutes"] == 5
        assert first["step"][0]["step_number"] == 0
        assert first["step"][0]["webhook_action"][0]["url"] == "http://fake-receiver.invalid/hook1"
        assert first["step"][1]["delay_minutes"] == 10
        assert first["step"][1]["step_number"] == 1
        assert first["step"][1]["webhook_action"][0]["url"] == "http://fake-receiver.invalid/hook2"
        clean()
        print("PASS: goalert_escalation_policy atomic creation with nested steps and webhooks.", flush=True)

        # 2. In-place update: modify fields, update delay, add a 3rd step, remove repeat
        updated_steps = [
            {"delay_minutes": 7, "webhook_urls": ["http://fake-receiver.invalid/hook1-updated"]},
            {"delay_minutes": 10, "webhook_urls": ["http://fake-receiver.invalid/hook2"]},
            {"delay_minutes": 15, "webhook_urls": ["http://fake-receiver.invalid/hook3"]},
        ]
        variables(name="Acceptance Policy Renamed", description="", repeat=0, steps=updated_steps)
        update_plan = plan(2)
        policy_change = [rc for rc in update_plan["resource_changes"] if rc["address"] == "goalert_escalation_policy.test"][0]
        assert policy_change["change"]["actions"] == ["update"]
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        updated_state = get_resource_values("goalert_escalation_policy.test")
        assert updated_state["id"] == first["id"]
        assert updated_state["name"] == "Acceptance Policy Renamed"
        assert updated_state["description"] == ""
        assert updated_state["repeat"] == 0
        assert len(updated_state["step"]) == 3
        assert updated_state["step"][0]["delay_minutes"] == 7
        assert updated_state["step"][2]["delay_minutes"] == 15
        clean()
        print("PASS: goalert_escalation_policy in-place update (reordering, delay, webhook action, step addition).", flush=True)

        # 3. State rm & import
        policy_id = first["id"]
        tf("state", "rm", "goalert_escalation_policy.test")
        tf("import", "-input=false", "-no-color", "goalert_escalation_policy.test", policy_id)
        imported_state = get_resource_values("goalert_escalation_policy.test")
        assert imported_state["id"] == policy_id
        assert imported_state["name"] == "Acceptance Policy Renamed"
        clean()
        print("PASS: goalert_escalation_policy import and clean second plan.", flush=True)

        # 4. Drift detection and repair
        f.graphql(
            operation="ProviderUpdateEscalationPolicy",
            variables={"input": {"id": policy_id, "name": "Drifted Policy Name", "description": "Drifted Description", "repeat": 5}},
        )
        drift_plan = plan(2)
        policy_drift = [rc for rc in drift_plan["resource_changes"] if rc["address"] == "goalert_escalation_policy.test"][0]
        assert policy_drift["change"]["actions"] == ["update"]
        assert policy_drift["change"]["before"]["name"] == "Drifted Policy Name"
        assert policy_drift["change"]["before"]["repeat"] == 5
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        repaired = get_resource_values("goalert_escalation_policy.test")
        assert repaired["name"] == "Acceptance Policy Renamed"
        assert repaired["repeat"] == 0
        clean()
        print("PASS: goalert_escalation_policy drift detected and repaired.", flush=True)

        # 5. Service attachment and referential integrity check (POL-07)
        variables(name="Acceptance Policy Renamed", description="", repeat=0, steps=updated_steps, with_service=True)
        plan(2)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        svc_state = get_resource_values("goalert_service.svc[0]")
        assert svc_state["escalation_policy_id"] == policy_id
        clean()
        print("PASS: goalert_service successfully attached to goalert_escalation_policy.", flush=True)

        # Create an external service attached to this policy to test referential integrity
        ext_svc = f.graphql(
            operation="ProviderCreateService",
            variables={"name": "External Lock Service", "description": "Locks policy", "escalationPolicyID": policy_id},
        )["createService"]
        ext_svc_id = ext_svc["id"]

        destroy_locked_res = tf("destroy", "-input=false", "-auto-approve", "-no-color", accepted=(0, 1))
        destroy_out = destroy_locked_res.stdout + destroy_locked_res.stderr
        assert "currently in use" in destroy_out or "referenced by a service" in destroy_out
        print("PASS: referential integrity enforced; destroying policy in use rejected.", flush=True)

        # Unlock policy by deleting the external service
        f.graphql(operation="ProviderDeleteService", variables={"id": ext_svc_id})
        print("PASS: external service removed; policy unlocked.", flush=True)

        # 6. Key migration verification (POL-09)
        # Attempting refresh/plan with an old v0.0.1 key must fail AST hash validation without state loss
        before_state = (directory / "terraform.tfstate").read_bytes()
        v001_key = f.key("admin", document=V001_DOCUMENT)
        bad_key_env = dict(env, GOALERT_API_KEY=v001_key)
        v001_result = tf("plan", "-input=false", "-no-color", accepted=(1,), override=bad_key_env)
        v001_out = v001_result.stdout + v001_result.stderr
        assert any(expected in v001_out for expected in [
            "wrong query for API key",
            "API key document mismatch",
            "HTTP status 422",
        ])
        assert (directory / "terraform.tfstate").read_bytes() == before_state
        print("PASS: old v0.0.1 API key rejected by GoAlert AST hash validation for policy operations.", flush=True)

        # 7. Clean teardown: destroy service and policy
        tf("destroy", "-input=false", "-auto-approve", "-no-color")
        clean_after = f.graphql(operation="ProviderReadEscalationPolicy", variables={"id": policy_id}, raw=True)
        assert clean_after.get("data", {}).get("escalationPolicy") is None
        print("PASS: clean teardown destroys service and policy on remote GoAlert.", flush=True)


def integration_key_acceptance(f, provider_dir):
    with tempfile.TemporaryDirectory(prefix="goalert-ik-acceptance-") as tmp:
        directory = Path(tmp)
        cli = directory / "development.tfrc"
        cli.write_text('provider_installation {\n  dev_overrides {\n    "registry.terraform.io/healdropper/goalert" = '
                       + json.dumps(str(provider_dir.resolve()).replace("\\", "/"))
                       + '\n  }\n  direct { exclude = ["registry.terraform.io/healdropper/goalert"] }\n}\n')
        env = dict(os.environ, TF_CLI_CONFIG_FILE=str(cli), GOALERT_ENDPOINT=f.url+"/api/graphql",
                   GOALERT_API_KEY=f.token, TF_IN_AUTOMATION="1", CHECKPOINT_DISABLE="1")
        for key in list(env):
            if key.startswith("TF_CLI_ARGS") or key.startswith("TF_LOG") or key in ("TF_DATA_DIR","TF_WORKSPACE"):
                env.pop(key)

        policy_id = f.policy("IK Acceptance Policy")
        svc = f.graphql(
            operation="ProviderCreateService",
            variables={"name": "IK Acceptance Service", "description": "Parent for keys", "escalationPolicyID": policy_id}
        )["createService"]
        service_id = svc["id"]

        config = f"""
terraform {{
  required_providers {{
    goalert = {{ source = "healdropper/goalert" }}
  }}
}}
provider "goalert" {{}}

variable "key_name" {{ type = string }}
variable "key_type" {{ type = string }}

resource "goalert_integration_key" "test" {{
  service_id = "{service_id}"
  name       = var.key_name
  type       = var.key_type
}}

output "webhook_url" {{
  value     = goalert_integration_key.test.href
  sensitive = true
}}
"""
        (directory/"main.tf").write_text(config)

        def variables(name="Grafana Key", key_type="grafana"):
            (directory/"terraform.tfvars.json").write_text(json.dumps({
                "key_name": name,
                "key_type": key_type,
            }))

        def tf(*args, accepted=(0,), override=None):
            return run(["terraform", *args], cwd=directory, env=override or env, accepted=accepted)

        def plan(expected_changes):
            tf("plan", "-input=false", "-no-color", "-out=plan.bin")
            summary = json.loads(run(["terraform", "show", "-json", "plan.bin"], cwd=directory, env=env).stdout)
            changes = [c for c in summary.get("resource_changes", []) if c.get("change", {}).get("actions") != ["no-op"]]
            assert len(changes) == expected_changes, f"expected {expected_changes} changes, got {len(changes)}"

        def clean():
            res = tf("plan", "-detailed-exitcode", "-no-color", accepted=(0,))
            assert res.returncode == 0

        def get_resource_values(address):
            state = json.loads(tf("show", "-json").stdout)
            for res in state.get("values", {}).get("root_module", {}).get("resources", []):
                if res["address"] == address:
                    return res["values"]
            raise AssertionError(f"resource {address} not found in state")

        # 1. Create integration key (INT-01, INT-02)
        variables("Grafana Ingress", "grafana")
        plan(1)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()
        key_state = get_resource_values("goalert_integration_key.test")
        key_id = key_state["id"]
        assert key_state["service_id"] == service_id
        assert key_state["name"] == "Grafana Ingress"
        assert key_state["type"] == "grafana"
        assert key_id in key_state["href"]
        print("PASS: goalert_integration_key created with sensitive href token.", flush=True)

        # 2. Ingress delivery test: post Grafana alert payload to href
        payload = {
            "receiver": "goalert",
            "status": "firing",
            "alerts": [
                {
                    "status": "firing",
                    "labels": {"alertname": "AcceptanceAlert", "severity": "critical"},
                    "annotations": {"summary": "Acceptance test triggered"},
                    "startsAt": "2026-09-26T15:00:00Z",
                    "fingerprint": "acc123456"
                }
            ],
            "title": "[FIRING:1] AcceptanceAlert",
            "state": "alerting",
        }
        import urllib.request
        req = urllib.request.Request(
            key_state["href"],
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"}
        )
        with urllib.request.urlopen(req, timeout=10) as resp:
            assert resp.status == 200
        print("PASS: Grafana payload delivered to integration key href.", flush=True)

        # 3. Replacement on attribute change (INT-03)
        variables("Grafana Ingress Renamed", "grafana")
        plan(1)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()
        new_state = get_resource_values("goalert_integration_key.test")
        new_key_id = new_state["id"]
        assert new_key_id != key_id
        assert new_state["name"] == "Grafana Ingress Renamed"
        assert new_key_id in new_state["href"]
        # Old key should be deleted in GoAlert
        old_read = f.graphql(operation="ProviderReadIntegrationKey", variables={"id": key_id}, raw=True)
        assert old_read.get("data", {}).get("integrationKey") is None
        print("PASS: goalert_integration_key replaced on attribute change (old key deleted).", flush=True)

        # 4. Drift detection and repair (INT-05)
        f.graphql(operation="ProviderDeleteIntegrationKey", variables={"id": new_key_id})
        plan(1)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()
        recreated_state = get_resource_values("goalert_integration_key.test")
        recreated_id = recreated_state["id"]
        assert recreated_id != new_key_id
        print("PASS: external deletion detected; integration key recreated.", flush=True)

        # 5. Import verification (INT-08)
        import_config = config + f"""
resource "goalert_integration_key" "imported" {{
  service_id = "{service_id}"
  name       = "{recreated_state['name']}"
  type       = "{recreated_state['type']}"
}}
"""
        (directory/"main.tf").write_text(import_config)
        tf("import", "goalert_integration_key.imported", f"{service_id}/{recreated_id}")
        imported_state = get_resource_values("goalert_integration_key.imported")
        assert imported_state["id"] == recreated_id
        assert imported_state["href"] == recreated_state["href"]
        clean()
        print("PASS: goalert_integration_key imported via compound ID <service_id>/<key_id>.", flush=True)

        # 6. Key migration verification (INT-09): old v0.0.2 API key rejected
        before_state = (directory / "terraform.tfstate").read_bytes()
        v002_key = f.key("admin", document=V002_DOCUMENT)
        bad_key_env = dict(env, GOALERT_API_KEY=v002_key)
        v002_result = tf("plan", "-input=false", "-no-color", accepted=(1,), override=bad_key_env)
        v002_out = v002_result.stdout + v002_result.stderr
        assert any(expected in v002_out for expected in [
            "wrong query for API key",
            "API key document mismatch",
            "HTTP status 422",
        ])
        assert (directory / "terraform.tfstate").read_bytes() == before_state
        print("PASS: old v0.0.2 API key rejected by GoAlert AST hash validation for integration key operations.", flush=True)

        # 7. Clean destroy
        tf("destroy", "-input=false", "-auto-approve", "-no-color")
        clean_after = f.graphql(operation="ProviderReadIntegrationKey", variables={"id": recreated_id}, raw=True)
        assert clean_after.get("data", {}).get("integrationKey") is None
        f.graphql(operation="ProviderDeleteService", variables={"id": service_id})
        f.graphql(operation="ProviderDeleteEscalationPolicy", variables={"id": policy_id})
        print("PASS: clean teardown destroys integration key and parent resources.", flush=True)


def v004_acceptance(f, provider_dir):
    with tempfile.TemporaryDirectory(prefix="goalert-v004-acceptance-") as tmp:
        directory = Path(tmp)
        cli = directory / "development.tfrc"
        cli.write_text('provider_installation {\n  dev_overrides {\n    "registry.terraform.io/healdropper/goalert" = '
                       + json.dumps(str(provider_dir.resolve()).replace("\\", "/"))
                       + '\n  }\n  direct { exclude = ["registry.terraform.io/healdropper/goalert"] }\n}\n')
        env = dict(os.environ, TF_CLI_CONFIG_FILE=str(cli), GOALERT_ENDPOINT=f.url+"/api/graphql",
                   GOALERT_API_KEY=f.token, TF_IN_AUTOMATION="1", CHECKPOINT_DISABLE="1")
        for key in list(env):
            if key.startswith("TF_CLI_ARGS") or key.startswith("TF_LOG") or key in ("TF_DATA_DIR","TF_WORKSPACE"):
                env.pop(key)

        policy_id = f.policy("v004 Acceptance Policy")
        svc = f.graphql(
            operation="ProviderCreateService",
            variables={"name": "v004 Acceptance Service", "description": "Testing heartbeats, labels, data sources", "escalationPolicyID": policy_id}
        )["createService"]
        service_id = svc["id"]

        config = f"""
terraform {{
  required_providers {{
    goalert = {{ source = "healdropper/goalert" }}
  }}
}}
provider "goalert" {{}}

variable "hb_name" {{ type = string }}
variable "hb_timeout" {{ type = number }}
variable "label_value" {{ type = string }}

resource "goalert_heartbeat_monitor" "test" {{
  service_id      = "{service_id}"
  name            = var.hb_name
  timeout_minutes = var.hb_timeout
}}

resource "goalert_service_label" "env" {{
  service_id = "{service_id}"
  key        = "example.com/environment"
  value      = var.label_value
}}

data "goalert_service" "by_name" {{
  name = "v004 Acceptance Service"
  depends_on = [goalert_service_label.env]
}}

data "goalert_service" "by_id" {{
  id = "{service_id}"
}}

data "goalert_escalation_policy" "by_name" {{
  name = "v004 Acceptance Policy"
}}

data "goalert_heartbeat_monitor" "by_id" {{
  id = goalert_heartbeat_monitor.test.id
}}

output "hb_ping_url" {{
  value     = goalert_heartbeat_monitor.test.href
  sensitive = true
}}

output "ds_service_policy_id" {{
  value = data.goalert_service.by_name.escalation_policy_id
}}

output "ds_hb_timeout" {{
  value = data.goalert_heartbeat_monitor.by_id.timeout_minutes
}}
"""
        (directory/"main.tf").write_text(config)

        def variables(hb_name="Nightly Worker", hb_timeout=15, label_val="staging"):
            (directory/"terraform.tfvars.json").write_text(json.dumps({
                "hb_name": hb_name,
                "hb_timeout": hb_timeout,
                "label_value": label_val,
            }))

        def tf(*args, accepted=(0,), override=None):
            return run(["terraform", *args], cwd=directory, env=override or env, accepted=accepted)

        def plan(expected_changes):
            tf("plan", "-input=false", "-no-color", "-out=plan.bin")
            summary = json.loads(run(["terraform", "show", "-json", "plan.bin"], cwd=directory, env=env).stdout)
            changes = [c for c in summary.get("resource_changes", []) if c.get("mode") == "managed" and c.get("change", {}).get("actions") != ["no-op"]]
            assert len(changes) == expected_changes, f"expected {expected_changes} changes, got {len(changes)}"

        def clean():
            res = tf("plan", "-detailed-exitcode", "-no-color", accepted=(0,))
            assert res.returncode == 0
            (directory/"plan.bin").unlink(missing_ok=True)

        def get_resource_values(address):
            result = tf("show", "-json")
            state = json.loads(result.stdout)
            for res in state.get("values", {}).get("root_module", {}).get("resources", []):
                if res.get("address") == address:
                    return res.get("values", {})
            return None

        # 1. Initial creation
        variables("Nightly Worker", 15, "staging")
        plan(2)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()

        hb_state = get_resource_values("goalert_heartbeat_monitor.test")
        lbl_state = get_resource_values("goalert_service_label.env")
        assert hb_state is not None, "missing heartbeat monitor state"
        assert lbl_state is not None, "missing service label state"
        hb_id = hb_state["id"]
        assert hb_state["name"] == "Nightly Worker"
        assert hb_state["timeout_minutes"] == 15
        assert "/api/v2/heartbeat/" in hb_state["href"]
        assert lbl_state["value"] == "staging"

        # 2. Ping delivery: Send HTTP POST to href
        ping_req = urllib.request.Request(hb_state["href"], data=b"", method="POST")
        with urllib.request.urlopen(ping_req, timeout=10) as resp:
            assert resp.status == 200, f"expected HTTP 200 ping, got {resp.status}"
        print("PASS: goalert_heartbeat_monitor created and ping received successfully.", flush=True)

        # 3. Data sources verified
        output_res = tf("output", "-json")
        outputs = json.loads(output_res.stdout)
        assert outputs["ds_service_policy_id"]["value"] == policy_id
        assert outputs["ds_hb_timeout"]["value"] == 15
        print("PASS: data.goalert_service, data.goalert_escalation_policy, and data.goalert_heartbeat_monitor resolved.", flush=True)

        # 4. In-place update of heartbeat monitor and service label
        variables("Nightly Worker Updated", 30, "production")
        plan(2)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()
        updated_hb = get_resource_values("goalert_heartbeat_monitor.test")
        updated_lbl = get_resource_values("goalert_service_label.env")
        assert updated_hb["id"] == hb_id, "heartbeat monitor was unexpectedly replaced instead of updated in-place"
        assert updated_hb["name"] == "Nightly Worker Updated"
        assert updated_hb["timeout_minutes"] == 30
        assert updated_lbl["value"] == "production"
        print("PASS: in-place update for heartbeat monitor and service label verified.", flush=True)

        # 5. Drift detection and repair
        f.graphql(operation="ProviderDeleteHeartbeatMonitor", variables={"id": hb_id})
        plan(1)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()
        recreated_hb = get_resource_values("goalert_heartbeat_monitor.test")
        recreated_id = recreated_hb["id"]
        assert recreated_id != hb_id
        print("PASS: external deletion detected; heartbeat monitor recreated.", flush=True)

        # 6. Import verification (UUID for heartbeat monitor, compound for label)
        import_config = config + f"""
resource "goalert_heartbeat_monitor" "imported" {{
  service_id      = "{service_id}"
  name            = "{recreated_hb['name']}"
  timeout_minutes = {recreated_hb['timeout_minutes']}
}}

resource "goalert_service_label" "imported_lbl" {{
  service_id = "{service_id}"
  key        = "example.com/environment"
  value      = "production"
}}
"""
        (directory/"main.tf").write_text(import_config)
        tf("import", "goalert_heartbeat_monitor.imported", recreated_id)
        imported_hb = get_resource_values("goalert_heartbeat_monitor.imported")
        assert imported_hb["id"] == recreated_id
        assert imported_hb["href"] == recreated_hb["href"]

        tf("import", "goalert_service_label.imported_lbl", f"{service_id}/example.com/environment")
        imported_lbl = get_resource_values("goalert_service_label.imported_lbl")
        assert imported_lbl["value"] == "production"
        clean()
        print("PASS: goalert_heartbeat_monitor and goalert_service_label imported successfully.", flush=True)

        # 7. Key migration test: old v0.0.3 key fails without state corruption
        before_state = (directory / "terraform.tfstate").read_bytes()
        v003_key = f.key("admin", document=V003_DOCUMENT)
        bad_key_env = dict(env, GOALERT_API_KEY=v003_key)
        v003_result = tf("plan", "-input=false", "-no-color", accepted=(1,), override=bad_key_env)
        v003_out = v003_result.stdout + v003_result.stderr
        assert any(expected in v003_out for expected in [
            "wrong query for API key",
            "API key document mismatch",
            "HTTP status 422",
        ])
        assert (directory / "terraform.tfstate").read_bytes() == before_state
        print("PASS: old v0.0.3 API key rejected by GoAlert AST hash validation for v0.0.4 operations.", flush=True)

        # 8. Clean destroy
        tf("destroy", "-input=false", "-auto-approve", "-no-color")
        clean_after = f.graphql(operation="ProviderReadHeartbeatMonitor", variables={"id": recreated_id}, raw=True)
        assert clean_after.get("data", {}).get("heartbeatMonitor") is None
        f.graphql(operation="ProviderDeleteService", variables={"id": service_id})
        f.graphql(operation="ProviderDeleteEscalationPolicy", variables={"id": policy_id})
        print("PASS: clean teardown destroys v0.0.4 resources on remote GoAlert.", flush=True)


def v005_acceptance(f, provider_dir):
    with tempfile.TemporaryDirectory(prefix="goalert-v005-acceptance-") as tmp:
        directory = Path(tmp)
        cli = directory / "development.tfrc"
        cli.write_text('provider_installation {\n  dev_overrides {\n    "registry.terraform.io/healdropper/goalert" = '
                       + json.dumps(str(provider_dir.resolve()).replace("\\", "/"))
                       + '\n  }\n  direct { exclude = ["registry.terraform.io/healdropper/goalert"] }\n}\n')
        env = dict(os.environ, TF_CLI_CONFIG_FILE=str(cli), GOALERT_ENDPOINT=f.url+"/api/graphql",
                   GOALERT_API_KEY=f.token, TF_IN_AUTOMATION="1", CHECKPOINT_DISABLE="1")
        for key in list(env):
            if key.startswith("TF_CLI_ARGS") or key.startswith("TF_LOG") or key in ("TF_DATA_DIR","TF_WORKSPACE"):
                env.pop(key)

        config = """
terraform {
  required_providers {
    goalert = { source = "healdropper/goalert" }
  }
}
provider "goalert" {}

variable "user_name" { type = string }
variable "user_email" { type = string }
variable "user_role" { type = string }
variable "cm_name" { type = string }
variable "cm_value" { type = string }

resource "goalert_user" "operator" {
  name     = var.user_name
  email    = var.user_email
  role     = var.user_role
  username = "acceptance-operator"
  password = "Password123!"
}

resource "goalert_user_contact_method" "webhook" {
  user_id = goalert_user.operator.id
  name    = var.cm_name
  type    = "WEBHOOK"
  value   = var.cm_value
}

resource "goalert_user_notification_rule" "immediate" {
  user_id           = goalert_user.operator.id
  contact_method_id = goalert_user_contact_method.webhook.id
  delay_minutes     = 0
}

data "goalert_user" "by_id" {
  id = goalert_user.operator.id
}

data "goalert_user" "by_name" {
  name       = var.user_name
  depends_on = [goalert_user.operator]
}

data "goalert_user" "by_email" {
  email      = var.user_email
  depends_on = [goalert_user.operator]
}

output "ds_user_role" {
  value = data.goalert_user.by_id.role
}

output "ds_user_name" {
  value = data.goalert_user.by_email.name
}
"""
        (directory/"main.tf").write_text(config)

        def variables(u_name="Alice DevOps", u_email="alice.devops@example.com", u_role="user",
                      cm_name="OnCall Webhook", cm_val="https://example.com/alerts"):
            (directory/"terraform.tfvars.json").write_text(json.dumps({
                "user_name": u_name,
                "user_email": u_email,
                "user_role": u_role,
                "cm_name": cm_name,
                "cm_value": cm_val,
            }))

        def tf(*args, accepted=(0,), override=None):
            return run(["terraform", *args], cwd=directory, env=override or env, accepted=accepted)

        def plan(expected_changes):
            tf("plan", "-input=false", "-no-color", "-out=plan.bin")
            summary = json.loads(run(["terraform", "show", "-json", "plan.bin"], cwd=directory, env=env).stdout)
            changes = [c for c in summary.get("resource_changes", []) if c.get("mode") == "managed" and c.get("change", {}).get("actions") != ["no-op"]]
            assert len(changes) == expected_changes, f"expected {expected_changes} changes, got {len(changes)}"

        def clean():
            res = tf("plan", "-detailed-exitcode", "-no-color", accepted=(0,))
            assert res.returncode == 0
            (directory/"plan.bin").unlink(missing_ok=True)

        def get_resource_values(address):
            result = tf("show", "-json")
            state = json.loads(result.stdout)
            for res in state.get("values", {}).get("root_module", {}).get("resources", []):
                if res.get("address") == address:
                    return res.get("values", {})
            return None

        # 1. Initial creation (user, contact method, notification rule: 3 resources)
        variables("Alice DevOps", "alice.devops@example.com", "user", "OnCall Webhook", "https://example.com/alerts")
        plan(3)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()

        user_state = get_resource_values("goalert_user.operator")
        cm_state = get_resource_values("goalert_user_contact_method.webhook")
        nr_state = get_resource_values("goalert_user_notification_rule.immediate")
        assert user_state is not None, "missing user state"
        assert cm_state is not None, "missing contact method state"
        assert nr_state is not None, "missing notification rule state"

        user_id = user_state["id"]
        cm_id = cm_state["id"]
        nr_id = nr_state["id"]
        assert user_state["name"] == "Alice DevOps"
        assert user_state["role"] == "user"
        assert cm_state["name"] == "OnCall Webhook"
        assert cm_state["value"] == "https://example.com/alerts"
        assert nr_state["delay_minutes"] == 0
        print("PASS: goalert_user, goalert_user_contact_method, goalert_user_notification_rule created successfully.", flush=True)

        # 2. Verify data sources
        output_res = tf("output", "-json")
        outputs = json.loads(output_res.stdout)
        assert outputs["ds_user_role"]["value"] == "user"
        assert outputs["ds_user_name"]["value"] == "Alice DevOps"
        print("PASS: data.goalert_user lookups by id, name, and email verified.", flush=True)

        # 3. In-place update of user and contact method
        variables("Alice Lead DevOps", "alice.lead@example.com", "admin", "Primary Webhook", "https://example.com/alerts")
        plan(2)
        tf("apply", "-input=false", "-auto-approve", "-no-color", "plan.bin")
        clean()

        updated_u = get_resource_values("goalert_user.operator")
        updated_cm = get_resource_values("goalert_user_contact_method.webhook")
        assert updated_u["id"] == user_id, "user was unexpectedly replaced"
        assert updated_u["name"] == "Alice Lead DevOps"
        assert updated_u["role"] == "admin"
        assert updated_cm["id"] == cm_id, "contact method was unexpectedly replaced"
        assert updated_cm["name"] == "Primary Webhook"
        assert updated_cm["value"] == "https://example.com/alerts"
        print("PASS: in-place update for user and contact method verified.", flush=True)

        # 4. Drift detection and repair
        f.graphql(operation="ProviderDeleteUserContactMethod", variables={"id": cm_id})
        # Deleting contact method also cascade-deletes the notification rule in GoAlert, so both will be recreated
        tf("apply", "-input=false", "-auto-approve", "-no-color")
        clean()
        recreated_cm = get_resource_values("goalert_user_contact_method.webhook")
        recreated_cm_id = recreated_cm["id"]
        assert recreated_cm_id != cm_id
        recreated_nr = get_resource_values("goalert_user_notification_rule.immediate")
        recreated_nr_id = recreated_nr["id"]
        assert recreated_nr_id != nr_id
        print("PASS: external contact method deletion detected; contact method and rule recreated.", flush=True)

        # 5. Import verification
        import_config = config + f"""
resource "goalert_user" "imported_user" {{
  name     = "{updated_u['name']}"
  email    = "{updated_u['email']}"
  role     = "{updated_u['role']}"
  username = "imported-user"
}}

resource "goalert_user_contact_method" "imported_cm" {{
  user_id = "{user_id}"
  name    = "{recreated_cm['name']}"
  type    = "WEBHOOK"
  value   = "{recreated_cm['value']}"
}}

resource "goalert_user_notification_rule" "imported_nr" {{
  user_id           = "{user_id}"
  contact_method_id = "{recreated_cm_id}"
  delay_minutes     = 0
}}
"""
        (directory/"main.tf").write_text(import_config)
        tf("import", "goalert_user.imported_user", f"{user_id}/imported-user")
        imported_u = get_resource_values("goalert_user.imported_user")
        assert imported_u["id"] == user_id
        assert imported_u["username"] == "imported-user"

        tf("import", "goalert_user_contact_method.imported_cm", f"{user_id}/{recreated_cm_id}")
        imported_cm = get_resource_values("goalert_user_contact_method.imported_cm")
        assert imported_cm["id"] == recreated_cm_id

        tf("import", "goalert_user_notification_rule.imported_nr", f"{user_id}/{recreated_nr_id}")
        imported_nr = get_resource_values("goalert_user_notification_rule.imported_nr")
        assert imported_nr["id"] == recreated_nr_id
        clean()
        print("PASS: user, contact method, and notification rule imported successfully.", flush=True)

        # 6. AST key migration test: v0.0.4 key rejected for v0.0.5 operations
        before_state = (directory / "terraform.tfstate").read_bytes()
        v004_key = f.key("admin", document=V004_DOCUMENT)
        bad_key_env = dict(env, GOALERT_API_KEY=v004_key)
        v004_result = tf("plan", "-input=false", "-no-color", accepted=(1,), override=bad_key_env)
        v004_out = v004_result.stdout + v004_result.stderr
        assert any(expected in v004_out for expected in [
            "wrong query for API key",
            "API key document mismatch",
            "HTTP status 422",
        ])
        assert (directory / "terraform.tfstate").read_bytes() == before_state
        print("PASS: old v0.0.4 API key rejected by GoAlert AST hash validation for v0.0.5 operations.", flush=True)

        # 7. Clean destroy
        tf("destroy", "-input=false", "-auto-approve", "-no-color")
        clean_after = f.graphql(operation="ProviderReadUser", variables={"id": user_id}, raw=True)
        assert clean_after.get("data", {}).get("user") is None
        print("PASS: clean teardown destroys v0.0.5 resources on remote GoAlert.", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--provider-dir", type=Path, default=ROOT/"bin")
    parser.add_argument("--config-template", type=Path)
    parser.add_argument("--suite", choices=["all", "poc", "service", "policy", "integration_key", "v004", "v005"], default="all")
    args = parser.parse_args()
    with disposable() as fixture:
        if args.suite in ("all", "poc"):
            poc(fixture)
        if args.suite in ("all", "service"):
            service_acceptance(fixture, args.provider_dir, args.config_template)
        if args.suite in ("all", "policy"):
            policy_acceptance(fixture, args.provider_dir)
        if args.suite in ("all", "integration_key"):
            integration_key_acceptance(fixture, args.provider_dir)
        if args.suite in ("all", "v004"):
            v004_acceptance(fixture, args.provider_dir)
        if args.suite in ("all", "v005"):
            v005_acceptance(fixture, args.provider_dir)
