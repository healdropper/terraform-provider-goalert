package provider

import (
	"context"
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

var rotationNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\-_' ]+$`)

type rotationResource struct {
	client *client.Client
}

type rotationModel struct {
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
	_ resource.Resource                = &rotationResource{}
	_ resource.ResourceWithConfigure   = &rotationResource{}
	_ resource.ResourceWithImportState = &rotationResource{}
)

func NewRotationResource() resource.Resource {
	return &rotationResource{}
}

func (r *rotationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation"
}

func (r *rotationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an on-call shift rotation in GoAlert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert rotation UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the rotation. Allowed characters: alphanumeric, hyphens, underscores, apostrophes, and spaces.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(rotationNamePattern, "can only contain letters, digits, hyphens, underscores, apostrophes, and spaces"),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional description of the rotation.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(6144),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Rotation frequency type: 'daily', 'weekly', or 'hourly'.",
				Validators: []validator.String{
					stringvalidator.OneOf("daily", "weekly", "hourly"),
				},
			},
			"start_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Anchor start time in ISO-8601/RFC3339 format determining handoff schedule.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"time_zone": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Valid IANA time zone identifier (e.g. 'Europe/Madrid', 'UTC', 'America/New_York').",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"shift_length": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Duration multiplier for the rotation type. Defaults to 1.",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"user_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "Ordered sequence of user UUIDs participating in the rotation.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"active_user_index": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Zero-indexed position of the currently active participant.",
			},
		},
	}
}

func (r *rotationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client")
		return
	}
	r.client = c
}

func modelFromRotation(rot *client.Rotation, prior *rotationModel) rotationModel {
	userIDs := make([]types.String, len(rot.UserIDs))
	for i, uid := range rot.UserIDs {
		userIDs[i] = types.StringValue(uid)
	}

	m := rotationModel{
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

	if prior != nil && !prior.Description.IsNull() && prior.Description.ValueString() == "" && rot.Description == "" {
		m.Description = prior.Description
	}

	return m
}

func (r *rotationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rotationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userIDs := make([]string, len(plan.UserIDs))
	for i, u := range plan.UserIDs {
		userIDs[i] = u.ValueString()
	}

	shiftLength := plan.ShiftLength.ValueInt64()
	input := client.CreateRotationInput{
		Name:        plan.Name.ValueString(),
		Type:        plan.Type.ValueString(),
		Start:       plan.StartTime.ValueString(),
		TimeZone:    plan.TimeZone.ValueString(),
		ShiftLength: &shiftLength,
		UserIDs:     userIDs,
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		input.Description = &desc
	}

	rot, err := r.client.CreateRotation(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create rotation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromRotation(rot, &plan))...)
}

func (r *rotationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rotationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rot, err := r.client.ReadRotation(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read rotation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromRotation(rot, &state))...)
}

func (r *rotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rotationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userIDs := make([]string, len(plan.UserIDs))
	for i, u := range plan.UserIDs {
		userIDs[i] = u.ValueString()
	}

	id := plan.ID.ValueString()
	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	rotType := plan.Type.ValueString()
	startTime := plan.StartTime.ValueString()
	timeZone := plan.TimeZone.ValueString()
	shiftLength := plan.ShiftLength.ValueInt64()

	input := client.UpdateRotationInput{
		ID:          id,
		Name:        &name,
		Description: &desc,
		Type:        &rotType,
		Start:       &startTime,
		TimeZone:    &timeZone,
		ShiftLength: &shiftLength,
		UserIDs:     &userIDs,
	}

	if err := r.client.UpdateRotation(ctx, input); err != nil {
		resp.Diagnostics.AddError("Update rotation failed", err.Error())
		return
	}

	updated, err := r.client.ReadRotation(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Read updated rotation failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromRotation(updated, &plan))...)
}

func (r *rotationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rotationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRotation(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete rotation failed", err.Error())
		return
	}
}

func (r *rotationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
