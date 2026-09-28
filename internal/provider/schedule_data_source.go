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

type scheduleDataSource struct {
	client *client.Client
}

type scheduleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	TimeZone    types.String `tfsdk:"time_zone"`
}

var (
	_ datasource.DataSource              = &scheduleDataSource{}
	_ datasource.DataSourceWithConfigure = &scheduleDataSource{}
)

func NewScheduleDataSource() datasource.DataSource {
	return &scheduleDataSource{}
}

func (d *scheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (d *scheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert schedule by UUID or exact name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the GoAlert schedule to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exact name of the GoAlert schedule to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the schedule.",
			},
			"time_zone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IANA timezone of the schedule.",
			},
		},
	}
}

func (d *scheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *scheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config scheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sched *client.Schedule
	var err error

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		id := config.ID.ValueString()
		sched, err = d.client.ReadSchedule(ctx, id)
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Schedule not found", fmt.Sprintf("No schedule found with ID: %s", id))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Read schedule failed", err.Error())
			return
		}
	} else if !config.Name.IsNull() && config.Name.ValueString() != "" {
		name := config.Name.ValueString()
		matches, sErr := d.client.SearchSchedules(ctx, name)
		if sErr != nil {
			resp.Diagnostics.AddError("Search schedules failed", sErr.Error())
			return
		}

		var exact []client.Schedule
		for _, m := range matches {
			if m.Name == name {
				exact = append(exact, m)
			}
		}

		if len(exact) == 0 {
			resp.Diagnostics.AddError("Schedule not found", fmt.Sprintf("No schedule found with exact name: %q", name))
			return
		}
		if len(exact) > 1 {
			resp.Diagnostics.AddError("Multiple schedules found", fmt.Sprintf("Found %d schedules with exact name %q; use 'id' to disambiguate", len(exact), name))
			return
		}
		sched = &exact[0]
	} else {
		resp.Diagnostics.AddError("Missing lookup attribute", "Specify either 'id' or 'name' to read goalert_schedule.")
		return
	}

	state := scheduleDataSourceModel{
		ID:          types.StringValue(sched.ID),
		Name:        types.StringValue(sched.Name),
		Description: types.StringValue(sched.Description),
		TimeZone:    types.StringValue(sched.TimeZone),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
