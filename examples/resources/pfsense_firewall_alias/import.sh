# Firewall aliases are imported by their pfSense `id`.
#
# CAVEAT: that id is the alias's *position* in the pfSense config array (0, 1,
# 2, ... in list order), not a stable identifier. Deleting an alias renumbers
# every alias after it, so an id captured earlier can later point at a different
# alias. Look the ids up immediately before importing, and re-check them if
# aliases are changed outside Terraform:
#
#   curl -sk -H "X-API-Key: $PFSENSE_API_KEY" \
#     "$PFSENSE_HOST/api/v2/firewall/aliases" | jq '.data[] | {id, name, type}'

terraform import pfsense_firewall_alias.web_ports 0
