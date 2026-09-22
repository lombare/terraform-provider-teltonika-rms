package provider

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

// -------------------------------------------------------------------------------------
// tag / tags
// -------------------------------------------------------------------------------------

var tagObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"color":       types.StringType,
	"company_id":  types.Int64Type,
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func tagObject(t rms.Tag) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(tagObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(t.ID.String()),
		"name":        types.StringValue(t.Name),
		"description": types.StringValue(t.Description),
		"color":       types.StringValue(t.Color),
		"company_id":  numberToInt64(t.CompanyID),
		"created_at":  timestampToString(t.CreatedAt),
		"updated_at":  timestampToString(t.UpdatedAt),
	})
}

type tagDataSource struct{ client *rms.Client }

type tagDSModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewTagDataSource() datasource.DataSource { return &tagDataSource{} }

func (d *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *tagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *tagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A tag by id.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"color":       schema.StringAttribute{Computed: true},
			"company_id":  schema.Int64Attribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg tagDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := d.client.GetTag(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch tag", err.Error())
		return
	}
	cfg.Name = types.StringValue(t.Name)
	cfg.Description = types.StringValue(t.Description)
	cfg.Color = types.StringValue(t.Color)
	cfg.CompanyID = numberToInt64(t.CompanyID)
	cfg.CreatedAt = timestampToString(t.CreatedAt)
	cfg.UpdatedAt = timestampToString(t.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type tagsDataSource struct{ client *rms.Client }

type tagsModel struct {
	Tags types.List `tfsdk:"tags"`
}

func NewTagsDataSource() datasource.DataSource { return &tagsDataSource{} }

func (d *tagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *tagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *tagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All device tags.",
		Attributes: map[string]schema.Attribute{
			"tags": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: tagObjectAttrs}},
		},
	}
}

func (d *tagsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListTags(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list tags", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, t := range items {
		obj, diags := tagObject(t)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: tagObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, tagsModel{Tags: l})...)
}

// -------------------------------------------------------------------------------------
// alert / alerts
// -------------------------------------------------------------------------------------

var alertObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"device_id":   types.Int64Type,
	"name":        types.StringType,
	"type":        types.StringType,
	"severity":    types.StringType,
	"description": types.StringType,
	"message":     types.StringType,
	"status":      types.StringType,
	"created_at":  types.StringType,
}

func alertObject(a rms.Alert) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(alertObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(a.ID.String()),
		"device_id":   numberToInt64(a.DeviceID),
		"name":        types.StringValue(a.Name),
		"type":        types.StringValue(a.Type),
		"severity":    types.StringValue(a.Severity),
		"description": types.StringValue(a.Description),
		"message":     types.StringValue(a.Message),
		"status":      types.StringValue(a.Status),
		"created_at":  timestampToString(a.CreatedAt),
	})
}

type alertDataSource struct{ client *rms.Client }

type alertDSModel struct {
	ID          types.String `tfsdk:"id"`
	DeviceID    types.Int64  `tfsdk:"device_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Severity    types.String `tfsdk:"severity"`
	Description types.String `tfsdk:"description"`
	Message     types.String `tfsdk:"message"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewAlertDataSource() datasource.DataSource { return &alertDataSource{} }

func (d *alertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (d *alertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *alertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single alert.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"device_id":   schema.Int64Attribute{Computed: true},
			"name":        schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"severity":    schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"message":     schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *alertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg alertDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := d.client.GetAlert(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch alert", err.Error())
		return
	}
	cfg.DeviceID = numberToInt64(a.DeviceID)
	cfg.Name = types.StringValue(a.Name)
	cfg.Type = types.StringValue(a.Type)
	cfg.Severity = types.StringValue(a.Severity)
	cfg.Description = types.StringValue(a.Description)
	cfg.Message = types.StringValue(a.Message)
	cfg.Status = types.StringValue(a.Status)
	cfg.CreatedAt = timestampToString(a.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type alertsDataSource struct{ client *rms.Client }

type alertsModel struct {
	Alerts types.List `tfsdk:"alerts"`
}

func NewAlertsDataSource() datasource.DataSource { return &alertsDataSource{} }

func (d *alertsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerts"
}

func (d *alertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *alertsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All alerts.",
		Attributes: map[string]schema.Attribute{
			"alerts": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: alertObjectAttrs}},
		},
	}
}

func (d *alertsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListAlerts(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list alerts", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, a := range items {
		obj, diags := alertObject(a)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: alertObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, alertsModel{Alerts: l})...)
}

