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

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type serviceResource struct{ client *client.Client }
type serviceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	EscalationPolicyID types.String `tfsdk:"escalation_policy_id"`
}

var (
	_ resource.Resource                = &serviceResource{}
	_ resource.ResourceWithConfigure   = &serviceResource{}
	_ resource.ResourceWithImportState = &serviceResource{}
)

func NewServiceResource() resource.Resource { return &serviceResource{} }
func (r *serviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}
func (r *serviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A GoAlert service. The escalation policy must already exist; deleting the service does not delete its policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "GoAlert service UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "Service name.",
				Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
			"description": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Service description. Omission means an empty description.",
				Validators:          []validator.String{stringvalidator.LengthAtMost(6144)}},
			"escalation_policy_id": schema.StringAttribute{Required: true, MarkdownDescription: "UUID of an existing escalation policy.",
				Validators: []validator.String{stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID")}},
		},
	}
}
func (r *serviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func modelFromService(s *client.Service) serviceModel {
	return serviceModel{ID: types.StringValue(s.ID), Name: types.StringValue(s.Name),
		Description: types.StringValue(s.Description), EscalationPolicyID: types.StringValue(s.EscalationPolicy.ID)}
}
func (r *serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	svc, err := r.client.CreateService(ctx, plan.Name.ValueString(), plan.Description.ValueString(), plan.EscalationPolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create service failed", err.Error()+". If the server committed before a connection failure, locate and import the service before retrying.")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromService(svc))...)
}
func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	svc, err := r.client.ReadService(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read service failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromService(svc))...)
}
func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.UpdateService(ctx, plan.ID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString(), plan.EscalationPolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update service failed", err.Error())
		return
	}
	svc, err := r.client.ReadService(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read updated service failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromService(svc))...)
}
func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteService(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete service failed", err.Error())
	}
}
func (r *serviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid import ID", "Import using the service's lowercase UUID.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
