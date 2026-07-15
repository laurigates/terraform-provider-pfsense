package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_nat_one_to_one_mapping"
	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_nat_outbound_mapping"
	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_nat_port_forward"
)

// The three NAT resources share the alias/rule pattern: positional `id`
// addressing, /api/v2/firewall/apply for staging, all-scalar fields. Each is a
// thin mapper over the shared envelope-CRUD helper.

// -----------------------------------------------------------------------------
// firewall_nat_port_forward
// -----------------------------------------------------------------------------

func NewFirewallNATPortForwardResource() resource.Resource {
	return &envelopeResource[resource_firewall_nat_port_forward.FirewallNatPortForwardModel]{
		mapper: natPortForwardMapper{},
	}
}

type natPortForwardMapper struct{}

var _ resourceMapper[resource_firewall_nat_port_forward.FirewallNatPortForwardModel] = natPortForwardMapper{}

type natPortForwardAPI struct {
	ID               *int64  `json:"id,omitempty"`
	Interface        *string `json:"interface,omitempty"`
	Ipprotocol       *string `json:"ipprotocol,omitempty"`
	Protocol         *string `json:"protocol,omitempty"`
	Source           *string `json:"source,omitempty"`
	SourcePort       *string `json:"source_port,omitempty"`
	Destination      *string `json:"destination,omitempty"`
	DestinationPort  *string `json:"destination_port,omitempty"`
	Target           *string `json:"target,omitempty"`
	LocalPort        *string `json:"local_port,omitempty"`
	Disabled         *bool   `json:"disabled,omitempty"`
	Nordr            *bool   `json:"nordr,omitempty"`
	Nosync           *bool   `json:"nosync,omitempty"`
	Descr            *string `json:"descr,omitempty"`
	Natreflection    *string `json:"natreflection,omitempty"`
	AssociatedRuleId *string `json:"associated_rule_id,omitempty"`
	// Read-only.
	CreatedTime *int64  `json:"created_time,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	UpdatedTime *int64  `json:"updated_time,omitempty"`
	UpdatedBy   *string `json:"updated_by,omitempty"`
}

func (natPortForwardMapper) typeNameSuffix() string { return "_firewall_nat_port_forward" }
func (natPortForwardMapper) apiPath() string        { return "/api/v2/firewall/nat/port_forward" }
func (natPortForwardMapper) applyPath() string      { return "/api/v2/firewall/apply" }
func (natPortForwardMapper) schema(ctx context.Context) rschema.Schema {
	return resource_firewall_nat_port_forward.FirewallNatPortForwardResourceSchema(ctx)
}
func (natPortForwardMapper) id(m *resource_firewall_nat_port_forward.FirewallNatPortForwardModel) types.Int64 {
	return m.Id
}
func (natPortForwardMapper) setID(m *resource_firewall_nat_port_forward.FirewallNatPortForwardModel, id types.Int64) {
	m.Id = id
}
func (natPortForwardMapper) applyImmediately(m *resource_firewall_nat_port_forward.FirewallNatPortForwardModel) types.Bool {
	return m.ApplyImmediately
}
func (natPortForwardMapper) setApplyImmediately(m *resource_firewall_nat_port_forward.FirewallNatPortForwardModel, v types.Bool) {
	m.ApplyImmediately = v
}

func (natPortForwardMapper) toBody(_ context.Context, m *resource_firewall_nat_port_forward.FirewallNatPortForwardModel, id *int64, _ *diag.Diagnostics) any {
	return &natPortForwardAPI{
		ID:               id,
		Interface:        strPtr(m.Interface),
		Ipprotocol:       strPtr(m.Ipprotocol),
		Protocol:         strPtr(m.Protocol),
		Source:           strPtr(m.Source),
		SourcePort:       strPtr(m.SourcePort),
		Destination:      strPtr(m.Destination),
		DestinationPort:  strPtr(m.DestinationPort),
		Target:           strPtr(m.Target),
		LocalPort:        strPtr(m.LocalPort),
		Disabled:         boolPtr(m.Disabled),
		Nordr:            boolPtr(m.Nordr),
		Nosync:           boolPtr(m.Nosync),
		Descr:            strPtr(m.Descr),
		Natreflection:    strPtr(m.Natreflection),
		AssociatedRuleId: strPtr(m.AssociatedRuleId),
	}
}

func (natPortForwardMapper) fromData(_ context.Context, data json.RawMessage, diags *diag.Diagnostics) *resource_firewall_nat_port_forward.FirewallNatPortForwardModel {
	var pf natPortForwardAPI
	if err := json.Unmarshal(data, &pf); err != nil {
		diags.AddError("Error decoding NAT port forward", err.Error())
		return nil
	}
	return &resource_firewall_nat_port_forward.FirewallNatPortForwardModel{
		Id:               int64Val(pf.ID),
		Interface:        strVal(pf.Interface),
		Ipprotocol:       strVal(pf.Ipprotocol),
		Protocol:         strVal(pf.Protocol),
		Source:           strVal(pf.Source),
		SourcePort:       strVal(pf.SourcePort),
		Destination:      strVal(pf.Destination),
		DestinationPort:  strVal(pf.DestinationPort),
		Target:           strVal(pf.Target),
		LocalPort:        strVal(pf.LocalPort),
		Disabled:         boolVal(pf.Disabled),
		Nordr:            boolVal(pf.Nordr),
		Nosync:           boolVal(pf.Nosync),
		Descr:            strVal(pf.Descr),
		Natreflection:    strVal(pf.Natreflection),
		AssociatedRuleId: strVal(pf.AssociatedRuleId),
		CreatedTime:      int64Val(pf.CreatedTime),
		CreatedBy:        strVal(pf.CreatedBy),
		UpdatedTime:      int64Val(pf.UpdatedTime),
		UpdatedBy:        strVal(pf.UpdatedBy),
	}
}

// -----------------------------------------------------------------------------
// firewall_nat_outbound_mapping
// -----------------------------------------------------------------------------

func NewFirewallNATOutboundMappingResource() resource.Resource {
	return &envelopeResource[resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel]{
		mapper: natOutboundMapper{},
	}
}

type natOutboundMapper struct{}

var _ resourceMapper[resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel] = natOutboundMapper{}

type natOutboundAPI struct {
	ID              *int64  `json:"id,omitempty"`
	Interface       *string `json:"interface,omitempty"`
	Protocol        *string `json:"protocol,omitempty"`
	Disabled        *bool   `json:"disabled,omitempty"`
	Nonat           *bool   `json:"nonat,omitempty"`
	Nosync          *bool   `json:"nosync,omitempty"`
	Source          *string `json:"source,omitempty"`
	SourcePort      *string `json:"source_port,omitempty"`
	Destination     *string `json:"destination,omitempty"`
	DestinationPort *string `json:"destination_port,omitempty"`
	Target          *string `json:"target,omitempty"`
	TargetSubnet    *int64  `json:"target_subnet,omitempty"`
	NatPort         *string `json:"nat_port,omitempty"`
	StaticNatPort   *bool   `json:"static_nat_port,omitempty"`
	Poolopts        *string `json:"poolopts,omitempty"`
	SourceHashKey   *string `json:"source_hash_key,omitempty"`
	Descr           *string `json:"descr,omitempty"`
}

func (natOutboundMapper) typeNameSuffix() string { return "_firewall_nat_outbound_mapping" }
func (natOutboundMapper) apiPath() string        { return "/api/v2/firewall/nat/outbound/mapping" }
func (natOutboundMapper) applyPath() string      { return "/api/v2/firewall/apply" }
func (natOutboundMapper) schema(ctx context.Context) rschema.Schema {
	return resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingResourceSchema(ctx)
}
func (natOutboundMapper) id(m *resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel) types.Int64 {
	return m.Id
}
func (natOutboundMapper) setID(m *resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel, id types.Int64) {
	m.Id = id
}
func (natOutboundMapper) applyImmediately(m *resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel) types.Bool {
	return m.ApplyImmediately
}
func (natOutboundMapper) setApplyImmediately(m *resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel, v types.Bool) {
	m.ApplyImmediately = v
}

func (natOutboundMapper) toBody(_ context.Context, m *resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel, id *int64, _ *diag.Diagnostics) any {
	return &natOutboundAPI{
		ID:              id,
		Interface:       strPtr(m.Interface),
		Protocol:        strPtr(m.Protocol),
		Disabled:        boolPtr(m.Disabled),
		Nonat:           boolPtr(m.Nonat),
		Nosync:          boolPtr(m.Nosync),
		Source:          strPtr(m.Source),
		SourcePort:      strPtr(m.SourcePort),
		Destination:     strPtr(m.Destination),
		DestinationPort: strPtr(m.DestinationPort),
		Target:          strPtr(m.Target),
		TargetSubnet:    int64Ptr(m.TargetSubnet),
		NatPort:         strPtr(m.NatPort),
		StaticNatPort:   boolPtr(m.StaticNatPort),
		Poolopts:        strPtr(m.Poolopts),
		SourceHashKey:   strPtr(m.SourceHashKey),
		Descr:           strPtr(m.Descr),
	}
}

func (natOutboundMapper) fromData(_ context.Context, data json.RawMessage, diags *diag.Diagnostics) *resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel {
	var o natOutboundAPI
	if err := json.Unmarshal(data, &o); err != nil {
		diags.AddError("Error decoding NAT outbound mapping", err.Error())
		return nil
	}
	return &resource_firewall_nat_outbound_mapping.FirewallNatOutboundMappingModel{
		Id:              int64Val(o.ID),
		Interface:       strVal(o.Interface),
		Protocol:        strVal(o.Protocol),
		Disabled:        boolVal(o.Disabled),
		Nonat:           boolVal(o.Nonat),
		Nosync:          boolVal(o.Nosync),
		Source:          strVal(o.Source),
		SourcePort:      strVal(o.SourcePort),
		Destination:     strVal(o.Destination),
		DestinationPort: strVal(o.DestinationPort),
		Target:          strVal(o.Target),
		TargetSubnet:    int64Val(o.TargetSubnet),
		NatPort:         strVal(o.NatPort),
		StaticNatPort:   boolVal(o.StaticNatPort),
		Poolopts:        strVal(o.Poolopts),
		SourceHashKey:   strVal(o.SourceHashKey),
		Descr:           strVal(o.Descr),
	}
}

// -----------------------------------------------------------------------------
// firewall_nat_one_to_one_mapping
// -----------------------------------------------------------------------------

func NewFirewallNATOneToOneMappingResource() resource.Resource {
	return &envelopeResource[resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel]{
		mapper: natOneToOneMapper{},
	}
}

type natOneToOneMapper struct{}

var _ resourceMapper[resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel] = natOneToOneMapper{}

type natOneToOneAPI struct {
	ID            *int64  `json:"id,omitempty"`
	Interface     *string `json:"interface,omitempty"`
	Disabled      *bool   `json:"disabled,omitempty"`
	Nobinat       *bool   `json:"nobinat,omitempty"`
	Natreflection *string `json:"natreflection,omitempty"`
	Ipprotocol    *string `json:"ipprotocol,omitempty"`
	External      *string `json:"external,omitempty"`
	Source        *string `json:"source,omitempty"`
	Destination   *string `json:"destination,omitempty"`
	Descr         *string `json:"descr,omitempty"`
}

func (natOneToOneMapper) typeNameSuffix() string { return "_firewall_nat_one_to_one_mapping" }
func (natOneToOneMapper) apiPath() string        { return "/api/v2/firewall/nat/one_to_one/mapping" }
func (natOneToOneMapper) applyPath() string      { return "/api/v2/firewall/apply" }
func (natOneToOneMapper) schema(ctx context.Context) rschema.Schema {
	return resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingResourceSchema(ctx)
}
func (natOneToOneMapper) id(m *resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel) types.Int64 {
	return m.Id
}
func (natOneToOneMapper) setID(m *resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel, id types.Int64) {
	m.Id = id
}
func (natOneToOneMapper) applyImmediately(m *resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel) types.Bool {
	return m.ApplyImmediately
}
func (natOneToOneMapper) setApplyImmediately(m *resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel, v types.Bool) {
	m.ApplyImmediately = v
}

func (natOneToOneMapper) toBody(_ context.Context, m *resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel, id *int64, _ *diag.Diagnostics) any {
	return &natOneToOneAPI{
		ID:            id,
		Interface:     strPtr(m.Interface),
		Disabled:      boolPtr(m.Disabled),
		Nobinat:       boolPtr(m.Nobinat),
		Natreflection: strPtr(m.Natreflection),
		Ipprotocol:    strPtr(m.Ipprotocol),
		External:      strPtr(m.External),
		Source:        strPtr(m.Source),
		Destination:   strPtr(m.Destination),
		Descr:         strPtr(m.Descr),
	}
}

func (natOneToOneMapper) fromData(_ context.Context, data json.RawMessage, diags *diag.Diagnostics) *resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel {
	var o natOneToOneAPI
	if err := json.Unmarshal(data, &o); err != nil {
		diags.AddError("Error decoding NAT one-to-one mapping", err.Error())
		return nil
	}
	return &resource_firewall_nat_one_to_one_mapping.FirewallNatOneToOneMappingModel{
		Id:            int64Val(o.ID),
		Interface:     strVal(o.Interface),
		Disabled:      boolVal(o.Disabled),
		Nobinat:       boolVal(o.Nobinat),
		Natreflection: strVal(o.Natreflection),
		Ipprotocol:    strVal(o.Ipprotocol),
		External:      strVal(o.External),
		Source:        strVal(o.Source),
		Destination:   strVal(o.Destination),
		Descr:         strVal(o.Descr),
	}
}