// -------------------------------------------------------------------------------------
// alert_configuration / alert_configurations
// -------------------------------------------------------------------------------------

var alertConfigObjectAttrs = map[string]attr.Type{
	"id":         types.StringType,
	"name":       types.StringType,
	"type":       types.StringType,
	"company_id": types.Int64Type,
	"enabled":    types.BoolType,
	"conditions": types.StringType,
	"actions":    types.StringType,
	"created_at": types.StringType,
	"updated_at": types.StringType,
}

func alertConfigObject(a rms.AlertConfiguration) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(alertConfigObjectAttrs, map[string]attr.Value{
		"id":         types.StringValue(a.ID.String()),
		"name":       types.StringValue(a.Name),
		"type":       types.StringValue(a.Type),
		"company_id": numberToInt64(a.CompanyID),
		"enabled":    types.BoolValue(a.Enabled),
		"conditions": rawToString(a.Conditions),
		"actions":    rawToString(a.Actions),
		"created_at": timestampToString(a.CreatedAt),
		"updated_at": timestampToString(a.UpdatedAt),
	})
}

type alertConfigDataSource struct{ client *rms.Client }

type alertConfigDSModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	CompanyID  types.Int64  `tfsdk:"company_id"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Conditions types.String `tfsdk:"conditions"`
	Actions    types.String `tfsdk:"actions"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func NewAlertConfigurationDataSource() datasource.DataSource { return &alertConfigDataSource{} }

func (d *alertConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_configuration"
}

func (d *alertConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *alertConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single alert configuration by id.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Required: true},
			"name":       schema.StringAttribute{Computed: true},
			"type":       schema.StringAttribute{Computed: true},
			"company_id": schema.Int64Attribute{Computed: true},
			"enabled":    schema.BoolAttribute{Computed: true},
			"conditions": schema.StringAttribute{Computed: true, Description: "Raw JSON."},
			"actions":    schema.StringAttribute{Computed: true, Description: "Raw JSON."},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *alertConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg alertConfigDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := d.client.GetAlertConfiguration(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch alert configuration", err.Error())
		return
	}
	cfg.Name = types.StringValue(a.Name)
	cfg.Type = types.StringValue(a.Type)
	cfg.CompanyID = numberToInt64(a.CompanyID)
	cfg.Enabled = types.BoolValue(a.Enabled)
	cfg.Conditions = rawToString(a.Conditions)
	cfg.Actions = rawToString(a.Actions)
	cfg.CreatedAt = timestampToString(a.CreatedAt)
	cfg.UpdatedAt = timestampToString(a.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type alertConfigsDataSource struct{ client *rms.Client }

type alertConfigsModel struct {
	AlertConfigurations types.List `tfsdk:"alert_configurations"`
}

func NewAlertConfigurationsDataSource() datasource.DataSource { return &alertConfigsDataSource{} }

func (d *alertConfigsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_configurations"
}

func (d *alertConfigsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *alertConfigsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All alert configurations.",
		Attributes: map[string]schema.Attribute{
			"alert_configurations": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: alertConfigObjectAttrs}},
		},
	}
}

func (d *alertConfigsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListAlertConfigurations(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list alert configurations", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, a := range items {
		obj, diags := alertConfigObject(a)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: alertConfigObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, alertConfigsModel{AlertConfigurations: l})...)
}

// -------------------------------------------------------------------------------------
// email_configuration / email_configurations
// -------------------------------------------------------------------------------------

var emailConfigObjectAttrs = map[string]attr.Type{
	"id":         types.StringType,
	"name":       types.StringType,
	"host":       types.StringType,
	"port":       types.Int64Type,
	"email":      types.StringType,
	"username":   types.StringType,
	"company_id": types.Int64Type,
	"created_at": types.StringType,
	"updated_at": types.StringType,
}

func emailConfigObject(e rms.EmailConfiguration) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(emailConfigObjectAttrs, map[string]attr.Value{
		"id":         types.StringValue(e.ID.String()),
		"name":       types.StringValue(e.Name),
		"host":       types.StringValue(e.Host),
		"port":       numberToInt64(e.Port),
		"email":      types.StringValue(e.Email),
		"username":   types.StringValue(e.Username),
		"company_id": numberToInt64(e.CompanyID),
		"created_at": timestampToString(e.CreatedAt),
		"updated_at": timestampToString(e.UpdatedAt),
	})
}

