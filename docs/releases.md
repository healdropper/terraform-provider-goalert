# Signed release preparation

The first review milestone is **v0.0.1 candidate**. Do not infer a published
release from source code, a snapshot package or a successful local test.

## Release contract

GoReleaser v2.18.2 produces ZIP files for Linux, Windows and macOS on amd64 and
arm64. Binary names and archive names follow Terraform's provider conventions.
Each package includes the license and canonical GraphQL operations. The
protocol-6 Registry manifest is included in the SHA256 checksums, which receive
a detached GPG signature.

The Release workflow triggers only on `v0.0.x` tags. It checks that the tagged
commit is reachable from main, reruns formatting, unit tests, vet and real
acceptance tests, then signs packages and creates a **draft GitHub release**.
It does not publish to the Terraform Registry or modify any consumer.

Set up the GitHub `release` environment with the maintained signing identity:

- Secret `GPG_PRIVATE_KEY`: armored private signing key.
- Secret `GPG_PASSPHRASE`: its passphrase.
- Variable `GPG_FINGERPRINT`: the expected full public-key fingerprint.

No signing identity is generated or adopted implicitly. Store the recovery copy
outside GitHub, publish the public key/fingerprint through a reviewed channel,
and document expiration and rotation. The workflow checks the imported
fingerprint and fails closed when material is missing. It only grants
`contents: write` to the packaging job. Third-party actions are pinned to commits.

## Local verification

```console
make check-format test vet build acceptance
make release-check
make release-snapshot
```

These require GoReleaser in PATH (or pass `GORELEASER=/absolute/path/goreleaser`).
Snapshots skip publishing and signing; they are build verification artifacts,
not trusted distribution releases. `make release-smoke` additionally verifies
the configured GPG signing step with an ephemeral local test identity and never
publishes or uploads that private key. Run this target on Linux (including CI);
the Git-bundled GPG agent may not support native Windows IPC.

After a reviewed merge, tag the reviewed commit as `v0.0.1` and push that tag.
Inspect the private draft's packages, manifest, checksum file and signature.
Verify with the independently trusted public key:

```console
gpg --verify terraform-provider-goalert_0.0.1_SHA256SUMS.sig terraform-provider-goalert_0.0.1_SHA256SUMS
sha256sum --check terraform-provider-goalert_0.0.1_SHA256SUMS
```

Then explicitly publish the private draft when ready. A release in a private
repository is not a public Registry publication.

## Three independent decisions

1. **GitHub visibility:** review generic source, licensing, history and absence of
   environment data before making the repository public.
2. **Terraform Registry:** review signed, non-prerelease semantic-version assets,
   protocol manifest, documentation and trusted signing key before registering
   the public provider. Registry publication requires its own approval.
3. **Production adoption:** confirm required resources, compatibility, auth/key
   rotation, acceptance evidence and rollback; then propose an exact published
   version and dependency locks in a separate consumer PR. Remove all
   development overrides from that execution environment.

Do not pin production to v0.0.1 merely because service tests pass. The first
milestone requires an existing escalation policy; useful routing may require
future resources for policy steps and integration keys. Propose the production
version only when that scope and a signed distributable release are verified.

Reference: [HashiCorp publishing requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing).
