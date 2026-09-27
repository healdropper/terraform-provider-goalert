package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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

var userCmUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type userContactMethodResource struct {
	client *client.Client
}

type userContactMethodModel struct {
	ID       types.String `tfsdk:"id"`
	UserID   types.String `tfsdk:"user_id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

var (
	_ resource.Resource                = &userContactMethodResource{}
	_ resource.ResourceWithConfigure   = &userContactMethodResource{}
	_ resource.ResourceWithImportState = &userContactMethodResource{}
)

func NewUserContactMethodResource() resource.Resource {
	return &userContactMethodResource{}
}

func (r *userContactMethodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_contact_method"
}

func (r *userContactMethodResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a notification contact method channel on a GoAlert user account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert contact method UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target GoAlert user UUID owning this contact method. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(userCmUUIDPattern, "must be a valid UUID"),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Descriptive label for the contact method (e.g. 'Personal Cell', 'Work Email').",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Channel destination type (`SMS`, `VOICE`, `EMAIL`, `WEBHOOK`, `SLACK_DM`). Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("SMS", "VOICE", "EMAIL", "WEBHOOK", "SLACK_DM"),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination value (e.g. E.164 phone `+15555550199`, email address, or webhook URL). Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"disabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the contact method is currently disabled.",
			},
		},
	}
}

func (r *userContactMethodResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *userContactMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userContactMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm, err := r.client.CreateUserContactMethod(ctx, client.CreateUserContactMethodInput{
		UserID: plan.UserID.ValueString(),
		Name:   plan.Name.ValueString(),
		Type:   plan.Type.ValueString(),
		Value:  plan.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating GoAlert user contact method", err.Error())
		return
	}

	plan.ID = types.StringValue(cm.ID)
	plan.Name = types.StringValue(cm.Name)
	plan.Disabled = types.BoolValue(cm.Disabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userContactMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userContactMethodModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm, err := r.client.ReadUserContactMethod(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GoAlert user contact method", err.Error())
		return
	}

	state.Name = types.StringValue(cm.Name)
	state.Disabled = types.BoolValue(cm.Disabled)
	if cmVal := cm.Value(); cmVal != "" {
		state.Value = types.StringValue(cmVal)
	}
	if cmType := cm.Type(); cmType != "" {
		state.Type = types.StringValue(cmType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userContactMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userContactMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateUserContactMethod(ctx, client.UpdateUserContactMethodInput{
		ID:   plan.ID.ValueString(),
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating GoAlert user contact method", err.Error())
		return
	}

	cm, err := r.client.ReadUserContactMethod(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error refreshing GoAlert user contact method after update", err.Error())
		return
	}

	plan.Name = types.StringValue(cm.Name)
	plan.Disabled = types.BoolValue(cm.Disabled)
	if cmVal := cm.Value(); cmVal != "" {
		plan.Value = types.StringValue(cmVal)
	}
	if cmType := cm.Type(); cmType != "" {
		plan.Type = types.StringValue(cmType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userContactMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userContactMethodModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUserContactMethod(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting GoAlert user contact method", err.Error())
		return
	}
}

func parseUserContactMethodImportID(raw string) (userID, cmID string, err error) {
	parts := strings.Split(raw, "/")
	if len(parts) == 2 && userCmUUIDPattern.MatchString(parts[0]) && userCmUUIDPattern.MatchString(parts[1]) {
		return parts[0], parts[1], nil
	}
	if len(parts) == 1 && userCmUUIDPattern.MatchString(parts[0]) {
		return "", parts[0], nil
	}
	return "", "", errors.New("import using compound ID <user_id>/<contact_method_id> or standalone UUID")
}

func (r *userContactMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userID, cmID, err := parseUserContactMethodImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), cmID)...)
	if userID != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), userID)...)
	}
}
