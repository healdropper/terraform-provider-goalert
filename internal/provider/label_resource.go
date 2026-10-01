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

type labelResource struct {
	client *client.Client
}

type labelModel struct {
	ID         types.String `tfsdk:"id"`
	TargetType types.String `tfsdk:"target_type"`
	TargetID   types.String `tfsdk:"target_id"`
	Key        types.String `tfsdk:"key"`
	Value      types.String `tfsdk:"value"`
}

var (
	_ resource.Resource                = &labelResource{}
	_ resource.ResourceWithConfigure   = &labelResource{}
	_ resource.ResourceWithImportState = &labelResource{}
)

func NewLabelResource() resource.Resource {
	return &labelResource{}
}

func (r *labelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (r *labelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a key-value label attached to a GoAlert service, escalation policy, schedule, or rotation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Compound identifier in the format '<target_type>:<target_id>/<key>'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target entity type (`service`, `escalation_policy`, `schedule`, or `rotation`). Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("service", "escalation_policy", "schedule", "rotation"),
				},
			},
			"target_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the target GoAlert entity. Changing this forces replacement.",
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

func (r *labelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *labelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan labelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := plan.TargetType.ValueString()
	targetID := plan.TargetID.ValueString()
	key := plan.Key.ValueString()
	value := plan.Value.ValueString()

	if err := r.client.SetLabel(ctx, targetType, targetID, key, value); err != nil {
		resp.Diagnostics.AddError("Create label failed", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s/%s", targetType, targetID, key))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *labelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state labelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := state.TargetType.ValueString()
	targetID := state.TargetID.ValueString()
	key := state.Key.ValueString()

	labels, err := r.client.ReadTargetLabels(ctx, targetType, targetID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read labels failed", err.Error())
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

func (r *labelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan labelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := plan.TargetType.ValueString()
	targetID := plan.TargetID.ValueString()
	key := plan.Key.ValueString()
	value := plan.Value.ValueString()

	if err := r.client.SetLabel(ctx, targetType, targetID, key, value); err != nil {
		resp.Diagnostics.AddError("Update label failed", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s/%s", targetType, targetID, key))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *labelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state labelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := state.TargetType.ValueString()
	targetID := state.TargetID.ValueString()
	key := state.Key.ValueString()

	if err := r.client.SetLabel(ctx, targetType, targetID, key, ""); err != nil {
		resp.Diagnostics.AddError("Delete label failed", err.Error())
		return
	}
}

func parseLabelImportID(raw string) (targetType, targetID, key string, err error) {
	colonParts := strings.SplitN(raw, ":", 2)
	if len(colonParts) != 2 || colonParts[0] == "" || colonParts[1] == "" {
		return "", "", "", errors.New("label import ID must be in the format '<target_type>:<target_id>/<key>'")
	}
	targetType = colonParts[0]
	switch targetType {
	case "service", "escalation_policy", "schedule", "rotation":
	default:
		return "", "", "", fmt.Errorf("unsupported target_type %q; must be one of service, escalation_policy, schedule, rotation", targetType)
	}
	slashParts := strings.SplitN(colonParts[1], "/", 2)
	if len(slashParts) != 2 || slashParts[0] == "" || slashParts[1] == "" {
		return "", "", "", errors.New("label import ID must be in the format '<target_type>:<target_id>/<key>'")
	}
	targetID = slashParts[0]
	key = slashParts[1]
	if !uuidPattern.MatchString(targetID) {
		return "", "", "", errors.New("the target_id component must be a lowercase UUID")
	}
	return targetType, targetID, key, nil
}

func (r *labelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	targetType, targetID, key, err := parseLabelImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_type"), targetType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_id"), targetID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), key)...)
}
