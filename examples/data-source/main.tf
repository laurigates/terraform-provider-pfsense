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

data "pfsense_firewall_aliases" "all" {}

output "alias_count" {
  value = length(data.pfsense_firewall_aliases.all.data)
}

output "alias_names" {
  value = [for a in data.pfsense_firewall_aliases.all.data : a.name]
}
