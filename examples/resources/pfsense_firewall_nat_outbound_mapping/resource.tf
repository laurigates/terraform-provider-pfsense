# Outbound SNAT: rewrite the source address of traffic leaving WAN from the LAN
# subnet to the WAN interface address. Requires outbound NAT mode to be set to
# hybrid or manual on the firewall. Omitting `protocol` matches any protocol.
resource "pfsense_firewall_nat_outbound_mapping" "lan_to_wan" {
  interface   = "wan"
  source      = "10.0.10.0/24"
  destination = "any"
  target      = "wanip"
  descr       = "SNAT LAN clients behind the WAN address"

  # Stage the change without reloading the ruleset. A later resource that does
  # apply immediately (or a manual apply) commits the whole batch at once.
  apply_immediately = false
}
