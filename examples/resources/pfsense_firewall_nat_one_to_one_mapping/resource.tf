# Static 1:1 NAT: bind a spare public address on WAN to a single internal host,
# translating both directions. `external` is the public side, `source` the
# internal host it maps to.
resource "pfsense_firewall_nat_one_to_one_mapping" "mail_server" {
  interface   = "wan"
  external    = "203.0.113.10"
  source      = "10.0.10.25"
  destination = "any"
  descr       = "1:1 NAT for the mail server"

  # Let clients on internal networks reach the mapped host by its public
  # address; omit to fall back to the system default.
  natreflection = "enable"
}
