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
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// resourceMapper supplies the per-resource specifics that envelopeResource
// needs. M is the generated Terraform model struct for the resource.
//
// Every pfSense REST API v2 object resource shares the same CRUD shape:
// POST/PATCH the endpoint with a JSON body, GET/DELETE by an `id` query
// parameter, and the response `data` is the object. The mapper isolates the
// four things that vary between resources: the type-name suffix, the generated
// schema, the endpoint path, and the model<->wire conversions (including the
// object id accessors).
type resourceMapper[M any] interface {
	// typeNameSuffix is appended to the provider type name, e.g.
	// "_firewall_alias" -> pfsense_firewall_alias.
	typeNameSuffix() string
	// schema returns the generated resource schema.
	schema(context.Context) rschema.Schema
	// apiPath is the REST endpoint, e.g. "/api/v2/firewall/alias".
	apiPath() string
	// toBody converts a plan model to the JSON request body. On update, id is
	// non-nil and the mapper must stamp it onto the body (PATCH targets by id);
	// on create, id is nil and the body omits it.
	toBody(ctx context.Context, m *M, id *int64, diags *diag.Diagnostics) any
	// fromData converts an unwrapped API `data` payload to a model.
	fromData(ctx context.Context, data json.RawMessage, diags *diag.Diagnostics) *M
	// id reads the object id from a model.
	id(*M) types.Int64
	// setID writes the object id onto a model.
	setID(*M, types.Int64)
	// applyPath is the endpoint that applies staged changes for this resource,
	// e.g. "/api/v2/firewall/apply". Return "" for resources that need no apply
	// step.
	applyPath() string
	// applyImmediately reads the apply_immediately attribute from a model.
	applyImmediately(*M) types.Bool
	// setApplyImmediately writes the apply_immediately attribute onto a model.
	// The API never echoes it, so the helper carries it from plan/prior-state
	// into the model produced from an API response.
	setApplyImmediately(*M, types.Bool)
}

// envelopeResource is a generic resource.Resource implementing the shared
// envelope-CRUD lifecycle for any pfSense object resource. The resource-specific
// behaviour is provided by mapper.
type envelopeResource[M any] struct {
	client *Client
	mapper resourceMapper[M]
}

var (
	_ resource.Resource                = (*envelopeResource[struct{}])(nil)
	_ resource.ResourceWithConfigure   = (*envelopeResource[struct{}])(nil)
	_ resource.ResourceWithImportState = (*envelopeResource[struct{}])(nil)
)

func (r *envelopeResource[M]) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.mapper.typeNameSuffix()
}

func (r *envelopeResource[M]) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.mapper.schema(ctx)
}

func (r *envelopeResource[M]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *envelopeResource[M]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := r.mapper.toBody(ctx, &plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.Do(ctx, http.MethodPost, r.mapper.apiPath(), nil, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating pfSense"+r.mapper.typeNameSuffix(), err.Error())
		return
	}

	state := r.mapper.fromData(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	applyImmediately := r.mapper.applyImmediately(&plan)
	r.mapper.setApplyImmediately(state, applyImmediately)

	if err := r.applyIfRequested(ctx, applyImmediately); err != nil {
		resp.Diagnostics.AddError("Error applying pfSense"+r.mapper.typeNameSuffix()+" change", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *envelopeResource[M]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateID := r.mapper.id(&state)
	query := url.Values{"id": []string{strconv.FormatInt(stateID.ValueInt64(), 10)}}
	data, err := r.client.Do(ctx, http.MethodGet, r.mapper.apiPath(), query, nil)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading pfSense"+r.mapper.typeNameSuffix(), err.Error())
		return
	}

	newState := r.mapper.fromData(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.mapper.id(newState).IsNull() {
		// The read-one endpoint may not echo `id`; keep the one we queried by.
		r.mapper.setID(newState, stateID)
	}
	// apply_immediately is provider-side config, not returned by the API.
	r.mapper.setApplyImmediately(newState, r.mapper.applyImmediately(&state))
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *envelopeResource[M]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateID := r.mapper.id(&state)
	idVal := stateID.ValueInt64()
	body := r.mapper.toBody(ctx, &plan, &idVal, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.Do(ctx, http.MethodPatch, r.mapper.apiPath(), nil, body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating pfSense"+r.mapper.typeNameSuffix(), err.Error())
		return
	}

	newState := r.mapper.fromData(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.mapper.id(newState).IsNull() {
		r.mapper.setID(newState, stateID)
	}
	applyImmediately := r.mapper.applyImmediately(&plan)
	r.mapper.setApplyImmediately(newState, applyImmediately)

	if err := r.applyIfRequested(ctx, applyImmediately); err != nil {
		resp.Diagnostics.AddError("Error applying pfSense"+r.mapper.typeNameSuffix()+" change", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *envelopeResource[M]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateID := r.mapper.id(&state)
	query := url.Values{"id": []string{strconv.FormatInt(stateID.ValueInt64(), 10)}}
	_, err := r.client.Do(ctx, http.MethodDelete, r.mapper.apiPath(), query, nil)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			return // already gone
		}
		resp.Diagnostics.AddError("Error deleting pfSense"+r.mapper.typeNameSuffix(), err.Error())
		return
	}

	if err := r.applyIfRequested(ctx, r.mapper.applyImmediately(&state)); err != nil {
		resp.Diagnostics.AddError("Error applying pfSense"+r.mapper.typeNameSuffix()+" deletion", err.Error())
	}
}

// applyIfRequested applies staged pfSense changes when apply_immediately is
// true (its default) and the resource declares an apply endpoint. pfSense
// stages config writes; without this the change persists in config but never
// reaches the running firewall until applied.
func (r *envelopeResource[M]) applyIfRequested(ctx context.Context, applyImmediately types.Bool) error {
	path := r.mapper.applyPath()
	if path == "" {
		return nil
	}
	// Default to applying when the value is null/unknown (matches the schema
	// default of true).
	if !applyImmediately.IsNull() && !applyImmediately.IsUnknown() && !applyImmediately.ValueBool() {
		return nil
	}
	_, err := r.client.Do(ctx, http.MethodPost, path, nil, struct{}{})
	return err
}

// ImportState imports by the pfSense object id, e.g.
// `tofu import pfsense_firewall_alias.example 3`.
func (r *envelopeResource[M]) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
