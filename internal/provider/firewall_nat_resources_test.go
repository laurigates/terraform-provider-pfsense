package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_nat_outbound_mapping"
	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_nat_port_forward"
)

// TestNATPortForwardRoundTrip verifies a create body omits id + read-only keys,
// and a response payload round-trips into the model.
func TestNATPortForwardRoundTrip(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	m := &resource_firewall_nat_port_forward.FirewallNatPortForwardModel{
		Interface:       types.StringValue("wan"),
		Protocol:        types.StringValue("tcp"),
		Destination:     types.StringValue("wanip"),
		DestinationPort: types.StringValue("443"),
		Target:          types.StringValue("10.0.0.5"),
		LocalPort:       types.StringValue("443"),
		Descr:           types.StringValue("https in"),
	}
	body := natPortForwardMapper{}.toBody(ctx, m, nil, &diags).(*natPortForwardAPI)
	if body.ID != nil {
		t.Errorf("create body must omit id")
	}
	raw, _ := json.Marshal(body)
	for _, k := range []string{"created_time", "updated_by", "id"} {
		if strings.Contains(string(raw), `"`+k+`":`) {
			t.Errorf("create body unexpectedly contains %q: %s", k, raw)
		}
	}

	data := json.RawMessage(`{"id":4,"interface":"wan","protocol":"tcp","target":"10.0.0.5","local_port":"443","created_by":"admin","associated_rule_id":"nat_abc"}`)
	got := natPortForwardMapper{}.fromData(ctx, data, &diags)
	if diags.HasError() {
		t.Fatalf("fromData diags: %v", diags)
	}
	if got.Id.ValueInt64() != 4 {
		t.Errorf("id = %d, want 4", got.Id.ValueInt64())
	}
	if got.CreatedBy.ValueString() != "admin" {
		t.Errorf("created_by = %q", got.CreatedBy.ValueString())
	}
	if got.AssociatedRuleId.ValueString() != "nat_abc" {
		t.Errorf("associated_rule_id = %q", got.AssociatedRuleId.ValueString())
	}
}

// TestNATOutboundInt64Field verifies the integer target_subnet field round-trips.
func TestNATOutboundInt64Field(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	m := &resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel{
		Interface:    types.StringValue("wan"),
		Source:       types.StringValue("10.0.0.0/24"),
		Target:       types.StringValue("wanip"),
		TargetSubnet: types.Int64Value(32),
	}
	id := int64(2)
	body := natOutboundMapper{}.toBody(ctx, m, &id, &diags).(*natOutboundAPI)
	if body.ID == nil || *body.ID != 2 {
		t.Errorf("update body id = %v, want 2", body.ID)
	}
	if body.TargetSubnet == nil || *body.TargetSubnet != 32 {
		t.Errorf("target_subnet = %v, want 32", body.TargetSubnet)
	}

	data := json.RawMessage(`{"id":2,"interface":"wan","target_subnet":24}`)
	got := natOutboundMapper{}.fromData(ctx, data, &diags)
	if got.TargetSubnet.ValueInt64() != 24 {
		t.Errorf("target_subnet = %d, want 24", got.TargetSubnet.ValueInt64())
	}
}
