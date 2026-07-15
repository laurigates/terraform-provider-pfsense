package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFirewallAlias exercises the full create -> update -> import lifecycle
// of pfsense_firewall_alias against a live pfSense box. It only runs under
// TF_ACC=1 and requires PFSENSE_HOST/PFSENSE_API_KEY pointing at the disposable
// test VM (snapshot before, roll back after). This is the template the rule and
// NAT resources follow.
func TestAccFirewallAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccFirewallAliasConfig("acc_ports", "port", []string{"80", "443"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pfsense_firewall_alias.test", "name", "acc_ports"),
					resource.TestCheckResourceAttr("pfsense_firewall_alias.test", "type", "port"),
					resource.TestCheckResourceAttr("pfsense_firewall_alias.test", "address.#", "2"),
					resource.TestCheckResourceAttrSet("pfsense_firewall_alias.test", "id"),
				),
			},
			{ // update (add an entry)
				Config: testAccFirewallAliasConfig("acc_ports", "port", []string{"80", "443", "8080"}),
				Check: resource.TestCheckResourceAttr("pfsense_firewall_alias.test", "address.#", "3"),
			},
			{ // import
				ResourceName:      "pfsense_firewall_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
				// apply_immediately is provider-side config not returned by the
				// API, so it is not part of imported state.
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func testAccFirewallAliasConfig(name, typ string, addrs []string) string {
	quoted := ""
	for i, a := range addrs {
		if i > 0 {
			quoted += ", "
		}
		quoted += fmt.Sprintf("%q", a)
	}
	return fmt.Sprintf(`
resource "pfsense_firewall_alias" "test" {
  name    = %q
  type    = %q
  descr   = "acceptance test"
  address = [%s]
}
`, name, typ, quoted)
}
