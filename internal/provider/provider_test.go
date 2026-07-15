package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories wires the provider under test into the
// acceptance-test harness. Terraform loads it via the dev_overrides / factory,
// so no registry install is needed.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pfsense": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails fast if the box coordinates are not configured. Every
// acceptance test (TF_ACC=1) mutates a real pfSense box, so it must never run
// against production — point PFSENSE_HOST at the disposable test VM and snapshot
// before the run.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("PFSENSE_HOST") == "" {
		t.Fatal("PFSENSE_HOST must be set for acceptance tests (point it at the disposable test VM, never production)")
	}
	if os.Getenv("PFSENSE_API_KEY") == "" {
		t.Fatal("PFSENSE_API_KEY must be set for acceptance tests")
	}
}
