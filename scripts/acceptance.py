"""End-to-end tests of the actual Framework binary, Terraform CLI and GoAlert."""
import argparse
import json
import os
from pathlib import Path
import tempfile
from fixture import ROOT, disposable, poc, run

def acceptance(f, provider_dir, template=None):
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

if __name__=="__main__":
    parser=argparse.ArgumentParser()
    parser.add_argument("--provider-dir",type=Path,default=ROOT/"bin")
    parser.add_argument("--config-template",type=Path)
    args=parser.parse_args()
    with disposable() as fixture:
        poc(fixture)
        acceptance(fixture,args.provider_dir,args.config_template)
