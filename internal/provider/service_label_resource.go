package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

type serviceLabelResource struct {
	client *client.Client
}

type serviceLabelModel struct {
	ID        types.String `tfsdk:"id"`
	ServiceID types.String `tfsdk:"service_id"`
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
}

var (
	_ resource.Resource                = &serviceLabelResource{}
	_ resource.ResourceWithConfigure   = &serviceLabelResource{}
	_ resource.ResourceWithImportState = &serviceLabelResource{}
)

func NewServiceLabelResource() resource.Resource {
	return &serviceLabelResource{}
}

func (r *serviceLabelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_label"
}

func (r *serviceLabelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a key-value label attached to a GoAlert service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Compound identifier in the format '<service_id>/<key>'.",
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
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Label key in '<domain>/<suffix>' format (e.g. 'example.com/environment'). Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Label value (3 to 255 printable characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
				},
			},
		},
	}
}

func (r *serviceLabelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceLabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceLabelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	key := plan.Key.ValueString()
	value := plan.Value.ValueString()

	if err := r.client.SetServiceLabel(ctx, serviceID, key, value); err != nil {
		resp.Diagnostics.AddError("Create service label failed", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", serviceID, key))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceLabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceLabelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	key := state.Key.ValueString()

	labels, err := r.client.ReadServiceLabels(ctx, serviceID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read service labels failed", err.Error())
		return
	}

	val, found := labels[key]
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Value = types.StringValue(val)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceLabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceLabelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := plan.ServiceID.ValueString()
	key := plan.Key.ValueString()
	value := plan.Value.ValueString()

	if err := r.client.SetServiceLabel(ctx, serviceID, key, value); err != nil {
		resp.Diagnostics.AddError("Update service label failed", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", serviceID, key))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceLabelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceID := state.ServiceID.ValueString()
	key := state.Key.ValueString()

	// Setting value to empty string removes the label in GoAlert
	if err := r.client.SetServiceLabel(ctx, serviceID, key, ""); err != nil {
		resp.Diagnostics.AddError("Delete service label failed", err.Error())
		return
	}
}

func (r *serviceLabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <service_id>/<key>
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Service label import ID must be in the format '<service_id>/<key>'.")
		return
	}

	serviceID := parts[0]
	key := parts[1]

	if !uuidPattern.MatchString(serviceID) {
		resp.Diagnostics.AddError("Invalid import ID", "The service_id component must be a lowercase UUID.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_id"), serviceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), key)...)
}
