package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/laurigates/terraform-provider-pfsense/internal/provider/resource_firewall_alias"
)

const firewallAliasPath = "/api/v2/firewall/alias"

var (
	_ resource.Resource                = (*firewallAliasResource)(nil)
	_ resource.ResourceWithConfigure   = (*firewallAliasResource)(nil)
	_ resource.ResourceWithImportState = (*firewallAliasResource)(nil)
)

// NewFirewallAliasResource returns the pfsense_firewall_alias resource.
func NewFirewallAliasResource() resource.Resource {
	return &firewallAliasResource{}
}

type firewallAliasResource struct {
	client *Client
}

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

func (r *firewallAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_alias"
}

func (r *firewallAliasResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_firewall_alias.FirewallAliasResourceSchema(ctx)
}

func (r *firewallAliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = client
}

func (r *firewallAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_firewall_alias.FirewallAliasModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := modelToAPI(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE: no `apply` flag is sent, so the change is staged in pfSense but
	// not applied to the running firewall. Apply-staging (per-request
	// apply=true vs a deferred pfsense_firewall_apply resource) is M1 scope —
	// see the PRP.
	data, err := r.client.Do(ctx, http.MethodPost, firewallAliasPath, nil, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall alias", err.Error())
		return
	}

	state := apiDataToModel(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *firewallAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_firewall_alias.FirewallAliasModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{"id": []string{strconv.FormatInt(state.Id.ValueInt64(), 10)}}
	data, err := r.client.Do(ctx, http.MethodGet, firewallAliasPath, query, nil)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall alias", err.Error())
		return
	}

	newState := apiDataToModel(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if newState.Id.IsNull() {
		// The read-one endpoint may not echo `id`; keep the one we queried by.
		newState.Id = state.Id
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *firewallAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_firewall_alias.FirewallAliasModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state resource_firewall_alias.FirewallAliasModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := modelToAPI(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueInt64()
	body.ID = &id

	data, err := r.client.Do(ctx, http.MethodPatch, firewallAliasPath, nil, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating firewall alias", err.Error())
		return
	}

	newState := apiDataToModel(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if newState.Id.IsNull() {
		newState.Id = state.Id
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *firewallAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_firewall_alias.FirewallAliasModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{"id": []string{strconv.FormatInt(state.Id.ValueInt64(), 10)}}
	_, err := r.client.Do(ctx, http.MethodDelete, firewallAliasPath, query, nil)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			return // already gone
		}
		resp.Diagnostics.AddError("Error deleting firewall alias", err.Error())
	}
}

// ImportState imports by the pfSense object id, e.g.
// `tofu import pfsense_firewall_alias.example 3`.
func (r *firewallAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("expected a numeric pfSense object id, got %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// modelToAPI converts the Terraform model to the API wire form.
func modelToAPI(ctx context.Context, m *resource_firewall_alias.FirewallAliasModel, diags *diag.Diagnostics) *firewallAliasAPI {
	body := &firewallAliasAPI{
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

// apiDataToModel converts an unwrapped API `data` payload to the Terraform model.
func apiDataToModel(ctx context.Context, data json.RawMessage, diags *diag.Diagnostics) *resource_firewall_alias.FirewallAliasModel {
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
