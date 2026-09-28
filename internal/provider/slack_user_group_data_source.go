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

type slackUserGroupDataSource struct {
	client *client.Client
}

type slackUserGroupDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Handle types.String `tfsdk:"handle"`
}

var (
	_ datasource.DataSource              = &slackUserGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &slackUserGroupDataSource{}
)

func NewSlackUserGroupDataSource() datasource.DataSource {
	return &slackUserGroupDataSource{}
}

func (d *slackUserGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slack_user_group"
}

func (d *slackUserGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing Slack user group integrated with GoAlert by ID or exact name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Slack user group ID to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exact name of the Slack user group to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"handle": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Slack user group handle / mention name.",
			},
		},
	}
}

func (d *slackUserGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *slackUserGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config slackUserGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sug *client.SlackUserGroup
	var err error

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		id := config.ID.ValueString()
		sug, err = d.client.ReadSlackUserGroup(ctx, id)
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Slack user group not found", fmt.Sprintf("No Slack user group found with ID: %s", id))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Read Slack user group failed", err.Error())
			return
		}
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		matches, sErr := d.client.SearchSlackUserGroups(ctx, name)
		if sErr != nil {
			resp.Diagnostics.AddError("Search Slack user groups failed", sErr.Error())
			return
		}

		var exact []client.SlackUserGroup
		for _, m := range matches {
			if m.Name == name {
				exact = append(exact, m)
			}
		}

		if len(exact) == 0 {
			resp.Diagnostics.AddError("Slack user group not found", fmt.Sprintf("No Slack user group found with exact name: %q", name))
			return
		}
		if len(exact) > 1 {
			resp.Diagnostics.AddError("Multiple Slack user groups found", fmt.Sprintf("Found %d Slack user groups with exact name %q; use 'id' to disambiguate", len(exact), name))
			return
		}
		sug = &exact[0]
	} else {
		resp.Diagnostics.AddError("Missing lookup attribute", "Specify either 'id' or 'name' to read goalert_slack_user_group.")
		return
	}

	state := slackUserGroupDataSourceModel{
		ID:     types.StringValue(sug.ID),
		Name:   types.StringValue(sug.Name),
		Handle: types.StringValue(sug.Handle),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
