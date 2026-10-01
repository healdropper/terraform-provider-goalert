# Security Policy

## Supported Versions

| Version | Supported | GoAlert Compatibility |
| :--- | :--- | :--- |
| `1.0.x` | Yes | GoAlert `v0.35.0` (and `v0.34.x` for core resources) |
| `< 1.0.0` | No | Pre-GA development candidates |

## Reporting a Vulnerability

Please **do not** report security vulnerabilities through public GitHub issues, discussions, or pull requests.

Instead, report vulnerabilities privately through **[GitHub Private Vulnerability Reporting](https://github.com/healdropper/terraform-provider-goalert/security/advisories/new)**.

When reporting a potential vulnerability, please include:
- Affected provider version(s) and resource/data-source names.
- Steps or minimal Terraform configuration required to reproduce the issue (using placeholder/synthetic values only—never include live API keys or private endpoints).
- Potential impact (e.g., sensitive attribute exposure in logs/state, transport security bypass, or GraphQL operation scope escalation).

We will acknowledge receipt of your report and coordinate a fix and advisory disclosure before public release.
