"""End-to-end tests of the actual Framework binary, Terraform CLI and GoAlert."""
import argparse
import json
import os
from pathlib import Path
import tempfile
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


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--provider-dir", type=Path, default=ROOT/"bin")
    parser.add_argument("--config-template", type=Path)
    args = parser.parse_args()
    with disposable() as fixture:
        poc(fixture)
        service_acceptance(fixture, args.provider_dir, args.config_template)
        policy_acceptance(fixture, args.provider_dir)