type emailConfigDataSource struct{ client *rms.Client }

type emailConfigDSModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int64  `tfsdk:"port"`
	Email     types.String `tfsdk:"email"`
	Username  types.String `tfsdk:"username"`
	CompanyID types.Int64  `tfsdk:"company_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewEmailConfigurationDataSource() datasource.DataSource { return &emailConfigDataSource{} }

func (d *emailConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_configuration"
}

func (d *emailConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *emailConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single email configuration by id.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Required: true},
			"name":       schema.StringAttribute{Computed: true},
			"host":       schema.StringAttribute{Computed: true},
			"port":       schema.Int64Attribute{Computed: true},
			"email":      schema.StringAttribute{Computed: true},
			"username":   schema.StringAttribute{Computed: true},
			"company_id": schema.Int64Attribute{Computed: true},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *emailConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg emailConfigDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	e, err := d.client.GetEmailConfiguration(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch email configuration", err.Error())
		return
	}
	cfg.Name = types.StringValue(e.Name)
	cfg.Host = types.StringValue(e.Host)
	cfg.Port = numberToInt64(e.Port)
	cfg.Email = types.StringValue(e.Email)
	cfg.Username = types.StringValue(e.Username)
	cfg.CompanyID = numberToInt64(e.CompanyID)
	cfg.CreatedAt = timestampToString(e.CreatedAt)
	cfg.UpdatedAt = timestampToString(e.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type emailConfigsDataSource struct{ client *rms.Client }

type emailConfigsModel struct {
	EmailConfigurations types.List `tfsdk:"email_configurations"`
}

func NewEmailConfigurationsDataSource() datasource.DataSource { return &emailConfigsDataSource{} }

func (d *emailConfigsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_configurations"
}

func (d *emailConfigsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *emailConfigsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All email configurations.",
		Attributes: map[string]schema.Attribute{
			"email_configurations": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: emailConfigObjectAttrs}},
		},
	}
}

func (d *emailConfigsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListEmailConfigurations(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list email configurations", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, e := range items {
		obj, diags := emailConfigObject(e)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: emailConfigObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, emailConfigsModel{EmailConfigurations: l})...)
}

// -------------------------------------------------------------------------------------
// role / roles / role_permissions
// -------------------------------------------------------------------------------------

var roleObjectAttrs = map[string]attr.Type{
	"id":             types.StringType,
	"title":          types.StringType,
	"description":    types.StringType,
	"company_ids":    types.ListType{ElemType: types.Int64Type},
	"permission_ids": types.ListType{ElemType: types.Int64Type},
	"created_at":     types.StringType,
	"updated_at":     types.StringType,
}

func roleObject(r rms.Role) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(roleObjectAttrs, map[string]attr.Value{
		"id":             types.StringValue(r.ID.String()),
		"title":          types.StringValue(r.Title),
		"description":    types.StringValue(r.Description),
		"company_ids":    listFromInt64sNumbers(r.CompanyID),
		"permission_ids": listFromInt64sNumbers(r.Permissions),
		"created_at":     timestampToString(r.CreatedAt),
		"updated_at":     timestampToString(r.UpdatedAt),
	})
}

type roleDataSource struct{ client *rms.Client }

type roleDSModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	CompanyIDs  types.List   `tfsdk:"company_ids"`
	Permissions types.List   `tfsdk:"permission_ids"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewRoleDataSource() datasource.DataSource { return &roleDataSource{} }

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single role by id.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Required: true},
			"title":          schema.StringAttribute{Computed: true},
			"description":    schema.StringAttribute{Computed: true},
			"company_ids":    schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"permission_ids": schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"created_at":     schema.StringAttribute{Computed: true},
			"updated_at":     schema.StringAttribute{Computed: true},
		},
	}
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg roleDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r, err := d.client.GetRole(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch role", err.Error())
		return
	}
	cfg.Title = types.StringValue(r.Title)
	cfg.Description = types.StringValue(r.Description)
	cfg.CompanyIDs = listFromInt64sNumbers(r.CompanyID)
	cfg.Permissions = listFromInt64sNumbers(r.Permissions)
	cfg.CreatedAt = timestampToString(r.CreatedAt)
	cfg.UpdatedAt = timestampToString(r.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type rolesDataSource struct{ client *rms.Client }

type rolesModel struct {
	Roles types.List `tfsdk:"roles"`
}

func NewRolesDataSource() datasource.DataSource { return &rolesDataSource{} }

func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All roles.",
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: roleObjectAttrs}},
		},
	}
}

