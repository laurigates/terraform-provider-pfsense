# Firewall rules are imported by their pfSense object id, which is the rule's
# position (index) in the config array — not the stable `tracker` value.
# Deleting a rule outside Terraform renumbers every rule after it, so re-check
# the id (GET /api/v2/firewall/rules) immediately before importing.
terraform import pfsense_firewall_rule.allow_web 3
