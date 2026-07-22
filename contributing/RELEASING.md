# Releasing

Releases are automated with **release-please** and published to the
Terraform/OpenTofu registry with **GoReleaser** + GPG-signed checksums.

## The automated flow

1. Conventional commits (`feat:`, `fix:`, …) land on `main`.
2. release-please maintains a release PR. Merging it publishes a GitHub Release
   and pushes a `v*` tag.
3. The `Release: publish to registry` workflow (`.github/workflows/release.yml`)
   fires on the tag: it imports the GPG key, runs GoReleaser (`.goreleaser.yml`),
   and uploads cross-compiled zips, a `*_SHA256SUMS` file, its `.sig`, and the
   `*_manifest.json` to the GitHub Release.
4. The registry ingests the tag automatically once the provider is connected
   (one-time setup below).

## One-time setup (manual — required before the first publish)

These steps create the signing key and connect the registry. They involve a
private signing key, so a human performs them.

1. **Generate a GPG signing key** (RSA 4096, no expiry is fine for a provider):
   ```
   gpg --batch --gen-key <<EOF
   %no-protection
   Key-Type: RSA
   Key-Length: 4096
   Name-Real: <your name>
   Name-Email: <your email>
   Expire-Date: 0
   %commit
   EOF
   ```
   (Or with a passphrase — then provide it as `GPG_PASSPHRASE` below.)

   > **Done** (2026-07): key generated as RSA 4096, no passphrase, fingerprint
   > `E4F6F4128E17E32AB2C8772A3F53BA4CE9A6E2A9`.
2. **Add the repo secrets** — `GPG_PRIVATE_KEY` (ASCII-armored private key,
   `gpg --armor --export-secret-keys <fingerprint>`) and `GPG_PASSPHRASE`
   (empty string if the key has no passphrase). In this org these are pushed by
   gitops rather than set by hand: the key lives in the `gpg-private-key`
   Secret Manager container and is pushed to repos flagged
   `terraform_registry = true` in `gitops/repositories.tf`. A no-passphrase
   key needs no `GPG_PASSPHRASE` secret at all — the workflow's empty-string
   default is correct.
3. **Upload the GPG _public_ key to the registry** — Terraform Registry → your
   namespace → Settings → GPG Keys → add
   `gpg --armor --export <fingerprint>`.
4. **Connect the provider** — Terraform Registry → Publish → Provider → select
   `laurigates/terraform-provider-pfsense`. The registry then watches for signed
   `v*` releases.

After this, every merged release PR publishes a new version with no manual
steps.

## Verifying a published version

```
terraform {
  required_providers {
    pfsense = {
      source  = "laurigates/pfsense"
      version = "0.1.0"
    }
  }
}
```

`terraform init` (or `tofu init`) should resolve and download the provider from
the registry.
