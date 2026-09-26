package provider

import (
	"context"
	"errors"
	"strings"

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

type integrationKeyResource struct {
	client *client.Client
}

type integrationKeyModel struct {
	ID        types.String `tfsdk:"id"`
	ServiceID types.String `tfsdk:"service_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Href      types.String `tfsdk:"href"`
}

var (
	_ resource.Resource                = &integrationKeyResource{}
	_ resource.ResourceWithConfigure   = &integrationKeyResource{}
	_ resource.ResourceWithImportState = &integrationKeyResource{}
)

func NewIntegrationKeyResource() resource.Resource {
	return &integrationKeyResource{}
}

func (r *integrationKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_key"
}

func (r *integrationKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a GoAlert integration key providing incoming webhook endpoints for services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert integration key UUID.",
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
				MarkdownDescription: "Integration key name. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("grafana"),
				MarkdownDescription: "Integration key type (generic, grafana, site24x7, prometheusAlertmanager, email, universal). Defaults to 'grafana'. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("generic", "grafana", "site24x7", "prometheusAlertmanager", "email", "universal"),
				},
			},
			"href": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Complete incoming webhook URL containing the authentication token.",
			},
		},
	}
}

func (r *integrationKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelFromIntegrationKey(ik *client.IntegrationKey) integrationKeyModel {
	return integrationKeyModel{
		ID:        types.StringValue(ik.ID),
		ServiceID: types.StringValue(ik.ServiceID),
		Name:      types.StringValue(ik.Name),
		Type:      types.StringValue(ik.Type),
		Href:      types.StringValue(ik.Href),
	}
}

func (r *integrationKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := client.CreateIntegrationKeyInput{
		ServiceID: plan.ServiceID.ValueString(),
		Name:      plan.Name.ValueString(),
		Type:      plan.Type.ValueString(),
	}

	ik, err := r.client.CreateIntegrationKey(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create integration key failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromIntegrationKey(ik))...)
}

func (r *integrationKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ik, err := r.client.ReadIntegrationKey(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read integration key failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromIntegrationKey(ik))...)
}

func (r *integrationKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"GoAlert integration keys are strictly immutable. Any modification to attributes forces resource replacement.",
	)
}

func (r *integrationKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state integrationKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteIntegrationKey(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete integration key failed", err.Error())
	}
}

func parseImportID(id string) (string, error) {
	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		id = parts[len(parts)-1]
	}
	if !uuidPattern.MatchString(id) {
		return "", errors.New("import using the integration key's lowercase UUID or compound format <service_id>/<key_id>")
	}
	return id, nil
}

func (r *integrationKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	req.ID = id
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
