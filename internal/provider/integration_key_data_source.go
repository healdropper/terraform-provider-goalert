package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

type integrationKeyDataSource struct {
	client *client.Client
}

type integrationKeyDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ServiceID types.String `tfsdk:"service_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Href      types.String `tfsdk:"href"`
}

var (
	_ datasource.DataSource              = &integrationKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &integrationKeyDataSource{}
)

func NewIntegrationKeyDataSource() datasource.DataSource {
	return &integrationKeyDataSource{}
}

func (d *integrationKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_key"
}

func (d *integrationKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert integration key by UUID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the GoAlert integration key.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"service_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the parent GoAlert service.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Integration key name.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Integration key type.",
			},
			"href": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Complete incoming webhook URL containing the authentication token.",
			},
		},
	}
}

func (d *integrationKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider client", "Expected a GoAlert API client.")
		return
	}
	d.client = c
}

func (d *integrationKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config integrationKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ik, err := d.client.ReadIntegrationKey(ctx, config.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Integration key not found", fmt.Sprintf("Integration key with ID %q was not found.", config.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Read integration key failed", err.Error())
		return
	}

	config.ID = types.StringValue(ik.ID)
	config.ServiceID = types.StringValue(ik.ServiceID)
	config.Name = types.StringValue(ik.Name)
	config.Type = types.StringValue(ik.Type)
	config.Href = types.StringValue(ik.Href)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
