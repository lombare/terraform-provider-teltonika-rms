package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

type emailConfigurationResource struct{ client *rms.Client }

type emailConfigurationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int64  `tfsdk:"port"`
	Email     types.String `tfsdk:"email"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewEmailConfigurationResource() resource.Resource { return &emailConfigurationResource{} }

func (r *emailConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_configuration"
}

func (r *emailConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *emailConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SMTP email configuration used by RMS alerting (`/email-configurations`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":     schema.StringAttribute{Required: true},
			"host":     schema.StringAttribute{Required: true},
			"port":     schema.Int64Attribute{Required: true},
			"email":    schema.StringAttribute{Required: true},
			"username": schema.StringAttribute{Required: true},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "SMTP password. Never surfaced by the RMS API on read; the value is kept in state as-configured.",
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *emailConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateEmailConfiguration(ctx, rms.EmailConfigurationCreate{
		Name:     plan.Name.ValueString(),
		Host:     plan.Host.ValueString(),
		Port:     int(plan.Port.ValueInt64()),
		Email:    plan.Email.ValueString(),
		Username: plan.Username.ValueString(),
		Password: plan.Password.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create email configuration", err.Error())
		return
	}
	writeEmailConfiguration(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetEmailConfiguration(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "email configuration", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeEmailConfiguration(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *emailConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state emailConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateEmailConfiguration(ctx, state.ID.ValueString(), rms.EmailConfigurationUpdate{
		Name:     plan.Name.ValueString(),
		Host:     plan.Host.ValueString(),
		Port:     int(plan.Port.ValueInt64()),
		Email:    plan.Email.ValueString(),
		Username: plan.Username.ValueString(),
		Password: plan.Password.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Failed to update email configuration", err.Error())
		return
	}
	fresh, err := r.client.GetEmailConfiguration(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read email configuration after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeEmailConfiguration(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEmailConfiguration(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete email configuration", err.Error())
	}
}

func (r *emailConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeEmailConfiguration(m *emailConfigurationResourceModel, c *rms.EmailConfiguration) {
	m.ID = types.StringValue(c.ID.String())
	m.Name = types.StringValue(c.Name)
	m.Host = types.StringValue(c.Host)
	m.Port = numberToInt64(c.Port)
	m.Email = types.StringValue(c.Email)
	m.Username = types.StringValue(c.Username)
	m.CreatedAt = timestampToString(c.CreatedAt)
	m.UpdatedAt = timestampToString(c.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*emailConfigurationResource)(nil)
