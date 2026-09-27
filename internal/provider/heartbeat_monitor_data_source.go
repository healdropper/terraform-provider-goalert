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

type heartbeatMonitorDataSource struct {
	client *client.Client
}

type heartbeatMonitorDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ServiceID      types.String `tfsdk:"service_id"`
	Name           types.String `tfsdk:"name"`
	TimeoutMinutes types.Int64  `tfsdk:"timeout_minutes"`
	Href           types.String `tfsdk:"href"`
}

var (
	_ datasource.DataSource              = &heartbeatMonitorDataSource{}
	_ datasource.DataSourceWithConfigure = &heartbeatMonitorDataSource{}
)

func NewHeartbeatMonitorDataSource() datasource.DataSource {
	return &heartbeatMonitorDataSource{}
}

func (d *heartbeatMonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_heartbeat_monitor"
}

func (d *heartbeatMonitorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert heartbeat monitor by UUID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the GoAlert heartbeat monitor.",
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
				MarkdownDescription: "Heartbeat monitor name.",
			},
			"timeout_minutes": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Heartbeat timeout threshold in minutes.",
			},
			"href": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The ping URL to which periodic keep-alive requests must be sent.",
			},
		},
	}
}

func (d *heartbeatMonitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *heartbeatMonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config heartbeatMonitorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hb, err := d.client.ReadHeartbeatMonitor(ctx, config.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Heartbeat monitor not found", fmt.Sprintf("Heartbeat monitor with ID %q was not found.", config.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Read heartbeat monitor failed", err.Error())
		return
	}

	config.ID = types.StringValue(hb.ID)
	config.ServiceID = types.StringValue(hb.ServiceID)
	config.Name = types.StringValue(hb.Name)
	config.TimeoutMinutes = types.Int64Value(hb.TimeoutMinutes)
	config.Href = types.StringValue(hb.Href)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
