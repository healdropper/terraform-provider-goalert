# Private provider development

Build with `make build`. The binary is in `bin/`. Use a dedicated local
Terraform root with local state and a development CLI configuration. Do not
modify a production root, remote backend, shared CLI configuration or deployment
workflow to load a development binary.

On Windows, Terraform normally reads `%APPDATA%/terraform.rc`; on Unix it reads
`~/.terraformrc`. For this workflow, use neither global file. Create an ignored
file such as `.local/development.tfrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/healdropper/goalert" = "C:/Repos/healdropper/terraform-provider-goalert/bin"
  }
  direct {
    exclude = ["registry.terraform.io/healdropper/goalert"]
  }
}
```

In PowerShell:

```powershell
$previousCLI = $env:TF_CLI_CONFIG_FILE
try {
    $env:TF_CLI_CONFIG_FILE = (Resolve-Path .local/development.tfrc).Path
    $env:GOALERT_ENDPOINT = "http://127.0.0.1:18081/api/graphql"
    # Inject GOALERT_API_KEY from your secret store; never commit or print it.
    terraform -chdir=examples/resources/goalert_service plan
} finally {
    $env:TF_CLI_CONFIG_FILE = $previousCLI
    Remove-Item Env:GOALERT_API_KEY -ErrorAction SilentlyContinue
}
```

Supply `TF_VAR_escalation_policy_id` with an existing policy UUID for the
standalone example. Environment settings are scoped to your terminal.
For Bash use `TF_CLI_CONFIG_FILE=/absolute/path/development.tfrc terraform plan`.
Forward slashes work in Windows HCL paths.

**Skip `terraform init` in this provider-only development root.** The override
lets validate/plan/apply/import load the binary directly, but initialization still
attempts provider version discovery. The source address is intentionally excluded
from direct Registry installation. If testing other published providers,
initialize them separately or use a separate test root.

`make acceptance` automates this workflow with a temporary CLI config, local
state, a generated API key and an isolated GoAlert instance. It restores the
parent environment by passing settings only to child processes. Test containers
are named with a random Compose project, bound to loopback, and removed in a
`finally` block. Database storage is ephemeral. Port 18081 must be available;
override it with `FIXTURE_PORT` if needed. Interrupted processes may need cleanup
of their printed/observed Compose project; never run global Docker pruning.

For a consuming repository, put its experimental configuration under a local
test directory. Its resource address must be `goalert_service.test` and its
variables `service_name`, `description`, `policy_id` to reuse this harness:

```console
python scripts/acceptance.py --provider-dir /absolute/provider/bin --config-template /absolute/consumer/main.tf
```

This validates the consumer's Terraform contract against a real disposable
GoAlert, not the deployed consumer's production service or credentials.
No temporary overrides, local binary paths or unversioned references should be
promoted to production.

Reference: [HashiCorp development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).