func (d *rolesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListRoles(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list roles", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, r := range items {
		obj, diags := roleObject(r)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: roleObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, rolesModel{Roles: l})...)
}

var permissionObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"title":       types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"category":    types.StringType,
}

func permissionObject(p rms.Permission) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(permissionObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(p.ID.String()),
		"title":       types.StringValue(p.Title),
		"name":        types.StringValue(p.Name),
		"description": types.StringValue(p.Description),
		"category":    types.StringValue(p.Category),
	})
}

type rolePermissionsDataSource struct{ client *rms.Client }

type rolePermissionsModel struct {
	RoleID      types.String `tfsdk:"role_id"`
	Permissions types.List   `tfsdk:"permissions"`
}

func NewRolePermissionsDataSource() datasource.DataSource { return &rolePermissionsDataSource{} }

func (d *rolePermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_permissions"
}

func (d *rolePermissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *rolePermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Permission catalogue — for a specific role when `role_id` is set, otherwise all permissions the token is entitled to (`/roles/permissions`).",
		Attributes: map[string]schema.Attribute{
			"role_id":     schema.StringAttribute{Optional: true},
			"permissions": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: permissionObjectAttrs}},
		},
	}
}

func (d *rolePermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg rolePermissionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var items []rms.Permission
	var err error
	if !cfg.RoleID.IsNull() && !cfg.RoleID.IsUnknown() && cfg.RoleID.ValueString() != "" {
		items, err = d.client.RolePermissions(ctx, cfg.RoleID.ValueString())
	} else {
		items, err = d.client.ListPermissions(ctx)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch permissions", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, p := range items {
		obj, diags := permissionObject(p)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: permissionObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	cfg.Permissions = l
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

// -------------------------------------------------------------------------------------
// automation / automations
// -------------------------------------------------------------------------------------

var automationObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"company_id":  types.Int64Type,
	"enabled":     types.BoolType,
	"trigger":     types.StringType,
	"conditions":  types.StringType,
	"actions":     types.StringType,
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func automationObject(a rms.Automation) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(automationObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(a.ID.String()),
		"name":        types.StringValue(a.Name),
		"description": types.StringValue(a.Description),
		"company_id":  numberToInt64(a.CompanyID),
		"enabled":     types.BoolValue(a.Enabled),
		"trigger":     rawToString(a.Trigger),
		"conditions":  rawToString(a.Conditions),
		"actions":     rawToString(a.Actions),
		"created_at":  timestampToString(a.CreatedAt),
		"updated_at":  timestampToString(a.UpdatedAt),
	})
}

type automationDataSource struct{ client *rms.Client }

type automationDSModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Trigger     types.String `tfsdk:"trigger"`
	Conditions  types.String `tfsdk:"conditions"`
	Actions     types.String `tfsdk:"actions"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewAutomationDataSource() datasource.DataSource { return &automationDataSource{} }

func (d *automationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

func (d *automationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *automationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single automation by id.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"company_id":  schema.Int64Attribute{Computed: true},
			"enabled":     schema.BoolAttribute{Computed: true},
			"trigger":     schema.StringAttribute{Computed: true, Description: "Raw JSON."},
			"conditions":  schema.StringAttribute{Computed: true, Description: "Raw JSON."},
			"actions":     schema.StringAttribute{Computed: true, Description: "Raw JSON."},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *automationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg automationDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := d.client.GetAutomation(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch automation", err.Error())
		return
	}
	cfg.Name = types.StringValue(a.Name)
	cfg.Description = types.StringValue(a.Description)
	cfg.CompanyID = numberToInt64(a.CompanyID)
	cfg.Enabled = types.BoolValue(a.Enabled)
	cfg.Trigger = rawToString(a.Trigger)
	cfg.Conditions = rawToString(a.Conditions)
	cfg.Actions = rawToString(a.Actions)
	cfg.CreatedAt = timestampToString(a.CreatedAt)
	cfg.UpdatedAt = timestampToString(a.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type automationsDataSource struct{ client *rms.Client }

type automationsModel struct {
	Automations types.List `tfsdk:"automations"`
}

func NewAutomationsDataSource() datasource.DataSource { return &automationsDataSource{} }

func (d *automationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automations"
}

func (d *automationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *automationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All automations.",
		Attributes: map[string]schema.Attribute{
			"automations": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: automationObjectAttrs}},
		},
	}
}

func (d *automationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListAutomations(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list automations", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, a := range items {
		obj, diags := automationObject(a)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: automationObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, automationsModel{Automations: l})...)
}

// -------------------------------------------------------------------------------------
// vpn_hub / vpn_hubs
// -------------------------------------------------------------------------------------

var vpnHubObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"company_id":  types.Int64Type,
	"hub_zone":    types.StringType,
	"vpn_type":    types.StringType,
	"enabled":     types.BoolType,
	"tag_ids":     types.ListType{ElemType: types.Int64Type},
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func vpnHubObject(h rms.VPNHub) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(vpnHubObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(h.ID.String()),
		"name":        types.StringValue(h.Name),
		"description": types.StringValue(h.Description),
		"company_id":  numberToInt64(h.CompanyID),
		"hub_zone":    types.StringValue(h.HubZone),
		"vpn_type":    types.StringValue(h.VPNType),
		"enabled":     types.BoolValue(h.Enabled),
		"tag_ids":     listFromInt64sNumbers(h.TagIDs),
		"created_at":  timestampToString(h.CreatedAt),
		"updated_at":  timestampToString(h.UpdatedAt),
	})
}

type vpnHubDataSource struct{ client *rms.Client }

type vpnHubDSModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	HubZone     types.String `tfsdk:"hub_zone"`
	VPNType     types.String `tfsdk:"vpn_type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	TagIDs      types.List   `tfsdk:"tag_ids"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewVPNHubDataSource() datasource.DataSource { return &vpnHubDataSource{} }

func (d *vpnHubDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_hub"
}

func (d *vpnHubDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *vpnHubDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single VPN hub by id.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"company_id":  schema.Int64Attribute{Computed: true},
			"hub_zone":    schema.StringAttribute{Computed: true},
			"vpn_type":    schema.StringAttribute{Computed: true},
			"enabled":     schema.BoolAttribute{Computed: true},
			"tag_ids":     schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *vpnHubDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg vpnHubDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	h, err := d.client.GetVPNHub(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch VPN hub", err.Error())
		return
	}
	cfg.Name = types.StringValue(h.Name)
	cfg.Description = types.StringValue(h.Description)
	cfg.CompanyID = numberToInt64(h.CompanyID)
	cfg.HubZone = types.StringValue(h.HubZone)
	cfg.VPNType = types.StringValue(h.VPNType)
	cfg.Enabled = types.BoolValue(h.Enabled)
	cfg.TagIDs = listFromInt64sNumbers(h.TagIDs)
	cfg.CreatedAt = timestampToString(h.CreatedAt)
	cfg.UpdatedAt = timestampToString(h.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type vpnHubsDataSource struct{ client *rms.Client }

type vpnHubsModel struct {
	Hubs types.List `tfsdk:"hubs"`
}

func NewVPNHubsDataSource() datasource.DataSource { return &vpnHubsDataSource{} }

func (d *vpnHubsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_hubs"
}

func (d *vpnHubsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *vpnHubsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All VPN hubs.",
		Attributes: map[string]schema.Attribute{
			"hubs": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: vpnHubObjectAttrs}},
		},
	}
}

func (d *vpnHubsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListVPNHubs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list VPN hubs", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, h := range items {
		obj, diags := vpnHubObject(h)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: vpnHubObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, vpnHubsModel{Hubs: l})...)
}

// -------------------------------------------------------------------------------------
// files
// -------------------------------------------------------------------------------------

var fileObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"type":        types.StringType,
	"description": types.StringType,
	"company_id":  types.Int64Type,
	"size":        types.Int64Type,
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func fileObject(f rms.File) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(fileObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(f.ID.String()),
		"name":        types.StringValue(f.Name),
		"type":        types.StringValue(f.Type),
		"description": types.StringValue(f.Description),
		"company_id":  numberToInt64(f.CompanyID),
		"size":        numberToInt64(f.Size),
		"created_at":  timestampToString(f.CreatedAt),
		"updated_at":  timestampToString(f.UpdatedAt),
	})
}

type filesDataSource struct{ client *rms.Client }

