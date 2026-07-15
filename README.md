# terraform-provider-pfsense

A Terraform/OpenTofu provider for pfSense, built on the
[jaredhendrickson13 pfSense REST API v2](https://pfrest.org/)
(`pfSense-pkg-RESTAPI`). Schema and models are **code-generated from the
API's OpenAPI spec**; only the CRUD glue is hand-written.

Spec: `docs/prps/pfsense-terraform-provider.md` in the
[gitops](https://github.com/laurigates/gitops) repo.

## Status — M0 spike (complete)

M0 goal: pull spec → generate schema/model for `firewall_alias` only →
hand-write CRUD → import the 8 live aliases → confirm a clean plan.

**Result: pipeline proven end-to-end.**

- `go build ./...`, `go vet ./...`, `go test ./...` all pass (client
  envelope-unwrap unit tests use `httptest`, no live box needed).
- Live data-source smoke: `data.pfsense_firewall_aliases.all` read 8 aliases.
- Live import plan (`examples/import-aliases`, plan only — never applied):

  ```
  Plan: 8 to import, 0 to add, 0 to change, 0 to destroy.
  ```

Surface: resource `pfsense_firewall_alias` (full CRUD + `ImportState` by
`id`), data source `pfsense_firewall_aliases`.

> **Safety note:** the live box is production. All M0 verification was
> read-only (GETs + `tofu plan`, which only reads). The create/update/delete
> code paths exist but have never been exercised against the box; they need a
> disposable pfSense VM (PRP open question 6, M1 prerequisite).

## Resolved PRP open questions

| # | Question | Resolution (verified on the live box, 2026-07-15) |
|---|----------|---------------------------------------------------|
| 1 | HAProxy endpoints | **Confirmed present**: `/api/v2/services/haproxy/frontend\|backend\|server\|settings…` |
| 2 | WireGuard / Unbound endpoints | **Confirmed**: `/api/v2/vpn/wireguard/*`; Unbound host override = `/api/v2/services/dns_resolver/host_override` |
| — | ACME | **Confirmed present** (`/api/v2/services/acme/*`) |
| 4 | OpenAPI version | **3.0.0**, 212 paths. No 3.1-only constructs. Two preprocessing quirks instead — see "Spec preprocessing" below |
| 5 | `id` semantics | Live ids are **sequential config-array indices 0–7** in list order. They are *positional*, not stable identifiers: deleting an alias renumbers the ones after it. Fine for one-shot import + plan (M0 verified clean); **not** safe as a long-term identity if aliases are deleted outside Terraform. Revisit keying on `name` (natural key) before M4 GitOps cutover. Also: `id` is **not** part of the `FirewallAlias` component schema (it's modelled as a query/body parameter), but the API *does* include `id` (and `_links`) in response `data` payloads |
| 6 | Acceptance-test VM | Still open — required before exercising create/update/delete (M1) |

## How to regenerate

```
python3 scripts/flatten_spec.py
tfplugingen-openapi generate --config generator_config.yml --output provider-code-spec.json spec/openapi.flattened.json
python3 scripts/patch_spec.py
tfplugingen-framework generate all --input provider-code-spec.json --output internal/provider
go build ./...
```

Install the generators with `go install
github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@latest`
and `…codegen-framework/cmd/tfplugingen-framework@latest` (both are HashiCorp
**tech preview** tools).

### Spec preprocessing (`scripts/flatten_spec.py`)

Two things in the pfSense spec that `tfplugingen-openapi` cannot digest:

1. **`allOf` composition everywhere.** Every response is
   `allOf [Success, {data: <Model>}]` and every create body is
   `allOf [<Model>, {required: […]}]`; the generator skips schemas with
   >1 `allOf` subschema ("schema composition is currently not supported").
   The script resolves the `$ref`s and merges each `allOf` into a single
   object schema (`spec/openapi.flattened.json`).
2. **`oneOf: [integer, string]` on `id` parameters** — also unsupported;
   simplified to `integer` (live ids are integers).

Also note `spec/openapi.json` is re-serialized through `json.dump`: the raw
spec as served by the box uses `\/` escapes, which are valid JSON but choke
the YAML parser inside the generator ("found unknown escape character").

`scripts/patch_spec.py` then flips the resource `id` attribute from the
generated `computed_optional` to pure `computed` (it is assigned by the API).

**Response envelope handling:** every API response is wrapped in
`{code, status, response_id, message, data: {…}}`. The generator would
surface those as schema attributes, so `generator_config.yml` `ignores` them
for the resource; unwrapping lives in the hand-written client
(`internal/provider/client.go`, `Client.Do`). For the plural data source the
`data` attribute (the alias list) is kept as the schema shape.

## Running the live plan (read-only)

```
just build
just devrc                   # generates examples/dev.tfrc for this checkout
cd examples/import-aliases   # or examples/data-source
TF_CLI_CONFIG_FILE=../dev.tfrc PFSENSE_HOST=https://<box> PFSENSE_API_KEY=<key> tofu plan
```

- `examples/dev.tfrc` is **generated and gitignored** (`just devrc`) because
  `dev_overrides` requires absolute paths. It overrides **both**
  `registry.opentofu.org/...` and `registry.terraform.io/...` — OpenTofu
  resolves unprefixed provider sources to its own registry host.
- With `dev_overrides`, **do not run `tofu init`** — plan directly.
- **Never `tofu apply` against the production box.**
- `examples/import-aliases` is a genericized template. Configs mirroring your
  real alias inventory go in `examples/local/` (gitignored) — they describe
  the live network and must never be committed.

## What M1 needs next

1. **Disposable pfSense VM** with `pfSense-pkg-RESTAPI` for acceptance tests —
   the only safe way to exercise create/update/delete (PRP open question 6).
2. **Apply-staging** — writes currently send no `apply` flag, so changes
   would be staged but not applied to the running firewall. Implement the
   PRP's `apply_immediately` attribute (default `true`) and/or a
   `pfsense_firewall_apply` resource.
3. **`firewall_rule` resource + data source** (93 live rules) plus NAT —
   extend `generator_config.yml`, regenerate, replicate the thin CRUD layer.
   Consider extracting the generic envelope-CRUD into a shared helper first.
4. **`id` stability strategy** — positional ids (see open question 5) make
   drift detection fragile for rules, which are reordered more often than
   aliases; decide on natural-key lookup or `_links`-based tracking.
5. **Repo/publishing wiring** — create the GitHub repo, adopt into
   `gitops/repositories.tf`, `release_please = true`, registry publishing.
   **Publish with fresh history**: commits before the sanitization pass
   embedded the real home-network alias inventory in
   `examples/import-aliases/main.tf` — export the tree as a new initial
   commit (or `git filter-repo` that path) rather than pushing this local
   history, and run a leak scan (`192.168.`, hostnames, alias inventory)
   over the final tree first. The committed spec is clean: `servers` is `/`
   and its only `192.168.1.0` is the API package's own generic doc text.
6. **CI** — regenerate-and-diff check (codegen drift), build/vet/test.

## Layout

| Path | Purpose |
|------|---------|
| `spec/openapi.json` | OpenAPI 3.0.0 spec as pulled from the box (`GET /api/v2/schema/openapi`), re-serialized |
| `spec/openapi.flattened.json` | Generated: allOf-flattened input for the codegen (committed for reproducibility) |
| `generator_config.yml` | Endpoint → resource/data-source mapping for `tfplugingen-openapi` |
| `provider-code-spec.json` | Generated + patched IR consumed by `tfplugingen-framework` |
| `scripts/` | Spec preprocessing / patching (see above) |
| `internal/provider/*_gen.go` dirs | Generated schema + model code (`DO NOT EDIT`) |
| `internal/provider/*.go` | Hand-written: provider config, API client, alias CRUD, aliases data source |
| `examples/` | `dev.tfrc` + data-source smoke + import-blocks configs |
