"""Empirical feasibility probe for Issue #89 / Gate CFG-DISC: GoAlert v0.35.0 System Config.
Probes config(all: true), setConfig partial mutations, newline-separated stringList values,
cross-field validateEnable checks, and GraphQL field error extensions against GoAlert v0.35.0.
"""
import json
from fixture import DOCUMENT, disposable

READ_OP = """
query ProviderReadSystemConfig {
  config(all: true) {
    id
    description
    value
    type
    password
    deprecated
  }
}
"""

SET_OP = """
mutation ProviderSetSystemConfig($input: [ConfigValueInput!]!) {
  setConfig(input: $input)
}
"""


def main():
    print("Starting disposable GoAlert v0.35.0 instance...", flush=True)
    with disposable() as fix:
        doc = DOCUMENT.read_text(encoding="utf-8") + "\n" + READ_OP + "\n" + SET_OP
        admin_key = fix.key("admin", document=doc)
        user_key = fix.key("user", document=doc)

        # 1. Read all config with admin API key
        all_res = fix.graphql(operation="ProviderReadSystemConfig", token=admin_key)["config"]
        pwd_keys = [item["id"] for item in all_res if item["password"]]
        types = sorted({item["type"] for item in all_res})
        print(f"Total config keys (all: true): {len(all_res)}", flush=True)
        print(f"ConfigType values: {types}", flush=True)
        print(f"Password keys ({len(pwd_keys)}): {pwd_keys}", flush=True)
        assert len(all_res) == 68, f"Expected 68 keys, got {len(all_res)}"
        assert len(pwd_keys) == 9, f"Expected 9 password keys, got {len(pwd_keys)}"

        # 2. Verify user-role key is rejected for config(all: true) and setConfig
        user_read = fix.graphql(operation="ProviderReadSystemConfig", token=user_key, raw=True)
        assert user_read.get("errors"), "Expected user-role key to be denied config(all: true)"
        user_set = fix.graphql(
            operation="ProviderSetSystemConfig",
            variables={"input": [{"id": "General.ApplicationName", "value": "Unauthorized"}]},
            token=user_key,
            raw=True,
        )
        assert user_set.get("errors"), "Expected user-role key to be denied setConfig"
        print("PASS: user-role API key rejected for both config(all: true) and setConfig.", flush=True)

        # 3. Test cross-field validation error when enabling OIDC without required fields
        invalid_oidc = fix.graphql(
            operation="ProviderSetSystemConfig",
            variables={"input": [{"id": "OIDC.Enable", "value": "true"}]},
            token=admin_key,
            raw=True,
        )
        print(
            "Cross-field validation error (OIDC.Enable=true alone):",
            json.dumps(invalid_oidc.get("errors"), indent=2),
            flush=True,
        )
        assert invalid_oidc.get("errors"), "Expected validation failure when enabling OIDC without required fields"

        # 4. Test atomic OIDC + stringList update via setConfig
        ok = fix.graphql(
            operation="ProviderSetSystemConfig",
            variables={
                "input": [
                    {"id": "OIDC.Enable", "value": "true"},
                    {"id": "OIDC.IssuerURL", "value": "https://idp.example.com"},
                    {"id": "OIDC.ClientID", "value": "goalert-oidc-client"},
                    {"id": "OIDC.ClientSecret", "value": "super-secret-oidc-value-12345"},
                    {"id": "OIDC.NewUsers", "value": "true"},
                    {"id": "Webhook.AllowedURLs", "value": "https://hooks.example.com\nhttps://alerts.example.com"},
                ]
            },
            token=admin_key,
        )["setConfig"]
        assert ok is True

        after = {item["id"]: item for item in fix.graphql(operation="ProviderReadSystemConfig", token=admin_key)["config"]}
        assert after["OIDC.Enable"]["value"] == "true"
        assert after["OIDC.IssuerURL"]["value"] == "https://idp.example.com"
        assert after["OIDC.ClientID"]["value"] == "goalert-oidc-client"
        assert after["OIDC.ClientSecret"]["value"] == "super-secret-oidc-value-12345"
        assert after["OIDC.ClientSecret"]["password"] is True
        assert after["Webhook.AllowedURLs"]["value"] == "https://hooks.example.com\nhttps://alerts.example.com"
        print("PASS: atomic OIDC + stringList setConfig and readback verified.", flush=True)

        # 5. Test single-field validation errors (invalid integer & invalid URL)
        bad_int = fix.graphql(
            operation="ProviderSetSystemConfig",
            variables={"input": [{"id": "Maintenance.AlertCleanupDays", "value": "not-a-number"}]},
            token=admin_key,
            raw=True,
        )
        print("Invalid integer error:", json.dumps(bad_int.get("errors"), indent=2), flush=True)
        assert bad_int.get("errors")

        bad_url = fix.graphql(
            operation="ProviderSetSystemConfig",
            variables={"input": [{"id": "Webhook.AllowedURLs", "value": "not-a-url"}]},
            token=admin_key,
            raw=True,
        )
        print("Invalid URL error:", json.dumps(bad_url.get("errors"), indent=2), flush=True)
        assert bad_url.get("errors")

        print("ALL SYSTEM CONFIG PROBES PASSED!", flush=True)


if __name__ == "__main__":
    main()