type filesModel struct {
	Files types.List `tfsdk:"files"`
}

func NewFilesDataSource() datasource.DataSource { return &filesDataSource{} }

func (d *filesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_files"
}

func (d *filesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *filesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All files uploaded to RMS.",
		Attributes: map[string]schema.Attribute{
			"files": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: fileObjectAttrs}},
		},
	}
}

func (d *filesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListFiles(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list files", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, f := range items {
		obj, diags := fileObject(f)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: fileObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, filesModel{Files: l})...)
}

// -------------------------------------------------------------------------------------
// hotspots
// -------------------------------------------------------------------------------------

var hotspotObjectAttrs = map[string]attr.Type{
	"id":        types.StringType,
	"name":      types.StringType,
	"ssid":      types.StringType,
	"device_id": types.Int64Type,
	"enabled":   types.BoolType,
}

func hotspotObject(h rms.Hotspot) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(hotspotObjectAttrs, map[string]attr.Value{
		"id":        types.StringValue(h.ID.String()),
		"name":      types.StringValue(h.Name),
		"ssid":      types.StringValue(h.SSID),
		"device_id": numberToInt64(h.DeviceID),
		"enabled":   types.BoolValue(h.Enabled),
	})
}

type hotspotsDataSource struct{ client *rms.Client }

type hotspotsModel struct {
	Hotspots types.List `tfsdk:"hotspots"`
}

func NewHotspotsDataSource() datasource.DataSource { return &hotspotsDataSource{} }

func (d *hotspotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hotspots"
}

func (d *hotspotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *hotspotsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All hotspots visible via `/hotspots`.",
		Attributes: map[string]schema.Attribute{
			"hotspots": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: hotspotObjectAttrs}},
		},
	}
}

func (d *hotspotsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.Hotspots(ctx, url.Values{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list hotspots", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, h := range items {
		obj, diags := hotspotObject(h)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: hotspotObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, hotspotsModel{Hotspots: l})...)
}

// -------------------------------------------------------------------------------------
// data_collect_configs
// -------------------------------------------------------------------------------------

var dataCollectConfigObjectAttrs = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"company_id":  types.Int64Type,
	"interval":    types.Int64Type,
	"enabled":     types.BoolType,
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func dataCollectConfigObject(c rms.DataCollectConfig) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(dataCollectConfigObjectAttrs, map[string]attr.Value{
		"id":          types.StringValue(c.ID.String()),
		"name":        types.StringValue(c.Name),
		"description": types.StringValue(c.Description),
		"company_id":  numberToInt64(c.CompanyID),
		"interval":    numberToInt64(c.Interval),
		"enabled":     types.BoolValue(c.Enabled),
		"created_at":  timestampToString(c.CreatedAt),
		"updated_at":  timestampToString(c.UpdatedAt),
	})
}

type dataCollectConfigsDataSource struct{ client *rms.Client }

type dataCollectConfigsModel struct {
	Configs types.List `tfsdk:"configs"`
}

func NewDataCollectConfigsDataSource() datasource.DataSource { return &dataCollectConfigsDataSource{} }

func (d *dataCollectConfigsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_collect_configs"
}

func (d *dataCollectConfigsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *dataCollectConfigsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All data-collect configurations.",
		Attributes: map[string]schema.Attribute{
			"configs": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: dataCollectConfigObjectAttrs}},
		},
	}
}

func (d *dataCollectConfigsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListDataCollectConfigs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list data-collect configurations", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, c := range items {
		obj, diags := dataCollectConfigObject(c)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: dataCollectConfigObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, dataCollectConfigsModel{Configs: l})...)
}

// -------------------------------------------------------------------------------------
// credits_summary
// -------------------------------------------------------------------------------------

type creditsSummaryDataSource struct{ client *rms.Client }

type creditsSummaryModel struct {
	JSON types.String `tfsdk:"json"`
}

func NewCreditsSummaryDataSource() datasource.DataSource { return &creditsSummaryDataSource{} }

func (d *creditsSummaryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credits_summary"
}

func (d *creditsSummaryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *creditsSummaryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Credits summary (`/credits/summary`) returned as raw JSON so downstream automation can pick out the fields it needs.",
		Attributes: map[string]schema.Attribute{
			"json": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *creditsSummaryDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	raw, err := d.client.CreditsSummary(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch credits summary", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, creditsSummaryModel{JSON: rawToString(raw)})...)
}
