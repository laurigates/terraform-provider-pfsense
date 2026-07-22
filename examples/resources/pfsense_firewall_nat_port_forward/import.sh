# NAT port forward rules are addressed by their position (0-based index) in
# pfSense's port forward list — the same order shown under Firewall > NAT >
# Port Forward.
terraform import pfsense_firewall_nat_port_forward.https_to_web 0
