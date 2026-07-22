# Inbound DNAT: forward HTTPS arriving on the WAN address to an internal web
# server. `destination = "wanip"` matches the WAN interface address rather than
# its whole subnet; `local_port` must line up with `destination_port`.
resource "pfsense_firewall_nat_port_forward" "https_to_web" {
  interface        = "wan"
  protocol         = "tcp"
  source           = "any"
  destination      = "wanip"
  destination_port = "443"
  target           = "10.0.10.20"
  local_port       = "443"
  descr            = "HTTPS to internal web server"

  # Create the matching pass rule automatically instead of declaring a separate
  # pfsense_firewall_rule for this forward.
  associated_rule_id = "pass"
}
