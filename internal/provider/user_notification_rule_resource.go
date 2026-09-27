package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

var userNrUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type userNotificationRuleResource struct {
	client *client.Client
}

type userNotificationRuleModel struct {
	ID              types.String `tfsdk:"id"`
	UserID          types.String `tfsdk:"user_id"`
	ContactMethodID types.String `tfsdk:"contact_method_id"`
	DelayMinutes    types.Int64  `tfsdk:"delay_minutes"`
}

var (
	_ resource.Resource                = &userNotificationRuleResource{}
	_ resource.ResourceWithConfigure   = &userNotificationRuleResource{}
	_ resource.ResourceWithImportState = &userNotificationRuleResource{}
)

func NewUserNotificationRuleResource() resource.Resource {
	return &userNotificationRuleResource{}
}

func (r *userNotificationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_notification_rule"
}

func (r *userNotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an individual notification rule for an operator in GoAlert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert notification rule UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target GoAlert user UUID owning this rule. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(userNrUUIDPattern, "must be a valid UUID"),
				},
			},
			"contact_method_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target contact method UUID to notify. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(userNrUUIDPattern, "must be a valid UUID"),
				},
			},
			"delay_minutes": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Minutes to wait after alert trigger before notifying (non-negative). Defaults to 0.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (r *userNotificationRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userNotificationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userNotificationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nr, err := r.client.CreateUserNotificationRule(ctx, client.CreateUserNotificationRuleInput{
		UserID:          plan.UserID.ValueString(),
		ContactMethodID: plan.ContactMethodID.ValueString(),
		DelayMinutes:    plan.DelayMinutes.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating GoAlert user notification rule", err.Error())
		return
	}

	plan.ID = types.StringValue(nr.ID)
	plan.ContactMethodID = types.StringValue(nr.ContactMethodID)
	plan.DelayMinutes = types.Int64Value(nr.DelayMinutes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userNotificationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userNotificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nr, err := r.client.ReadUserNotificationRule(ctx, state.UserID.ValueString(), state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GoAlert user notification rule", err.Error())
		return
	}

	state.ContactMethodID = types.StringValue(nr.ContactMethodID)
	state.DelayMinutes = types.Int64Value(nr.DelayMinutes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userNotificationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Invalid Update on Immutable Resource",
		"GoAlert user notification rules are immutable. Any modification forces resource replacement.",
	)
}

func (r *userNotificationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userNotificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUserNotificationRule(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting GoAlert user notification rule", err.Error())
		return
	}
}

func parseUserNotificationRuleImportID(raw string) (userID, ruleID string, err error) {
	parts := strings.Split(raw, "/")
	if len(parts) == 2 && userNrUUIDPattern.MatchString(parts[0]) && userNrUUIDPattern.MatchString(parts[1]) {
		return parts[0], parts[1], nil
	}
	return "", "", errors.New("import requires compound format <user_id>/<rule_id>")
}

func (r *userNotificationRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userID, ruleID, err := parseUserNotificationRuleImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ruleID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), userID)...)
}
