package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

var webhookURLPattern = regexp.MustCompile(`^https?://.+$`)

type escalationPolicyResource struct {
	client *client.Client
}

type escalationPolicyModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Repeat      types.Int64  `tfsdk:"repeat"`
	Steps       []stepModel  `tfsdk:"step"`
}

type stepModel struct {
	ID             types.String         `tfsdk:"id"`
	StepNumber     types.Int64          `tfsdk:"step_number"`
	DelayMinutes   types.Int64          `tfsdk:"delay_minutes"`
	UserIDs        []types.String       `tfsdk:"user_ids"`
	RotationIDs    []types.String       `tfsdk:"rotation_ids"`
	WebhookActions []webhookActionModel `tfsdk:"webhook_action"`
}

type webhookActionModel struct {
	URL types.String `tfsdk:"url"`
}

var (
	_ resource.Resource                = &escalationPolicyResource{}
	_ resource.ResourceWithConfigure   = &escalationPolicyResource{}
	_ resource.ResourceWithImportState = &escalationPolicyResource{}
)

func NewEscalationPolicyResource() resource.Resource {
	return &escalationPolicyResource{}
}

func (r *escalationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_escalation_policy"
}

func (r *escalationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A GoAlert escalation policy with ordered steps and notification destinations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert escalation policy UUID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Escalation policy name.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Escalation policy description. Defaults to empty string.",
				Validators:          []validator.String{stringvalidator.LengthAtMost(6144)},
			},
			"repeat": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Number of times the escalation policy repeats its steps if unacknowledged. Defaults to 0.",
				Validators:          []validator.Int64{int64validator.AtLeast(0)},
			},
		},
		Blocks: map[string]schema.Block{
			"step": schema.ListNestedBlock{
				MarkdownDescription: "Ordered list of escalation steps (0-indexed).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "GoAlert escalation policy step UUID.",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()},
						},
						"step_number": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Position of the step in the escalation sequence (0-indexed).",
							PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseNonNullStateForUnknown()},
						},
						"delay_minutes": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "Delay in minutes before escalating to the next step. Must be at least 1.",
							Validators:          []validator.Int64{int64validator.AtLeast(1)},
						},
						"user_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: "List of user IDs to target in this escalation step.",
						},
						"rotation_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							MarkdownDescription: "List of rotation IDs to target in this escalation step.",
						},
					},
					Blocks: map[string]schema.Block{
						"webhook_action": schema.ListNestedBlock{
							MarkdownDescription: "Webhook notification targets for this step.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"url": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Target webhook URL. Scheme http:// or https:// is required.",
										Validators: []validator.String{
											stringvalidator.RegexMatches(webhookURLPattern, "must be a valid URL starting with http:// or https://"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *escalationPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelFromEscalationPolicy(ep *client.EscalationPolicy, prior *escalationPolicyModel) escalationPolicyModel {
	m := escalationPolicyModel{
		ID:          types.StringValue(ep.ID),
		Name:        types.StringValue(ep.Name),
		Description: types.StringValue(ep.Description),
		Repeat:      types.Int64Value(ep.Repeat),
	}

	if len(ep.Steps) == 0 {
		if prior != nil && prior.Steps != nil && len(prior.Steps) == 0 {
			m.Steps = []stepModel{}
		} else {
			m.Steps = nil
		}
		return m
	}

	sortedSteps := make([]client.EscalationPolicyStep, len(ep.Steps))
	copy(sortedSteps, ep.Steps)
	sort.Slice(sortedSteps, func(i, j int) bool {
		return sortedSteps[i].StepNumber < sortedSteps[j].StepNumber
	})

	m.Steps = make([]stepModel, len(sortedSteps))
	for i, step := range sortedSteps {
		var priorStep *stepModel
		if prior != nil && i < len(prior.Steps) {
			priorStep = &prior.Steps[i]
		}

		sm := stepModel{
			ID:           types.StringValue(step.ID),
			StepNumber:   types.Int64Value(step.StepNumber),
			DelayMinutes: types.Int64Value(step.DelayMinutes),
		}

		var webhookActions []webhookActionModel
		var userIDs []types.String
		var rotationIDs []types.String
		for _, action := range step.Actions {
			switch action.Type {
			case "builtin-webhook":
				urlVal := action.Args["webhook_url"]
				if urlVal == "" {
					urlVal = action.Args["url"]
				}
				webhookActions = append(webhookActions, webhookActionModel{
					URL: types.StringValue(urlVal),
				})
			case "builtin-user":
				if uid, ok := action.Args["user_id"]; ok && uid != "" {
					userIDs = append(userIDs, types.StringValue(uid))
				}
			case "builtin-rotation":
				if rid, ok := action.Args["rotation_id"]; ok && rid != "" {
					rotationIDs = append(rotationIDs, types.StringValue(rid))
				}
			}
		}

		if len(webhookActions) == 0 {
			if priorStep != nil && priorStep.WebhookActions != nil && len(priorStep.WebhookActions) == 0 {
				sm.WebhookActions = []webhookActionModel{}
			} else {
				sm.WebhookActions = nil
			}
		} else {
			sm.WebhookActions = webhookActions
		}

		if len(userIDs) == 0 {
			if priorStep != nil && priorStep.UserIDs != nil && len(priorStep.UserIDs) == 0 {
				sm.UserIDs = []types.String{}
			} else {
				sm.UserIDs = nil
			}
		} else {
			sm.UserIDs = userIDs
		}

		if len(rotationIDs) == 0 {
			if priorStep != nil && priorStep.RotationIDs != nil && len(priorStep.RotationIDs) == 0 {
				sm.RotationIDs = []types.String{}
			} else {
				sm.RotationIDs = nil
			}
		} else {
			sm.RotationIDs = rotationIDs
		}

		m.Steps[i] = sm
	}

	return m
}

func (r *escalationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan escalationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	input := client.CreateEscalationPolicyInput{
		Name: name,
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		input.Description = &desc
	}
	if !plan.Repeat.IsNull() {
		repeat := plan.Repeat.ValueInt64()
		input.Repeat = &repeat
	}

	if len(plan.Steps) > 0 {
		input.Steps = make([]client.CreateEscalationPolicyStepInput, len(plan.Steps))
		for i, s := range plan.Steps {
			if len(s.WebhookActions) == 0 && len(s.UserIDs) == 0 && len(s.RotationIDs) == 0 {
				resp.Diagnostics.AddError("Invalid escalation policy step", fmt.Sprintf("step %d must specify at least one target (user_ids, rotation_ids) or webhook_action", i))
				return
			}
			stepInput := client.CreateEscalationPolicyStepInput{
				DelayMinutes: s.DelayMinutes.ValueInt64(),
			}
			var actions []client.DestinationInput
			for _, uid := range s.UserIDs {
				actions = append(actions, client.DestinationInput{
					Type: "builtin-user",
					Args: map[string]string{
						"user_id": uid.ValueString(),
					},
				})
			}
			for _, rid := range s.RotationIDs {
				actions = append(actions, client.DestinationInput{
					Type: "builtin-rotation",
					Args: map[string]string{
						"rotation_id": rid.ValueString(),
					},
				})
			}
			for _, act := range s.WebhookActions {
				actions = append(actions, client.DestinationInput{
					Type: "builtin-webhook",
					Args: map[string]string{
						"webhook_url": act.URL.ValueString(),
					},
				})
			}
			stepInput.Actions = actions
			input.Steps[i] = stepInput
		}
	}

	ep, err := r.client.CreateEscalationPolicy(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create escalation policy failed", err.Error()+". If the server committed before a connection failure, locate and import the escalation policy before retrying.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromEscalationPolicy(ep, &plan))...)
}

func (r *escalationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state escalationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.client.ReadEscalationPolicy(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read escalation policy failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromEscalationPolicy(ep, &state))...)
}

func actionsMatch(serverActions []client.Destination, planActions []webhookActionModel) bool {
	var serverWebhooks []string
	for _, a := range serverActions {
		if a.Type == "builtin-webhook" {
			u := a.Args["webhook_url"]
			if u == "" {
				u = a.Args["url"]
			}
			serverWebhooks = append(serverWebhooks, u)
		}
	}
	if len(serverWebhooks) != len(planActions) {
		return false
	}
	for i := range planActions {
		if serverWebhooks[i] != planActions[i].URL.ValueString() {
			return false
		}
	}
	return true
}

func stepMatches(serverStep client.EscalationPolicyStep, planStep stepModel) bool {
	if serverStep.DelayMinutes != planStep.DelayMinutes.ValueInt64() {
		return false
	}
	if !actionsMatch(serverStep.Actions, planStep.WebhookActions) {
		return false
	}
	var serverUsers []string
	var serverRotations []string
	for _, a := range serverStep.Actions {
		switch a.Type {
		case "builtin-user":
			if uid, ok := a.Args["user_id"]; ok && uid != "" {
				serverUsers = append(serverUsers, uid)
			}
		case "builtin-rotation":
			if rid, ok := a.Args["rotation_id"]; ok && rid != "" {
				serverRotations = append(serverRotations, rid)
			}
		}
	}

	if len(serverUsers) != len(planStep.UserIDs) {
		return false
	}
	for i := range planStep.UserIDs {
		if serverUsers[i] != planStep.UserIDs[i].ValueString() {
			return false
		}
	}

	if len(serverRotations) != len(planStep.RotationIDs) {
		return false
	}
	for i := range planStep.RotationIDs {
		if serverRotations[i] != planStep.RotationIDs[i].ValueString() {
			return false
		}
	}

	return true
}

func (r *escalationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan escalationPolicyModel
	var state escalationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID := plan.ID.ValueString()

	currentEP, err := r.client.ReadEscalationPolicy(ctx, policyID)
	if err != nil {
		resp.Diagnostics.AddError("Read escalation policy failed before update", err.Error())
		return
	}

	serverSteps := make(map[string]client.EscalationPolicyStep)
	for _, s := range currentEP.Steps {
		serverSteps[s.ID] = s
	}

	var stepIDs []string
	for i, pStep := range plan.Steps {
		if len(pStep.WebhookActions) == 0 && len(pStep.UserIDs) == 0 && len(pStep.RotationIDs) == 0 {
			resp.Diagnostics.AddError("Invalid escalation policy step", fmt.Sprintf("step %d must specify at least one target (user_ids, rotation_ids) or webhook_action", i))
			return
		}

		var actions []client.DestinationInput
		for _, uid := range pStep.UserIDs {
			actions = append(actions, client.DestinationInput{
				Type: "builtin-user",
				Args: map[string]string{
					"user_id": uid.ValueString(),
				},
			})
		}
		for _, rid := range pStep.RotationIDs {
			actions = append(actions, client.DestinationInput{
				Type: "builtin-rotation",
				Args: map[string]string{
					"rotation_id": rid.ValueString(),
				},
			})
		}
		for _, act := range pStep.WebhookActions {
			actions = append(actions, client.DestinationInput{
				Type: "builtin-webhook",
				Args: map[string]string{
					"webhook_url": act.URL.ValueString(),
				},
			})
		}

		stepID := ""
		if !pStep.ID.IsNull() && !pStep.ID.IsUnknown() && pStep.ID.ValueString() != "" {
			stepID = pStep.ID.ValueString()
		}

		if stepID != "" && serverSteps[stepID].ID != "" {
			s := serverSteps[stepID]
			delay := pStep.DelayMinutes.ValueInt64()
			if !stepMatches(s, pStep) {
				updateStepErr := r.client.UpdateEscalationPolicyStep(ctx, client.UpdateEscalationPolicyStepInput{
					ID:           stepID,
					DelayMinutes: &delay,
					Actions:      actions,
				})
				if updateStepErr != nil {
					resp.Diagnostics.AddError("Update escalation policy step failed", updateStepErr.Error())
					return
				}
			}
			stepIDs = append(stepIDs, stepID)
		} else {
			createdStep, createStepErr := r.client.CreateEscalationPolicyStep(ctx, client.CreateEscalationPolicyStepInput{
				EscalationPolicyID: &policyID,
				DelayMinutes:       pStep.DelayMinutes.ValueInt64(),
				Actions:            actions,
			})
			if createStepErr != nil {
				resp.Diagnostics.AddError("Create escalation policy step failed", createStepErr.Error())
				return
			}
			stepIDs = append(stepIDs, createdStep.ID)
		}
	}

	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	repeat := plan.Repeat.ValueInt64()

	updateInput := client.UpdateEscalationPolicyInput{
		ID:          policyID,
		Name:        &name,
		Description: &desc,
		Repeat:      &repeat,
		StepIDs:     &stepIDs,
	}

	if err := r.client.UpdateEscalationPolicy(ctx, updateInput); err != nil {
		resp.Diagnostics.AddError("Update escalation policy failed", err.Error())
		return
	}

	updatedEP, err := r.client.ReadEscalationPolicy(ctx, policyID)
	if err != nil {
		resp.Diagnostics.AddError("Read updated escalation policy failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromEscalationPolicy(updatedEP, &plan))...)
}

func (r *escalationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state escalationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEscalationPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		if strings.Contains(err.Error(), "currently in use") {
			resp.Diagnostics.AddError("Delete escalation policy failed", err.Error()+". The escalation policy cannot be deleted while it is referenced by a service. Remove or reassign any services using this policy before destroying it.")
			return
		}
		resp.Diagnostics.AddError("Delete escalation policy failed", err.Error())
	}
}

func (r *escalationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if !uuidPattern.MatchString(req.ID) {
		resp.Diagnostics.AddError("Invalid import ID", "Import using the escalation policy's lowercase UUID.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
