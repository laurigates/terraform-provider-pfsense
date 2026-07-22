# 1:1 NAT mappings are addressed by their position (0-based index) in pfSense's
# 1:1 mapping list — the same order shown under Firewall > NAT > 1:1.
terraform import pfsense_firewall_nat_one_to_one_mapping.mail_server 0
