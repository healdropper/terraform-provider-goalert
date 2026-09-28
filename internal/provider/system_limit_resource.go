package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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

type systemLimitResource struct {
	client *client.Client
}

type systemLimitModel struct {
	ID          types.String `tfsdk:"id"`
	Value       types.Int64  `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

var (
	_ resource.Resource                = &systemLimitResource{}
	_ resource.ResourceWithConfigure   = &systemLimitResource{}
	_ resource.ResourceWithImportState = &systemLimitResource{}
)

func NewSystemLimitResource() resource.Resource {
	return &systemLimitResource{}
}

func (r *systemLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_limit"
}

func (r *systemLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a global system limit configuration in GoAlert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identifier of the system limit matching GoAlert's SystemLimitID enum (e.g. 'RulesPerSchedule'). Requires replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"CalendarSubscriptionsPerUser",
						"ContactMethodsPerUser",
						"EPActionsPerStep",
						"EPStepsPerPolicy",
						"HeartbeatMonitorsPerService",
						"IntegrationKeysPerService",
						"NotificationRulesPerUser",
						"ParticipantsPerRotation",
						"PendingSignalsPerDestPerService",
						"PendingSignalsPerService",
						"RulesPerSchedule",
						"TargetsPerSchedule",
						"UnackedAlertsPerService",
						"UserOverridesPerSchedule",
					),
				},
			},
			"value": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Configured threshold for this system limit. Must be greater than or equal to 0.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the system limit provided by GoAlert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *systemLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *systemLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limitID := plan.ID.ValueString()
	val := plan.Value.ValueInt64()

	err := r.client.SetSystemLimits(ctx, []client.SystemLimitInput{
		{
			ID:    limitID,
			Value: val,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to set system limit", err.Error())
		return
	}

	lim, err := r.client.ReadSystemLimit(ctx, limitID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read system limit after setting", err.Error())
		return
	}

	plan.Value = types.Int64Value(lim.Value)
	plan.Description = types.StringValue(lim.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemLimitModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lim, err := r.client.ReadSystemLimit(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read system limit", err.Error())
		return
	}

	state.Value = types.Int64Value(lim.Value)
	state.Description = types.StringValue(lim.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *systemLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan systemLimitModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limitID := plan.ID.ValueString()
	val := plan.Value.ValueInt64()

	err := r.client.SetSystemLimits(ctx, []client.SystemLimitInput{
		{
			ID:    limitID,
			Value: val,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update system limit", err.Error())
		return
	}

	lim, err := r.client.ReadSystemLimit(ctx, limitID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read system limit after updating", err.Error())
		return
	}

	plan.Value = types.Int64Value(lim.Value)
	plan.Description = types.StringValue(lim.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *systemLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// GoAlert system limits cannot be deleted via API as they are permanent system configuration objects.
	// We cleanly remove the resource from Terraform state.
	resp.State.RemoveResource(ctx)
}

func (r *systemLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
