package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

type userDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
	Role  types.String `tfsdk:"role"`
}

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lookup an existing GoAlert user by UUID, exact name, or email.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "UUID of the GoAlert user to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name"), path.MatchRoot("email")),
					stringvalidator.RegexMatches(uuidPattern, "must be a lowercase UUID"),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Full name of the user to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name"), path.MatchRoot("email")),
				},
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Email address of the user to lookup.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name"), path.MatchRoot("email")),
				},
			},
			"role": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Permission role of the user (`user` or `admin`).",
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var found *client.User

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		u, err := d.client.ReadUser(ctx, config.ID.ValueString())
		if err != nil {
			if errors.Is(err, client.ErrNotFound) {
				resp.Diagnostics.AddError("User not found", fmt.Sprintf("No GoAlert user found with ID: %s", config.ID.ValueString()))
				return
			}
			resp.Diagnostics.AddError("Error reading GoAlert user", err.Error())
			return
		}
		found = u
	} else {
		searchTerm := ""
		isEmailSearch := false
		if !config.Email.IsNull() && config.Email.ValueString() != "" {
			searchTerm = config.Email.ValueString()
			isEmailSearch = true
		} else {
			searchTerm = config.Name.ValueString()
		}

		var searchArg string
		if isEmailSearch {
			parts := strings.Split(searchTerm, "@")
			if len(parts) > 0 && parts[0] != "" {
				searchArg = parts[0]
			}
		} else {
			searchArg = searchTerm
		}

		users, err := d.client.SearchUsers(ctx, searchArg)
		if err != nil {
			resp.Diagnostics.AddError("Error searching GoAlert users", err.Error())
			return
		}

		var matches []client.User
		for _, u := range users {
			if isEmailSearch {
				if strings.EqualFold(u.Email, searchTerm) {
					matches = append(matches, u)
				}
			} else {
				if strings.EqualFold(u.Name, searchTerm) {
					matches = append(matches, u)
				}
			}
		}

		if isEmailSearch && len(matches) == 0 {
			allUsers, err := d.client.SearchUsers(ctx, "")
			if err == nil {
				for _, u := range allUsers {
					if strings.EqualFold(u.Email, searchTerm) {
						matches = append(matches, u)
					}
				}
			}
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError("User not found", fmt.Sprintf("No GoAlert user found matching search: %q", searchTerm))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple users found", fmt.Sprintf("Found %d GoAlert users matching %q; specify an exact ID instead", len(matches), searchTerm))
			return
		}
		found = &matches[0]
	}

	config.ID = types.StringValue(found.ID)
	config.Name = types.StringValue(found.Name)
	config.Email = types.StringValue(found.Email)
	config.Role = types.StringValue(found.Role)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
