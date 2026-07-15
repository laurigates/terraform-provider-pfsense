package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_rule"
)

// NewFirewallRuleResource returns the pfsense_firewall_rule resource.
func NewFirewallRuleResource() resource.Resource {
	return &envelopeResource[resource_firewall_rule.FirewallRuleModel]{
		mapper: firewallRuleMapper{},
	}
}

type firewallRuleMapper struct{}

var _ resourceMapper[resource_firewall_rule.FirewallRuleModel] = firewallRuleMapper{}

// firewallRuleAPI is the wire representation of a firewall rule.
//
// The pfSense REST API addresses rules by their positional `id` (the config
// array index), same as aliases. The read-only `tracker` field is pfSense's
// stable per-rule identifier, surfaced here as a computed attribute; the
// tracker-based stable-identity read path (resolve tracker -> current id via
// the /firewall/rules list endpoint) is deferred until that list endpoint's
// production hang is diagnosed on the test VM — see the acceptance-testing doc.
type firewallRuleAPI struct {
	ID              *int64   `json:"id,omitempty"`
	Type            *string  `json:"type,omitempty"`
	Interface       []string `json:"interface,omitempty"`
	Ipprotocol      *string  `json:"ipprotocol,omitempty"`
	Protocol        *string  `json:"protocol,omitempty"`
	Icmptype        []string `json:"icmptype,omitempty"`
	Source          *string  `json:"source,omitempty"`
	SourcePort      *string  `json:"source_port,omitempty"`
	Destination     *string  `json:"destination,omitempty"`
	DestinationPort *string  `json:"destination_port,omitempty"`
	Descr           *string  `json:"descr,omitempty"`
	Disabled        *bool    `json:"disabled,omitempty"`
	Log             *bool    `json:"log,omitempty"`
	Tag             *string  `json:"tag,omitempty"`
	Statetype       *string  `json:"statetype,omitempty"`
	TcpFlagsAny     *bool    `json:"tcp_flags_any,omitempty"`
	TcpFlagsOutOf   []string `json:"tcp_flags_out_of,omitempty"`
	TcpFlagsSet     []string `json:"tcp_flags_set,omitempty"`
	Gateway         *string  `json:"gateway,omitempty"`
	Sched           *string  `json:"sched,omitempty"`
	Dnpipe          *string  `json:"dnpipe,omitempty"`
	Pdnpipe         *string  `json:"pdnpipe,omitempty"`
	Defaultqueue    *string  `json:"defaultqueue,omitempty"`
	Ackqueue        *string  `json:"ackqueue,omitempty"`
	Floating        *bool    `json:"floating,omitempty"`
	Quick           *bool    `json:"quick,omitempty"`
	Direction       *string  `json:"direction,omitempty"`
	// Read-only (never sent; populated from responses).
	Tracker          *int64  `json:"tracker,omitempty"`
	AssociatedRuleId *string `json:"associated_rule_id,omitempty"`
	CreatedTime      *int64  `json:"created_time,omitempty"`
	CreatedBy        *string `json:"created_by,omitempty"`
	UpdatedTime      *int64  `json:"updated_time,omitempty"`
	UpdatedBy        *string `json:"updated_by,omitempty"`
}

func (firewallRuleMapper) typeNameSuffix() string { return "_firewall_rule" }

func (firewallRuleMapper) schema(ctx context.Context) rschema.Schema {
	return resource_firewall_rule.FirewallRuleResourceSchema(ctx)
}

func (firewallRuleMapper) apiPath() string { return "/api/v2/firewall/rule" }

func (firewallRuleMapper) applyPath() string { return "/api/v2/firewall/apply" }

func (firewallRuleMapper) id(m *resource_firewall_rule.FirewallRuleModel) types.Int64 { return m.Id }

func (firewallRuleMapper) setID(m *resource_firewall_rule.FirewallRuleModel, id types.Int64) {
	m.Id = id
}

func (firewallRuleMapper) applyImmediately(m *resource_firewall_rule.FirewallRuleModel) types.Bool {
	return m.ApplyImmediately
}

func (firewallRuleMapper) setApplyImmediately(m *resource_firewall_rule.FirewallRuleModel, v types.Bool) {
	m.ApplyImmediately = v
}

