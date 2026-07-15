package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Conversion helpers between Terraform framework types and the Go pointer/slice
// forms used in the JSON wire structs. Pointers let create bodies omit unset
// optional fields (json:"...,omitempty"); the *Val helpers turn an absent wire
// value back into a Terraform null.

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func strVal(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func boolVal(p *bool) types.Bool {
	if p == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*p)
}

func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

func int64Val(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}

// listToStrings reads a types.List of strings into a []string. A null/unknown
// list yields nil (the field is then omitted from the wire body).
func listToStrings(ctx context.Context, v types.List, diags *diag.Diagnostics) []string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	out := []string{}
	diags.Append(v.ElementsAs(ctx, &out, false)...)
	return out
}

// stringsToList builds a types.List of strings, mapping a nil slice to an empty
// (non-null) list so a computed list attribute is never left unknown.
func stringsToList(ctx context.Context, s []string, diags *diag.Diagnostics) types.List {
	if s == nil {
		s = []string{}
	}
	l, d := types.ListValueFrom(ctx, types.StringType, s)
	diags.Append(d...)
	return l
}
