# Read every firewall alias defined on the pfSense box.
#
# The data source takes no arguments — it returns the full alias list in
# `data`, so filtering happens in Terraform expressions.

data "pfsense_firewall_aliases" "all" {}

output "alias_names" {
  value = [for alias in data.pfsense_firewall_aliases.all.data : alias.name]
}

# Pick one alias out of the list by name and reuse its entries elsewhere.
output "web_ports" {
  value = one([
    for alias in data.pfsense_firewall_aliases.all.data :
    alias.address if alias.name == "web_ports"
  ])
}
