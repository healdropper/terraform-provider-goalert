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

type heartbeatMonitorResource struct {
	client *client.Client
}

type heartbeatMonitorModel struct {
	ID             types.String `tfsdk:"id"`
	ServiceID      types.String `tfsdk:"service_id"`
	Name           types.String `tfsdk:"name"`
	TimeoutMinutes types.Int64  `tfsdk:"timeout_minutes"`
	Href           types.String `tfsdk:"href"`
}

var (
	_ resource.Resource                = &heartbeatMonitorResource{}
	_ resource.ResourceWithConfigure   = &heartbeatMonitorResource{}
	_ resource.ResourceWithImportState = &heartbeatMonitorResource{}
)

func NewHeartbeatMonitorResource() resource.Resource {
	return &heartbeatMonitorResource{}
}

func (r *heartbeatMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_heartbeat_monitor"
}

func (r *heartbeatMonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a GoAlert dead-man switch heartbeat monitor for a service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert heartbeat monitor UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the parent GoAlert service. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Heartbeat monitor name.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"timeout_minutes": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Heartbeat timeout threshold in minutes (must be at least 5). If no ping is received within this duration, an alert is triggered.",
				Validators: []validator.Int64{
					int64validator.AtLeast(5),
				},
			},
			"href": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The ping URL to which periodic keep-alive requests must be sent.",
			},
		},
	}
}

func (r *heartbeatMonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelFromHeartbeatMonitor(hb *client.HeartbeatMonitor) heartbeatMonitorModel {
	return heartbeatMonitorModel{
		ID:             types.StringValue(hb.ID),
		ServiceID:      types.StringValue(hb.ServiceID),
		Name:           types.StringValue(hb.Name),
		TimeoutMinutes: types.Int64Value(hb.TimeoutMinutes),
		Href:           types.StringValue(hb.Href),
	}
}

func (r *heartbeatMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan heartbeatMonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CreateHeartbeatMonitorInput{
		ServiceID:      plan.ServiceID.ValueString(),
		Name:           plan.Name.ValueString(),
		TimeoutMinutes: plan.TimeoutMinutes.ValueInt64(),
	}

	hb, err := r.client.CreateHeartbeatMonitor(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create heartbeat monitor failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromHeartbeatMonitor(hb))...)
}

func (r *heartbeatMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state heartbeatMonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hb, err := r.client.ReadHeartbeatMonitor(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read heartbeat monitor failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromHeartbeatMonitor(hb))...)
}

func (r *heartbeatMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan heartbeatMonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.UpdateHeartbeatMonitorInput{
		ID:             plan.ID.ValueString(),
		Name:           plan.Name.ValueString(),
		TimeoutMinutes: plan.TimeoutMinutes.ValueInt64(),
	}

	if err := r.client.UpdateHeartbeatMonitor(ctx, input); err != nil {
		resp.Diagnostics.AddError("Update heartbeat monitor failed", err.Error())
		return
	}

	hb, err := r.client.ReadHeartbeatMonitor(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read updated heartbeat monitor failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromHeartbeatMonitor(hb))...)
}

func (r *heartbeatMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state heartbeatMonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteHeartbeatMonitor(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete heartbeat monitor failed", err.Error())
		return
	}
}

func (r *heartbeatMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid import ID", "Heartbeat monitor ID must be a lowercase UUID.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
