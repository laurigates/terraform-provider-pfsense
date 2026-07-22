# Aliases name a group of hosts, networks or ports so firewall and NAT rules can
# refer to the group instead of repeating literal addresses. The `type` decides
# what `address` entries are allowed: "host", "network" or "port".

# Port alias — referenced by rules as the destination port group.
resource "pfsense_firewall_alias" "web_ports" {
  name    = "web_ports"
  type    = "port"
  descr   = "HTTP + HTTPS"
  address = ["80", "443"]

  # Optional per-entry labels; positionally matched to `address`.
  detail = ["http", "https"]
}

# Network alias — CIDRs (or FQDNs) grouped under one name.
resource "pfsense_firewall_alias" "internal_nets" {
  name    = "internal_nets"
  type    = "network"
  descr   = "Internal networks"
  address = ["10.0.10.0/24", "10.0.20.0/24"]
  detail  = ["LAN", "GUEST"]
}

# Host alias — individual IPs or FQDNs.
# apply_immediately = false stages the change without reloading the running
# firewall, so a batch of resources can be applied together.
resource "pfsense_firewall_alias" "dns_servers" {
  name              = "dns_servers"
  type              = "host"
  descr             = "Upstream resolvers"
  address           = ["1.1.1.1", "9.9.9.9"]
  apply_immediately = false
}
