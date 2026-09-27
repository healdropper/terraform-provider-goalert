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

type rotationDataSource struct {
	client *client.Client
}

type rotationDataSourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Description     types.String   `tfsdk:"description"`
	Type            types.String   `tfsdk:"type"`
	StartTime       types.String   `tfsdk:"start_time"`
	TimeZone        types.String   `tfsdk:"time_zone"`
	ShiftLength     types.Int64    `tfsdk:"shift_length"`
	UserIDs         []types.String `tfsdk:"user_ids"`
	ActiveUserIndex types.Int64    `tfsdk:"active_user_index"`
}

var (
	_ datasource.DataSource              = &rotationDataSource{}
	_ datasource.DataSourceWithConfigure = &rotationDataSource{}
)

func NewRotationDataSource() datasource.DataSource {
	return &rotationDataSource{}
}

func (d *rotationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation"
}

func (d *rotationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert rotation by UUID or exact name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the GoAlert rotation to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Exact name of the GoAlert rotation to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the rotation.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Frequency type of the rotation ('daily', 'weekly', 'hourly').",
			},
			"start_time": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Anchor start time in ISO-8601/RFC3339 format.",
			},
			"time_zone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IANA time zone identifier.",
			},
			"shift_length": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Duration multiplier for the rotation type.",
			},
			"user_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Ordered sequence of user UUIDs participating in the rotation.",
			},
			"active_user_index": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Zero-indexed position of the currently active participant.",
			},
		},
	}
}

func (d *rotationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *rotationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config rotationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rot *client.Rotation
	if !config.ID.IsNull() && !config.ID.IsUnknown() {
		id := config.ID.ValueString()
		var err error
		rot, err = d.client.ReadRotation(ctx, id)
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Rotation not found", fmt.Sprintf("no rotation found with ID %q", id))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Read rotation failed", err.Error())
			return
		}
	} else if !config.Name.IsNull() && !config.Name.IsUnknown() {
		name := config.Name.ValueString()
		rotations, err := d.client.SearchRotations(ctx, name)
		if err != nil {
			resp.Diagnostics.AddError("Search rotations failed", err.Error())
			return
		}

		var matched []client.Rotation
		for _, r := range rotations {
			if r.Name == name {
				matched = append(matched, r)
			}
		}

		if len(matched) == 0 {
			resp.Diagnostics.AddError("Rotation not found", fmt.Sprintf("no rotation found with name %q", name))
			return
		}
		if len(matched) > 1 {
			resp.Diagnostics.AddError("Multiple rotations found", fmt.Sprintf("found %d rotations matching name %q; specify id to disambiguate", len(matched), name))
			return
		}
		rot = &matched[0]
	} else {
		resp.Diagnostics.AddError("Missing lookup argument", "one of id or name must be specified")
		return
	}

	userIDs := make([]types.String, len(rot.UserIDs))
	for i, uid := range rot.UserIDs {
		userIDs[i] = types.StringValue(uid)
	}

	state := rotationDataSourceModel{
		ID:              types.StringValue(rot.ID),
		Name:            types.StringValue(rot.Name),
		Description:     types.StringValue(rot.Description),
		Type:            types.StringValue(rot.Type),
		StartTime:       types.StringValue(rot.Start),
		TimeZone:        types.StringValue(rot.TimeZone),
		ShiftLength:     types.Int64Value(rot.ShiftLength),
		UserIDs:         userIDs,
		ActiveUserIndex: types.Int64Value(rot.ActiveUserIndex),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
