package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_rule"
)

// TestFirewallRuleToBodyOmitsID verifies a create body omits the id (so the API
// assigns it) and carries the writable fields, including list fields.
func TestFirewallRuleToBodyOmitsID(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	iface, d := types.ListValueFrom(ctx, types.StringType, []string{"wan"})
	diags.Append(d...)

	m := &resource_firewall_rule.FirewallRuleModel{
		Type:        types.StringValue("pass"),
		Interface:   iface,
		Ipprotocol:  types.StringValue("inet"),
		Protocol:    types.StringValue("tcp"),
		Source:      types.StringValue("any"),
		Destination: types.StringValue("any"),
		Descr:       types.StringValue("allow web"),
		Disabled:    types.BoolValue(false),
	}

	body := firewallRuleMapper{}.toBody(ctx, m, nil, &diags).(*firewallRuleAPI)
	if diags.HasError() {
		t.Fatalf("toBody diags: %v", diags)
	}
	if body.ID != nil {
		t.Errorf("create body must omit id, got %v", *body.ID)
	}
	if body.Type == nil || *body.Type != "pass" {
		t.Errorf("type = %v, want pass", body.Type)
	}
	if len(body.Interface) != 1 || body.Interface[0] != "wan" {
		t.Errorf("interface = %v, want [wan]", body.Interface)
	}

	// Marshalled body must not contain the read-only tracker/created_* keys.
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, k := range []string{"tracker", "created_time", "updated_by", "id"} {
		if strings.Contains(string(raw), `"`+k+`":`) {
			t.Errorf("create body unexpectedly contains %q: %s", k, raw)
		}
	}
}

// TestFirewallRuleToBodyUpdateSetsID verifies an update body stamps the id.
func TestFirewallRuleToBodyUpdateSetsID(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	m := &resource_firewall_rule.FirewallRuleModel{Type: types.StringValue("block")}
	id := int64(7)
	body := firewallRuleMapper{}.toBody(ctx, m, &id, &diags).(*firewallRuleAPI)
	if body.ID == nil || *body.ID != 7 {
		t.Errorf("update body id = %v, want 7", body.ID)
	}
}

// TestFirewallRuleFromData verifies a response payload round-trips into the
// model, including the read-only tracker and list fields.
func TestFirewallRuleFromData(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	data := json.RawMessage(`{
		"id": 3,
		"type": "pass",
		"interface": ["lan"],
		"ipprotocol": "inet",
		"protocol": "tcp",
		"source": "any",
		"destination": "any",
		"descr": "web",
		"disabled": false,
		"tracker": 1700000000,
		"created_by": "admin@1.2.3.4",
		"tcp_flags_set": ["syn"]
	}`)
	m := firewallRuleMapper{}.fromData(ctx, data, &diags)
	if diags.HasError() {
		t.Fatalf("fromData diags: %v", diags)
	}
	if m.Id.ValueInt64() != 3 {
		t.Errorf("id = %d, want 3", m.Id.ValueInt64())
	}
	if m.Tracker.ValueInt64() != 1700000000 {
		t.Errorf("tracker = %d, want 1700000000", m.Tracker.ValueInt64())
	}
	if m.CreatedBy.ValueString() != "admin@1.2.3.4" {
		t.Errorf("created_by = %q", m.CreatedBy.ValueString())
	}
	var flags []string
	diags.Append(m.TcpFlagsSet.ElementsAs(ctx, &flags, false)...)
	if len(flags) != 1 || flags[0] != "syn" {
		t.Errorf("tcp_flags_set = %v, want [syn]", flags)
	}
	// An absent field becomes a Terraform null, not an empty string.
	if !m.Gateway.IsNull() {
		t.Errorf("absent gateway should be null, got %v", m.Gateway)
	}
}
