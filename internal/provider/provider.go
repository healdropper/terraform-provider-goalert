package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

type goalertProvider struct{ version string }
type providerModel struct {
	Endpoint  types.String `tfsdk:"endpoint"`
	APIKey    types.String `tfsdk:"api_key"`
	AllowHTTP types.Bool   `tfsdk:"allow_insecure_http"`
}

var _ provider.Provider = &goalertProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &goalertProvider{version: version} }
}
func (p *goalertProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "goalert"
	resp.Version = p.version
}
func (p *goalertProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage services on a GoAlert installation using a fixed-document GraphQL API key.",
		Attributes: map[string]schema.Attribute{
			"endpoint":            schema.StringAttribute{Optional: true, MarkdownDescription: "Full GraphQL URL, including /api/graphql. Defaults to GOALERT_ENDPOINT."},
			"api_key":             schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Admin-role key registered with the canonical operations document. Prefer GOALERT_API_KEY."},
			"allow_insecure_http": schema.BoolAttribute{Optional: true, MarkdownDescription: "Explicitly allow unencrypted HTTP beyond loopback. Defaults to false. HTTPS certificate verification is always enabled."},
		},
	}
}
func (p *goalertProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.Endpoint.IsUnknown() || config.APIKey.IsUnknown() || config.AllowHTTP.IsUnknown() {
		resp.Diagnostics.AddError("Unknown provider configuration", "Configure endpoint and API key before planning service resources.")
		return
	}
	endpoint, key := os.Getenv("GOALERT_ENDPOINT"), os.Getenv("GOALERT_API_KEY")
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if !config.APIKey.IsNull() {
		key = config.APIKey.ValueString()
	}
	c, err := client.New(endpoint, key, config.AllowHTTP.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Invalid provider configuration", err.Error())
		return
	}
	resp.ResourceData = c
}
func (p *goalertProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServiceResource,
		NewEscalationPolicyResource,
	}
}
func (p *goalertProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
