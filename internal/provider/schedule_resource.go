package provider

import (
	"context"
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

var scheduleNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\-_' ]+$`)

type scheduleResource struct {
	client *client.Client
}

type scheduleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	TimeZone    types.String `tfsdk:"time_zone"`
}

var (
	_ resource.Resource                = &scheduleResource{}
	_ resource.ResourceWithConfigure   = &scheduleResource{}
	_ resource.ResourceWithImportState = &scheduleResource{}
)

func NewScheduleResource() resource.Resource {
	return &scheduleResource{}
}

func (r *scheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *scheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an on-call schedule in GoAlert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert schedule UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the schedule.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(scheduleNamePattern, "can only contain letters, digits, hyphens, underscores, apostrophes, and spaces"),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional description of the schedule.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(6144),
				},
			},
			"time_zone": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IANA timezone for schedule shifts and daylight savings calculation (e.g. 'Europe/Madrid', 'UTC').",
			},
		},
	}
}

func (r *scheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelFromSchedule(sched *client.Schedule) scheduleModel {
	return scheduleModel{
		ID:          types.StringValue(sched.ID),
		Name:        types.StringValue(sched.Name),
		Description: types.StringValue(sched.Description),
		TimeZone:    types.StringValue(sched.TimeZone),
	}
}

func (r *scheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	tz := plan.TimeZone.ValueString()
	input := client.CreateScheduleInput{
		Name:     name,
		TimeZone: tz,
	}
	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		input.Description = &desc
	}

	sched, err := r.client.CreateSchedule(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create schedule failed", err.Error()+". If the server committed before a connection failure, locate and import the schedule before retrying.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromSchedule(sched))...)
}

func (r *scheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sched, err := r.client.ReadSchedule(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read schedule failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromSchedule(sched))...)
}

func (r *scheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scheduleModel
	var state scheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	tz := plan.TimeZone.ValueString()
	input := client.UpdateScheduleInput{
		ID:          state.ID.ValueString(),
		Name:        &name,
		Description: &desc,
		TimeZone:    &tz,
	}

	if err := r.client.UpdateSchedule(ctx, input); err != nil {
		resp.Diagnostics.AddError("Update schedule failed", err.Error())
		return
	}

	updated, err := r.client.ReadSchedule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read updated schedule failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromSchedule(updated))...)
}

func (r *scheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSchedule(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Delete schedule failed", err.Error())
	}
}

func (r *scheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid import ID", "Import using the schedule's lowercase UUID.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
