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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

var userUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type userResource struct {
	client *client.Client
}

type userModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Email    types.String `tfsdk:"email"`
	Role     types.String `tfsdk:"role"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a human operator user account in GoAlert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "GoAlert user UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Full name of the user.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email address of the user.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
				},
			},
			"role": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("user"),
				MarkdownDescription: "Permission role of the user (`user` or `admin`). Defaults to `user`.",
				Validators: []validator.String{
					stringvalidator.OneOf("user", "admin"),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Basic authentication login username. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial password for basic authentication. Changing this forces recreation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pwd := ""
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		pwd = plan.Password.ValueString()
	}

	u, err := r.client.CreateUser(ctx, client.CreateUserInput{
		Name:     plan.Name.ValueString(),
		Email:    plan.Email.ValueString(),
		Role:     plan.Role.ValueString(),
		Username: plan.Username.ValueString(),
		Password: pwd,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating GoAlert user", err.Error())
		return
	}

	plan.ID = types.StringValue(u.ID)
	plan.Name = types.StringValue(u.Name)
	plan.Email = types.StringValue(u.Email)
	plan.Role = types.StringValue(u.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.ReadUser(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GoAlert user", err.Error())
		return
	}

	state.Name = types.StringValue(u.Name)
	state.Email = types.StringValue(u.Email)
	state.Role = types.StringValue(u.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateUser(ctx, client.UpdateUserInput{
		ID:    plan.ID.ValueString(),
		Name:  plan.Name.ValueString(),
		Email: plan.Email.ValueString(),
		Role:  plan.Role.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating GoAlert user", err.Error())
		return
	}

	u, err := r.client.ReadUser(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error refreshing GoAlert user after update", err.Error())
		return
	}

	plan.Name = types.StringValue(u.Name)
	plan.Email = types.StringValue(u.Email)
	plan.Role = types.StringValue(u.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting GoAlert user", err.Error())
		return
	}
}

func parseUserImportID(raw string) (userID, username string, err error) {
	parts := strings.Split(raw, "/")
	if len(parts) == 2 && userUUIDPattern.MatchString(parts[0]) && len(parts[1]) > 0 {
		return parts[0], parts[1], nil
	}
	if len(parts) == 1 && userUUIDPattern.MatchString(parts[0]) {
		return parts[0], "", nil
	}
	return "", "", errors.New("import requires user UUID or compound <user_id>/<username>")
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	userID, username, err := parseUserImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), userID)...)
	if username != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), username)...)
	}
}
