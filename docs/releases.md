# Signed Releases & Terraform Registry Publication

`terraform-provider-goalert` publishes GPG-signed, semantic-versioned releases (`v1.0.0+`) for distribution through GitHub Releases and the public HashiCorp Terraform Registry (`registry.terraform.io/providers/healdropper/goalert`).

## Release Contract

GoReleaser v2.18.2 produces cross-platform archives (`.zip`) for Linux, Windows, and macOS on `amd64` and `arm64`. Binary names and archive names follow HashiCorp Terraform provider conventions (`terraform-provider-goalert_<version>_<os>_<arch>.zip`). Each archive includes `LICENSE` and `internal/client/operations.graphql`. The Protocol v6 Registry manifest (`terraform-registry-manifest.json`) is included in `SHA256SUMS`, which receives a detached GPG signature (`SHA256SUMS.sig`).

The `.github/workflows/release.yml` workflow triggers on SemVer tags (`v*` matching `^v[0-9]+\.[0-9]+\.[0-9]+$`). It verifies that the tagged commit is an ancestor of `origin/main`, runs formatting, unit tests, `go vet`, and real disposable GoAlert v0.35.0 acceptance tests, and then executes the `package` job in the protected `release` environment to sign and publish a **draft GitHub Release**.

## Signing Identity (`release` Environment)

The GitHub `release` environment holds the dedicated RSA-4096 signing identity:

- Secret `GPG_PRIVATE_KEY`: Armored private signing key.
- Secret `GPG_PASSPHRASE`: Passphrase protecting the private signing key.
- Variable `GPG_FINGERPRINT`: Expected full public-key fingerprint (`4C110F32FCEFCBE7A0662DFE8738CFE29D8E6C12`).

The corresponding public key is registered in the HashiCorp Terraform Registry under `@healdropper` (*User Settings -> Signing Keys*). The workflow verifies the imported fingerprint against `GPG_FINGERPRINT` before running GoReleaser and verifies `gpg --batch --verify dist/*_SHA256SUMS.sig dist/*_SHA256SUMS` before uploading release assets.

## Local Verification

```console
make check-format test vet build acceptance
make release-check
make release-snapshot
```

`make release-smoke` (run on Linux/WSL or CI) additionally exercises the GoReleaser GPG signing step using an ephemeral test identity that is discarded immediately after verification.

## Publishing a New Release (`v1.x.y`)

1. Ensure the target commit is merged to `main` and all 5 `Integration` CI checks pass.
2. Tag and push the semantic version from `main`:
   ```console
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. Wait for the `Release` GitHub Actions workflow (`verify` -> `package`) to finish and produce the signed draft release.
4. Verify the draft release assets (`12` platform `.zip` archives, `terraform-registry-manifest.json`, `SHA256SUMS`, `SHA256SUMS.sig`) and publish the GitHub Release (`gh release edit v1.0.0 --draft=false`).
5. The HashiCorp Terraform Registry webhook automatically ingests the published release and renders the documentation under `docs/`.

Reference: [HashiCorp Terraform Registry — Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing).