func (firewallRuleMapper) toBody(ctx context.Context, m *resource_firewall_rule.FirewallRuleModel, id *int64, diags *diag.Diagnostics) any {
	return &firewallRuleAPI{
		ID:              id,
		Type:            strPtr(m.Type),
		Interface:       listToStrings(ctx, m.Interface, diags),
		Ipprotocol:      strPtr(m.Ipprotocol),
		Protocol:        strPtr(m.Protocol),
		Icmptype:        listToStrings(ctx, m.Icmptype, diags),
		Source:          strPtr(m.Source),
		SourcePort:      strPtr(m.SourcePort),
		Destination:     strPtr(m.Destination),
		DestinationPort: strPtr(m.DestinationPort),
		Descr:           strPtr(m.Descr),
		Disabled:        boolPtr(m.Disabled),
		Log:             boolPtr(m.Log),
		Tag:             strPtr(m.Tag),
		Statetype:       strPtr(m.Statetype),
		TcpFlagsAny:     boolPtr(m.TcpFlagsAny),
		TcpFlagsOutOf:   listToStrings(ctx, m.TcpFlagsOutOf, diags),
		TcpFlagsSet:     listToStrings(ctx, m.TcpFlagsSet, diags),
		Gateway:         strPtr(m.Gateway),
		Sched:           strPtr(m.Sched),
		Dnpipe:          strPtr(m.Dnpipe),
		Pdnpipe:         strPtr(m.Pdnpipe),
		Defaultqueue:    strPtr(m.Defaultqueue),
		Ackqueue:        strPtr(m.Ackqueue),
		Floating:        boolPtr(m.Floating),
		Quick:           boolPtr(m.Quick),
		Direction:       strPtr(m.Direction),
	}
}

func (firewallRuleMapper) fromData(ctx context.Context, data json.RawMessage, diags *diag.Diagnostics) *resource_firewall_rule.FirewallRuleModel {
	var rule firewallRuleAPI
	if err := json.Unmarshal(data, &rule); err != nil {
		diags.AddError("Error decoding firewall rule", err.Error())
		return nil
	}
	return &resource_firewall_rule.FirewallRuleModel{
		Id:               int64Val(rule.ID),
		Type:             strVal(rule.Type),
		Interface:        stringsToList(ctx, rule.Interface, diags),
		Ipprotocol:       strVal(rule.Ipprotocol),
		Protocol:         strVal(rule.Protocol),
		Icmptype:         stringsToList(ctx, rule.Icmptype, diags),
		Source:           strVal(rule.Source),
		SourcePort:       strVal(rule.SourcePort),
		Destination:      strVal(rule.Destination),
		DestinationPort:  strVal(rule.DestinationPort),
		Descr:            strVal(rule.Descr),
		Disabled:         boolVal(rule.Disabled),
		Log:              boolVal(rule.Log),
		Tag:              strVal(rule.Tag),
		Statetype:        strVal(rule.Statetype),
		TcpFlagsAny:      boolVal(rule.TcpFlagsAny),
		TcpFlagsOutOf:    stringsToList(ctx, rule.TcpFlagsOutOf, diags),
		TcpFlagsSet:      stringsToList(ctx, rule.TcpFlagsSet, diags),
		Gateway:          strVal(rule.Gateway),
		Sched:            strVal(rule.Sched),
		Dnpipe:           strVal(rule.Dnpipe),
		Pdnpipe:          strVal(rule.Pdnpipe),
		Defaultqueue:     strVal(rule.Defaultqueue),
		Ackqueue:         strVal(rule.Ackqueue),
		Floating:         boolVal(rule.Floating),
		Quick:            boolVal(rule.Quick),
		Direction:        strVal(rule.Direction),
		Tracker:          int64Val(rule.Tracker),
		AssociatedRuleId: strVal(rule.AssociatedRuleId),
		CreatedTime:      int64Val(rule.CreatedTime),
		CreatedBy:        strVal(rule.CreatedBy),
		UpdatedTime:      int64Val(rule.UpdatedTime),
		UpdatedBy:        strVal(rule.UpdatedBy),
	}
}
