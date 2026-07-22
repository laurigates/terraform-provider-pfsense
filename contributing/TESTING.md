# Testing

## Unit tests (default)

```
just test        # go test ./... — httptest only, no live box
```

These cover the API client envelope handling, the apply-staging decision, and
every resource's model↔wire conversion (create-body id omission, update-body id
stamping, response round-trip). They run in CI and need no pfSense box.

## Acceptance tests (live box, opt-in)

Acceptance tests (`TestAcc*`) drive a **real** pfSense box through the full
create → update → import lifecycle. They only run when `TF_ACC=1` and **must
never point at a production firewall** — every step creates, mutates, and
deletes real config. Run them against a disposable VM, and snapshot before /
roll back after each run.

```
export TF_ACC=1
export PFSENSE_HOST=https://<test-vm-ip>
export PFSENSE_API_KEY=<test-vm-api-key>
go test ./internal/provider/ -run TestAcc -v
```

`TestAccFirewallAlias` is the template; the rule and NAT resources follow the
same shape.

## Standing up the disposable test VM

The acceptance substrate is a throwaway **pfSense CE** VM. On a TrueNAS SCALE
host (or any hypervisor):

1. **Create the VM** — amd64, 1 vCPU, 1 GB RAM, ~8 GB disk, one NIC on a
   management network reachable from where you run the tests. Attach the pfSense
   CE amd64 install ISO.
2. **Install pfSense** (interactive console) — accept defaults, install to disk,
   reboot, remove the ISO. Assign interfaces (WAN/LAN) and set a LAN IP you can
   reach.
3. **Install the REST API package** — from the pfSense shell (option 8) or
   Diagnostics → Command Prompt:
   ```
   pkg-static -C /dev/null add \
     https://github.com/jaredhendrickson13/pfsense-api/releases/latest/download/pfSense-2.7.2-pkg-RESTAPI.pkg
   /usr/local/pkg/RESTAPI/scripts/manage.php buildendpoints
   ```
4. **Mint an API key** — System → REST API → Keys → Add, or via the UI after
   enabling key auth. Use it as `PFSENSE_API_KEY`.
5. **Snapshot the VM** — take a clean snapshot named e.g. `restapi-clean`. Roll
   back to it between acceptance runs so each run starts from known state.

## Reading the production ruleset safely

The production box's **list** endpoints (`GET /firewall/rules`) hang, but the
**single-rule** endpoint works: `GET /firewall/rule?id=N` returns one rule in
~9 s. The safe read path for the whole ruleset is to iterate the singular
endpoint by positional id (`id=0,1,2,…` until a `404`), one request at a time —
never the list endpoint. This is how a read-only import of the live rules is
generated (see the import example under `examples/local/`, gitignored).
