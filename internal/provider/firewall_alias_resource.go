package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_alias"
)

// NewFirewallAliasResource returns the pfsense_firewall_alias resource.
func NewFirewallAliasResource() resource.Resource {
	return &envelopeResource[resource_firewall_alias.FirewallAliasModel]{
		mapper: firewallAliasMapper{},
	}
}

// firewallAliasMapper wires the generated firewall alias model into the shared
// envelope-CRUD helper.
type firewallAliasMapper struct{}

var _ resourceMapper[resource_firewall_alias.FirewallAliasModel] = firewallAliasMapper{}

// firewallAliasAPI is the wire representation of a firewall alias.
//
// The `id` field is not part of the FirewallAlias component schema (it is
// modelled as a query/body parameter in the OpenAPI spec) but the API does
// include it in response `data` payloads. It is a pointer so create bodies
// omit it.
type firewallAliasAPI struct {
	ID      *int64   `json:"id,omitempty"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Descr   string   `json:"descr"`
	Address []string `json:"address"`
	Detail  []string `json:"detail"`
}

func (firewallAliasMapper) typeNameSuffix() string { return "_firewall_alias" }

func (firewallAliasMapper) schema(ctx context.Context) rschema.Schema {
	return resource_firewall_alias.FirewallAliasResourceSchema(ctx)
}

func (firewallAliasMapper) apiPath() string { return "/api/v2/firewall/alias" }

func (firewallAliasMapper) id(m *resource_firewall_alias.FirewallAliasModel) types.Int64 {
	return m.Id
}

func (firewallAliasMapper) setID(m *resource_firewall_alias.FirewallAliasModel, id types.Int64) {
	m.Id = id
}

func (firewallAliasMapper) applyPath() string { return "/api/v2/firewall/apply" }

func (firewallAliasMapper) applyImmediately(m *resource_firewall_alias.FirewallAliasModel) types.Bool {
	return m.ApplyImmediately
}

func (firewallAliasMapper) setApplyImmediately(m *resource_firewall_alias.FirewallAliasModel, v types.Bool) {
	m.ApplyImmediately = v
}

func (firewallAliasMapper) toBody(ctx context.Context, m *resource_firewall_alias.FirewallAliasModel, id *int64, diags *diag.Diagnostics) any {
	body := &firewallAliasAPI{
		ID:      id,
		Name:    m.Name.ValueString(),
		Type:    m.Type.ValueString(),
		Descr:   m.Descr.ValueString(),
		Address: []string{},
		Detail:  []string{},
	}
	if !m.Address.IsNull() && !m.Address.IsUnknown() {
		diags.Append(m.Address.ElementsAs(ctx, &body.Address, false)...)
	}
	if !m.Detail.IsNull() && !m.Detail.IsUnknown() {
		diags.Append(m.Detail.ElementsAs(ctx, &body.Detail, false)...)
	}
	return body
}

func (firewallAliasMapper) fromData(ctx context.Context, data json.RawMessage, diags *diag.Diagnostics) *resource_firewall_alias.FirewallAliasModel {
	var alias firewallAliasAPI
	if err := json.Unmarshal(data, &alias); err != nil {
		diags.AddError("Error decoding firewall alias", err.Error())
		return nil
	}

	model := &resource_firewall_alias.FirewallAliasModel{
		Name:  types.StringValue(alias.Name),
		Type:  types.StringValue(alias.Type),
		Descr: types.StringValue(alias.Descr),
		Id:    types.Int64Null(),
	}
	if alias.ID != nil {
		model.Id = types.Int64Value(*alias.ID)
	}

	address, d := types.ListValueFrom(ctx, types.StringType, alias.Address)
	diags.Append(d...)
	model.Address = address

	detail, d := types.ListValueFrom(ctx, types.StringType, alias.Detail)
	diags.Append(d...)
	model.Detail = detail

	return model
}
