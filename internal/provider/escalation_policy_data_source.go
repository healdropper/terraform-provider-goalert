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

type escalationPolicyDataSource struct {
	client *client.Client
}

type escalationPolicyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Repeat      types.Int64  `tfsdk:"repeat"`
}

var (
	_ datasource.DataSource              = &escalationPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &escalationPolicyDataSource{}
)

func NewEscalationPolicyDataSource() datasource.DataSource {
	return &escalationPolicyDataSource{}
}

func (d *escalationPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_escalation_policy"
}

func (d *escalationPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert escalation policy by UUID or name search.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the GoAlert escalation policy to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the GoAlert escalation policy to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Escalation policy description.",
			},
			"repeat": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of times the policy steps will repeat before stopping.",
			},
		},
	}
}

func (d *escalationPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *escalationPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config escalationPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ep *client.EscalationPolicy
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		p, err := d.client.ReadEscalationPolicy(ctx, config.ID.ValueString())
		if err != nil {
			if errors.Is(err, client.ErrNotFound) {
				resp.Diagnostics.AddError("Escalation policy not found", fmt.Sprintf("Escalation policy with ID %q was not found.", config.ID.ValueString()))
				return
			}
			resp.Diagnostics.AddError("Read escalation policy failed", err.Error())
			return
		}
		ep = p
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		matches, err := d.client.SearchEscalationPolicies(ctx, name)
		if err != nil {
			resp.Diagnostics.AddError("Search escalation policies failed", err.Error())
			return
		}
		var exactMatches []client.EscalationPolicy
		for _, m := range matches {
			if m.Name == name {
				exactMatches = append(exactMatches, m)
			}
		}
		if len(exactMatches) == 0 {
			resp.Diagnostics.AddError("Escalation policy not found", fmt.Sprintf("No escalation policy with name %q was found.", name))
			return
		}
		if len(exactMatches) > 1 {
			resp.Diagnostics.AddError("Multiple escalation policies found", fmt.Sprintf("Found %d escalation policies with name %q. Use 'id' to specify a unique policy.", len(exactMatches), name))
			return
		}
		ep = &exactMatches[0]
	}

	config.ID = types.StringValue(ep.ID)
	config.Name = types.StringValue(ep.Name)
	config.Description = types.StringValue(ep.Description)
	config.Repeat = types.Int64Value(ep.Repeat)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
