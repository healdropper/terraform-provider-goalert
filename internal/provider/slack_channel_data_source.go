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

type slackChannelDataSource struct {
	client *client.Client
}

type slackChannelDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	TeamID types.String `tfsdk:"team_id"`
}

var (
	_ datasource.DataSource              = &slackChannelDataSource{}
	_ datasource.DataSourceWithConfigure = &slackChannelDataSource{}
)

func NewSlackChannelDataSource() datasource.DataSource {
	return &slackChannelDataSource{}
}

func (d *slackChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slack_channel"
}

func (d *slackChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing Slack channel integrated with GoAlert by ID or exact name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Slack channel ID to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exact name of the Slack channel to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"team_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Slack workspace team ID associated with the channel.",
			},
		},
	}
}

func (d *slackChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *slackChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config slackChannelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sc *client.SlackChannel
	var err error

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		id := config.ID.ValueString()
		sc, err = d.client.ReadSlackChannel(ctx, id)
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Slack channel not found", fmt.Sprintf("No Slack channel found with ID: %s", id))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Read Slack channel failed", err.Error())
			return
		}
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		matches, sErr := d.client.SearchSlackChannels(ctx, name)
		if sErr != nil {
			resp.Diagnostics.AddError("Search Slack channels failed", sErr.Error())
			return
		}

		var exact []client.SlackChannel
		for _, m := range matches {
			if m.Name == name {
				exact = append(exact, m)
			}
		}

		if len(exact) == 0 {
			resp.Diagnostics.AddError("Slack channel not found", fmt.Sprintf("No Slack channel found with exact name: %q", name))
			return
		}
		if len(exact) > 1 {
			resp.Diagnostics.AddError("Multiple Slack channels found", fmt.Sprintf("Found %d Slack channels with exact name %q; use 'id' to disambiguate", len(exact), name))
			return
		}
		sc = &exact[0]
	} else {
		resp.Diagnostics.AddError("Missing lookup attribute", "Specify either 'id' or 'name' to read goalert_slack_channel.")
		return
	}

	state := slackChannelDataSourceModel{
		ID:     types.StringValue(sc.ID),
		Name:   types.StringValue(sc.Name),
		TeamID: types.StringValue(sc.TeamID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
