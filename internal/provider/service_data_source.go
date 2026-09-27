package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

type serviceDataSource struct {
	client *client.Client
}

type serviceDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	EscalationPolicyID types.String `tfsdk:"escalation_policy_id"`
}

var (
	_ datasource.DataSource              = &serviceDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceDataSource{}
)

func NewServiceDataSource() datasource.DataSource {
	return &serviceDataSource{}
}

func (d *serviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *serviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert service by UUID or name search.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the GoAlert service to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the GoAlert service to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service description.",
			},
			"escalation_policy_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the assigned escalation policy.",
			},
		},
	}
}

func (d *serviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var svc *client.Service
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		s, err := d.client.ReadService(ctx, config.ID.ValueString())
		if err != nil {
			if errors.Is(err, client.ErrNotFound) {
				resp.Diagnostics.AddError("Service not found", fmt.Sprintf("Service with ID %q was not found.", config.ID.ValueString()))
				return
			}
			resp.Diagnostics.AddError("Read service failed", err.Error())
			return
		}
		svc = s
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		matches, err := d.client.SearchServices(ctx, name)
		if err != nil {
			resp.Diagnostics.AddError("Search services failed", err.Error())
			return
		}
		var exactMatches []client.Service
		for _, m := range matches {
			if m.Name == name {
				exactMatches = append(exactMatches, m)
			}
		}
		if len(exactMatches) == 0 {
			resp.Diagnostics.AddError("Service not found", fmt.Sprintf("No service with name %q was found.", name))
			return
		}
		if len(exactMatches) > 1 {
			resp.Diagnostics.AddError("Multiple services found", fmt.Sprintf("Found %d services with name %q. Use 'id' to specify a unique service.", len(exactMatches), name))
			return
		}
		svc = &exactMatches[0]
	}

	config.ID = types.StringValue(svc.ID)
	config.Name = types.StringValue(svc.Name)
	config.Description = types.StringValue(svc.Description)
	if svc.EscalationPolicy != nil {
		config.EscalationPolicyID = types.StringValue(svc.EscalationPolicy.ID)
	} else {
		config.EscalationPolicyID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
