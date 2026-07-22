# Two filter rules on the LAN interface: allow web traffic to a port alias,
# then block a host group. Rules are evaluated in pfSense's configured order.

# A port-type alias referenced by name from the rule below. pfSense resolves
# alias names inline wherever an address or port is accepted.
resource "pfsense_firewall_alias" "web_ports" {
  name    = "web_ports"
  type    = "port"
  descr   = "HTTP/HTTPS"
  address = ["80", "443"]
}

resource "pfsense_firewall_rule" "allow_web" {
  type       = "pass"
  interface  = ["lan"]
  ipprotocol = "inet"
  protocol   = "tcp"

  source           = "lan"
  destination      = "any"
  destination_port = pfsense_firewall_alias.web_ports.name

  descr = "Allow LAN hosts to reach the web"
  log   = true
}

resource "pfsense_firewall_rule" "block_guest_to_lan" {
  type       = "block"
  interface  = ["lan"]
  ipprotocol = "inet"
  protocol   = "tcp/udp"

  # `:ip` on an interface means the interface address itself, not its subnet.
  source      = "192.168.20.0/24"
  destination = "lan:ip"

  descr = "Guest subnet may not reach the firewall's LAN address"
  log   = true

  # Staged changes are applied to the running firewall immediately by default.
  # Set false to batch several rule changes and apply them in one reload —
  # the rule exists in the pfSense config but is not enforced until an apply
  # runs (for example, a later resource in the same plan with the default).
  apply_immediately = false
}
