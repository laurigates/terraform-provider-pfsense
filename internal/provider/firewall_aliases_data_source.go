package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/datasource_firewall_aliases"
)

const firewallAliasesPath = "/api/v2/firewall/aliases"

var (
	_ datasource.DataSource              = (*firewallAliasesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*firewallAliasesDataSource)(nil)
)

// NewFirewallAliasesDataSource returns the pfsense_firewall_aliases data source.
func NewFirewallAliasesDataSource() datasource.DataSource {
	return &firewallAliasesDataSource{}
}

type firewallAliasesDataSource struct {
	client *Client
}

func (d *firewallAliasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_aliases"
}

func (d *firewallAliasesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_firewall_aliases.FirewallAliasesDataSourceSchema(ctx)
}

func (d *firewallAliasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data",
			fmt.Sprintf("expected *Client, got %T", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *firewallAliasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datasource_firewall_aliases.FirewallAliasesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := d.client.Do(ctx, http.MethodGet, firewallAliasesPath, nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing firewall aliases", err.Error())
		return
	}

	var aliases []firewallAliasAPI
	if err := json.Unmarshal(data, &aliases); err != nil {
		resp.Diagnostics.AddError("Error decoding firewall aliases", err.Error())
		return
	}

	attrTypes := datasource_firewall_aliases.DataValue{}.AttributeTypes(ctx)
	values := make([]attr.Value, 0, len(aliases))
	for _, alias := range aliases {
		address, diags := types.ListValueFrom(ctx, types.StringType, alias.Address)
		resp.Diagnostics.Append(diags...)
		detail, diags := types.ListValueFrom(ctx, types.StringType, alias.Detail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		value, diags := datasource_firewall_aliases.NewDataValue(attrTypes, map[string]attr.Value{
			"address": address,
			"descr":   types.StringValue(alias.Descr),
			"detail":  detail,
			"name":    types.StringValue(alias.Name),
			"type":    types.StringValue(alias.Type),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values = append(values, value)
	}

	elementType := datasource_firewall_aliases.DataType{
		ObjectType: types.ObjectType{AttrTypes: attrTypes},
	}
	list, diags := types.ListValue(elementType, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Data = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
