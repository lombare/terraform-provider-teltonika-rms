package provider

import (
	"context"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

// ---- current_user ----

type currentUserDataSource struct{ client *rms.Client }

type currentUserModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Role      types.String `tfsdk:"role"`
	CompanyID types.Int64  `tfsdk:"company_id"`
}

func NewCurrentUserDataSource() datasource.DataSource { return &currentUserDataSource{} }

func (d *currentUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (d *currentUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *currentUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The user represented by the configured token (`GET /user`).",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"email":      schema.StringAttribute{Computed: true},
			"first_name": schema.StringAttribute{Computed: true},
			"last_name":  schema.StringAttribute{Computed: true},
			"role":       schema.StringAttribute{Computed: true},
			"company_id": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *currentUserDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	u, err := d.client.CurrentUser(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch current user", err.Error())
		return
	}
	out := currentUserModel{
		ID:        types.StringValue(u.ID.String()),
		Email:     types.StringValue(u.Email),
		FirstName: types.StringValue(u.FirstName),
		LastName:  types.StringValue(u.LastName),
		Role:      types.StringValue(u.Role),
		CompanyID: numberToInt64(u.CompanyID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}

// ---- user (by id) ----

type userDataSource struct{ client *rms.Client }

type userDSModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Role      types.String `tfsdk:"role"`
	CompanyID types.Int64  `tfsdk:"company_id"`
}

func NewUserDataSource() datasource.DataSource { return &userDataSource{} }

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single RMS user (`GET /users/{id}`).",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Required: true},
			"email":      schema.StringAttribute{Computed: true},
			"first_name": schema.StringAttribute{Computed: true},
			"last_name":  schema.StringAttribute{Computed: true},
			"role":       schema.StringAttribute{Computed: true},
			"company_id": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg userDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := d.client.GetUser(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch user", err.Error())
		return
	}
	cfg.Email = types.StringValue(u.Email)
	cfg.FirstName = types.StringValue(u.FirstName)
	cfg.LastName = types.StringValue(u.LastName)
	cfg.Role = types.StringValue(u.Role)
	cfg.CompanyID = numberToInt64(u.CompanyID)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

// ---- users (list) ----

type usersDataSource struct{ client *rms.Client }

type usersModel struct {
	CompanyID types.Int64  `tfsdk:"company_id"`
	Search    types.String `tfsdk:"search"`
	Users     types.List   `tfsdk:"users"`
}

var userObjectAttrs = map[string]attr.Type{
	"id":         types.StringType,
	"email":      types.StringType,
	"first_name": types.StringType,
	"last_name":  types.StringType,
	"role":       types.StringType,
	"company_id": types.Int64Type,
}

func NewUsersDataSource() datasource.DataSource { return &usersDataSource{} }

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All RMS users, optionally filtered by company and search string.",
		Attributes: map[string]schema.Attribute{
			"company_id": schema.Int64Attribute{Optional: true},
			"search":     schema.StringAttribute{Optional: true},
			"users": schema.ListAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: userObjectAttrs},
			},
		},
	}
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg usersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	q := url.Values{}
	if !cfg.CompanyID.IsNull() && !cfg.CompanyID.IsUnknown() {
		q.Set("company_id", strconv.FormatInt(cfg.CompanyID.ValueInt64(), 10))
	}
	if s := stringOr(cfg.Search, ""); s != "" {
		q.Set("q", s)
	}
	users, err := d.client.ListUsers(ctx, q)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list users", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(users))
	for _, u := range users {
		obj, diags := types.ObjectValue(userObjectAttrs, map[string]attr.Value{
			"id":         types.StringValue(u.ID.String()),
			"email":      types.StringValue(u.Email),
			"first_name": types.StringValue(u.FirstName),
			"last_name":  types.StringValue(u.LastName),
			"role":       types.StringValue(u.Role),
			"company_id": numberToInt64(u.CompanyID),
		})
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: userObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	cfg.Users = list
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}
