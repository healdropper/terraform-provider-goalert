package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

var clockTimePattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

type scheduleRuleResource struct {
	client *client.Client
}

type scheduleRuleModel struct {
	ID            types.String `tfsdk:"id"`
	ScheduleID    types.String `tfsdk:"schedule_id"`
	TargetType    types.String `tfsdk:"target_type"`
	TargetID      types.String `tfsdk:"target_id"`
	StartTime     types.String `tfsdk:"start_time"`
	EndTime       types.String `tfsdk:"end_time"`
	WeekdayFilter []types.Bool `tfsdk:"weekday_filter"`
}

var (
	_ resource.Resource                = &scheduleRuleResource{}
	_ resource.ResourceWithConfigure   = &scheduleRuleResource{}
	_ resource.ResourceWithImportState = &scheduleRuleResource{}
)

func NewScheduleRuleResource() resource.Resource {
	return &scheduleRuleResource{}
}

func (r *scheduleRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule_rule"
}

func (r *scheduleRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an active target coverage rule on a GoAlert schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Compound identifier `<schedule_id>:<target_type>:<target_id>`.",
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
			"target_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target entity type: 'rotation' or 'user'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("rotation", "user"),
				},
			},
			"target_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the target entity (rotation or user).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"start_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Shift start clock time in 24-hour 'HH:MM' format (e.g. '09:00').",
				Validators: []validator.String{
					stringvalidator.RegexMatches(clockTimePattern, "must be in 'HH:MM' 24-hour format"),
				},
			},
			"end_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Shift end clock time in 24-hour 'HH:MM' format (e.g. '17:00').",
				Validators: []validator.String{
					stringvalidator.RegexMatches(clockTimePattern, "must be in 'HH:MM' 24-hour format"),
				},
			},
			"weekday_filter": schema.ListAttribute{
				ElementType:         types.BoolType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Active days of the week as a 7-element boolean array for [Sun, Mon, Tue, Wed, Thu, Fri, Sat]. Defaults to all true.",
				Validators: []validator.List{
					listvalidator.SizeBetween(7, 7),
				},
			},
		},
	}
}

func (r *scheduleRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func defaultWeekdayFilter() []types.Bool {
	return []types.Bool{
		types.BoolValue(true),
		types.BoolValue(true),
		types.BoolValue(true),
		types.BoolValue(true),
		types.BoolValue(true),
		types.BoolValue(true),
		types.BoolValue(true),
	}
}

func weekdayBoolsToSlice(tb []types.Bool) []bool {
	res := make([]bool, len(tb))
	for i, b := range tb {
		res[i] = b.ValueBool()
	}
	return res
}

func sliceToWeekdayBools(b []bool) []types.Bool {
	res := make([]types.Bool, len(b))
	for i, val := range b {
		res[i] = types.BoolValue(val)
	}
	return res
}

func (r *scheduleRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scheduleRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := plan.WeekdayFilter
	if len(filter) == 0 {
		filter = defaultWeekdayFilter()
	}

	scheduleID := plan.ScheduleID.ValueString()
	targetType := plan.TargetType.ValueString()
	targetID := plan.TargetID.ValueString()

	input := client.ScheduleTargetInput{
		ScheduleID: scheduleID,
		Target: client.TargetInput{
			ID:   targetID,
			Type: targetType,
		},
		Rules: []client.ScheduleRuleInput{
			{
				Start:         plan.StartTime.ValueString(),
				End:           plan.EndTime.ValueString(),
				WeekdayFilter: weekdayBoolsToSlice(filter),
			},
		},
	}

	if err := r.client.UpdateScheduleTarget(ctx, input); err != nil {
		resp.Diagnostics.AddError("Create schedule rule failed", err.Error())
		return
	}

	compoundID := fmt.Sprintf("%s:%s:%s", scheduleID, targetType, targetID)
	plan.ID = types.StringValue(compoundID)
	plan.WeekdayFilter = filter
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scheduleRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := state.ScheduleID.ValueString()
	targetType := state.TargetType.ValueString()
	targetID := state.TargetID.ValueString()

	sched, err := r.client.ReadSchedule(ctx, scheduleID)
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read schedule rule failed", err.Error())
		return
	}

	var foundTarget *client.ScheduleTarget
	for i := range sched.Targets {
		t := &sched.Targets[i]
		if t.Target.ID == targetID && t.Target.Type == targetType {
			foundTarget = t
			break
		}
	}

	if foundTarget == nil || len(foundTarget.Rules) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	rule := foundTarget.Rules[0]
	state.StartTime = types.StringValue(rule.Start)
	state.EndTime = types.StringValue(rule.End)
	state.WeekdayFilter = sliceToWeekdayBools(rule.WeekdayFilter)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scheduleRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scheduleRuleModel
	var state scheduleRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := plan.WeekdayFilter
	if len(filter) == 0 {
		filter = defaultWeekdayFilter()
	}

	scheduleID := state.ScheduleID.ValueString()
	targetType := state.TargetType.ValueString()
	targetID := state.TargetID.ValueString()

	input := client.ScheduleTargetInput{
		ScheduleID: scheduleID,
		Target: client.TargetInput{
			ID:   targetID,
			Type: targetType,
		},
		Rules: []client.ScheduleRuleInput{
			{
				Start:         plan.StartTime.ValueString(),
				End:           plan.EndTime.ValueString(),
				WeekdayFilter: weekdayBoolsToSlice(filter),
			},
		},
	}

	if err := r.client.UpdateScheduleTarget(ctx, input); err != nil {
		resp.Diagnostics.AddError("Update schedule rule failed", err.Error())
		return
	}

	plan.ID = state.ID
	plan.WeekdayFilter = filter
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scheduleRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := state.ScheduleID.ValueString()
	targetType := state.TargetType.ValueString()
	targetID := state.TargetID.ValueString()

	input := client.ScheduleTargetInput{
		ScheduleID: scheduleID,
		Target: client.TargetInput{
			ID:   targetID,
			Type: targetType,
		},
		Rules: []client.ScheduleRuleInput{},
	}

	err := r.client.UpdateScheduleTarget(ctx, input)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		resp.Diagnostics.AddError("Delete schedule rule failed", err.Error())
	}
}

func (r *scheduleRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 || !uuidPattern.MatchString(parts[0]) || (parts[1] != "rotation" && parts[1] != "user") || !uuidPattern.MatchString(parts[2]) {
		resp.Diagnostics.AddError("Invalid import ID", "Import format must be '<schedule_id>:<target_type>:<target_id>' with lowercase UUIDs.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schedule_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_id"), parts[2])...)
}
