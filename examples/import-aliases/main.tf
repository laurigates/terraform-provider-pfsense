# Genericized import example. To adopt your box's real aliases:
#   1. List them: curl -sk -m 15 -H "X-API-Key: $PFSENSE_API_KEY" \
#        "$PFSENSE_HOST/api/v2/firewall/aliases" | jq '.data[] | {id, name, type}'
#   2. Write one import {} + resource pair per alias (id = its array index),
#      mirroring the live attributes until `tofu plan` shows
#      "N to import, 0 to add, 0 to change, 0 to destroy".
# Keep real-inventory configs in examples/local/ (gitignored) — they describe
# your network and don't belong in the repo.
terraform {
  required_providers {
    pfsense = {
      source = "laurigates/pfsense"
    }
  }
}

# host/api_key come from PFSENSE_HOST / PFSENSE_API_KEY env vars.
provider "pfsense" {
  insecure = true # self-signed box certificate
}

import {
  to = pfsense_firewall_alias.web_ports
  id = "0"
}

resource "pfsense_firewall_alias" "web_ports" {
  name    = "web_ports"
  type    = "port"
  descr   = "HTTP + HTTPS"
  address = ["80", "443"]
  detail  = ["http", "https"]
}

import {
  to = pfsense_firewall_alias.lan_nets
  id = "1"
}

resource "pfsense_firewall_alias" "lan_nets" {
  name    = "lan_nets"
  type    = "network"
  descr   = "Internal networks"
  address = ["10.0.10.0/24", "10.0.20.0/24"]
  detail  = ["LAN", "GUEST"]
}
