# Outbound NAT mappings are addressed by their position (0-based index) in
# pfSense's outbound mapping list — the same order shown under
# Firewall > NAT > Outbound.
terraform import pfsense_firewall_nat_outbound_mapping.lan_to_wan 0
