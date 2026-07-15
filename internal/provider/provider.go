package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*pfsenseProvider)(nil)

// New returns a provider.Provider factory for the given version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pfsenseProvider{version: version}
	}
}

type pfsenseProvider struct {
	version string
}

type pfsenseProviderModel struct {
	Host     types.String `tfsdk:"host"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *pfsenseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pfsense"
	resp.Version = p.version
}

func (p *pfsenseProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a pfSense firewall via the pfSense-pkg-RESTAPI (REST API v2) package.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the pfSense box, e.g. `https://192.168.0.1`. " +
					"Falls back to the `PFSENSE_HOST` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "REST API key, sent as the `X-API-Key` header. " +
					"Falls back to the `PFSENSE_API_KEY` environment variable.",
			},
			"insecure": schema.BoolAttribute{
				Optional: true,
				Description: "Skip TLS certificate verification (self-signed box certificate). " +
					"Defaults to false.",
			},
		},
	}
}

func (p *pfsenseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config pfsenseProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("PFSENSE_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	apiKey := os.Getenv("PFSENSE_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Missing pfSense host",
			"Set the provider `host` attribute or the PFSENSE_HOST environment variable.",
		)
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing pfSense API key",
			"Set the provider `api_key` attribute or the PFSENSE_API_KEY environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client := NewClient(host, apiKey, insecure)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *pfsenseProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFirewallAliasResource,
	}
}

func (p *pfsenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFirewallAliasesDataSource,
	}
}
