package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

type userOverrideResource struct {
	client *client.Client
}

type userOverrideModel struct {
	ID           types.String `tfsdk:"id"`
	ScheduleID   types.String `tfsdk:"schedule_id"`
	StartTime    types.String `tfsdk:"start_time"`
	EndTime      types.String `tfsdk:"end_time"`
	AddUserID    types.String `tfsdk:"add_user_id"`
	RemoveUserID types.String `tfsdk:"remove_user_id"`
}

var (
	_ resource.Resource                = &userOverrideResource{}
	_ resource.ResourceWithConfigure   = &userOverrideResource{}
	_ resource.ResourceWithImportState = &userOverrideResource{}
)

func NewUserOverrideResource() resource.Resource {
	return &userOverrideResource{}
}

func (r *userOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_override"
}

func (r *userOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages temporary shift coverage or user replacement on a GoAlert schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert user override UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the schedule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"start_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "RFC3339 / ISO-8601 start timestamp for the override.",
			},
			"end_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "RFC3339 / ISO-8601 end timestamp for the override.",
			},
			"add_user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "UUID of the user providing coverage. At least one of add_user_id or remove_user_id is required.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"remove_user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "UUID of the user being replaced. At least one of add_user_id or remove_user_id is required.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
		},
	}
}

func (r *userOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider client", "Expected a GoAlert API client.")
		return
	}
	r.client = c
}

func (r *userOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (plan.AddUserID.IsNull() || plan.AddUserID.ValueString() == "") &&
		(plan.RemoveUserID.IsNull() || plan.RemoveUserID.ValueString() == "") {
		resp.Diagnostics.AddError("Invalid user override", "At least one of add_user_id or remove_user_id must be provided.")
		return
	}

	input := client.CreateUserOverrideInput{
		ScheduleID: plan.ScheduleID.ValueString(),
		Start:      plan.StartTime.ValueString(),
		End:        plan.EndTime.ValueString(),
	}
	if !plan.AddUserID.IsNull() && plan.AddUserID.ValueString() != "" {
		addID := plan.AddUserID.ValueString()
		input.AddUserID = &addID
	}
	if !plan.RemoveUserID.IsNull() && plan.RemoveUserID.ValueString() != "" {
		remID := plan.RemoveUserID.ValueString()
		input.RemoveUserID = &remID
	}

	uo, err := r.client.CreateUserOverride(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create user override failed", err.Error()+". If the server committed before a connection failure, locate and import the override before retrying.")
		return
	}

	plan.ID = types.StringValue(uo.ID)
	plan.ScheduleID = types.StringValue(uo.Target.ID)
	plan.StartTime = types.StringValue(uo.Start)
	plan.EndTime = types.StringValue(uo.End)
	if uo.AddUserID != "" {
		plan.AddUserID = types.StringValue(uo.AddUserID)
	} else {
		plan.AddUserID = types.StringNull()
	}
	if uo.RemoveUserID != "" {
		plan.RemoveUserID = types.StringValue(uo.RemoveUserID)
	} else {
		plan.RemoveUserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uo, err := r.client.ReadUserOverride(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read user override failed", err.Error())
		return
	}

	state.StartTime = types.StringValue(uo.Start)
	state.EndTime = types.StringValue(uo.End)
	if uo.Target.ID != "" {
		state.ScheduleID = types.StringValue(uo.Target.ID)
	}
	if uo.AddUserID != "" {
		state.AddUserID = types.StringValue(uo.AddUserID)
	} else {
		state.AddUserID = types.StringNull()
	}
	if uo.RemoveUserID != "" {
		state.RemoveUserID = types.StringValue(uo.RemoveUserID)
	} else {
		state.RemoveUserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userOverrideModel
	var state userOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (plan.AddUserID.IsNull() || plan.AddUserID.ValueString() == "") &&
		(plan.RemoveUserID.IsNull() || plan.RemoveUserID.ValueString() == "") {
		resp.Diagnostics.AddError("Invalid user override", "At least one of add_user_id or remove_user_id must be provided.")
		return
	}

	start := plan.StartTime.ValueString()
	end := plan.EndTime.ValueString()
	input := client.UpdateUserOverrideInput{
		ID:    state.ID.ValueString(),
		Start: &start,
		End:   &end,
	}
	if !plan.AddUserID.IsNull() && plan.AddUserID.ValueString() != "" {
		addID := plan.AddUserID.ValueString()
		input.AddUserID = &addID
	}
	if !plan.RemoveUserID.IsNull() && plan.RemoveUserID.ValueString() != "" {
		remID := plan.RemoveUserID.ValueString()
		input.RemoveUserID = &remID
	}

	if err := r.client.UpdateUserOverride(ctx, input); err != nil {
		resp.Diagnostics.AddError("Update user override failed", err.Error())
		return
	}

	updated, err := r.client.ReadUserOverride(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read updated user override failed", err.Error())
		return
	}

	plan.ID = state.ID
	plan.ScheduleID = state.ScheduleID
	plan.StartTime = types.StringValue(updated.Start)
	plan.EndTime = types.StringValue(updated.End)
	if updated.AddUserID != "" {
		plan.AddUserID = types.StringValue(updated.AddUserID)
	} else {
		plan.AddUserID = types.StringNull()
	}
	if updated.RemoveUserID != "" {
		plan.RemoveUserID = types.StringValue(updated.RemoveUserID)
	} else {
		plan.RemoveUserID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUserOverride(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Delete user override failed", err.Error())
	}
}

func (r *userOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid import ID", "Import using the user override's lowercase UUID.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
