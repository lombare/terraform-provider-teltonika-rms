package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

// alertConfigurationResource — the RMS alert configuration schema varies a lot by alert
// type (SIM, data usage, signal strength, connection state, custom, …). Rather than
// enumerate every branch, the resource accepts a raw JSON body (as `payload`) that matches
// the OpenAPI `alert_config_add`/`alert_config_update` shape.
type alertConfigurationResource struct{ client *rms.Client }

type alertConfigurationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Payload   types.String `tfsdk:"payload"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewAlertConfigurationResource() resource.Resource { return &alertConfigurationResource{} }

func (r *alertConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_configuration"
}

func (r *alertConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *alertConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Alert configuration (`/alerts-configurations`). The payload varies wildly by alert type; supply the raw JSON body that matches the RMS OpenAPI shape.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"payload": schema.StringAttribute{
				Required:    true,
				Description: "Raw JSON string of a single alert configuration entry (the object under `data[0]` per the OpenAPI). Round-tripped verbatim on update.",
			},
			"name":       schema.StringAttribute{Computed: true},
			"type":       schema.StringAttribute{Computed: true},
			"enabled":    schema.BoolAttribute{Computed: true},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *alertConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var payload json.RawMessage
	if err := json.Unmarshal([]byte(plan.Payload.ValueString()), &payload); err != nil {
		resp.Diagnostics.AddError("Invalid `payload` JSON", err.Error())
		return
	}
	id, err := r.client.CreateAlertConfiguration(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create alert configuration", err.Error())
		return
	}
	fresh, err := r.client.GetAlertConfiguration(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read alert configuration after create", err.Error())
		return
	}
	writeAlertConfiguration(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetAlertConfiguration(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "alert configuration", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeAlertConfiguration(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alertConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state alertConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var payload json.RawMessage
	if err := json.Unmarshal([]byte(plan.Payload.ValueString()), &payload); err != nil {
		resp.Diagnostics.AddError("Invalid `payload` JSON", err.Error())
		return
	}
	if err := r.client.UpdateAlertConfiguration(ctx, state.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Failed to update alert configuration", err.Error())
		return
	}
	fresh, err := r.client.GetAlertConfiguration(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read alert configuration after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeAlertConfiguration(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAlertConfiguration(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete alert configuration", err.Error())
	}
}

func (r *alertConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeAlertConfiguration(m *alertConfigurationResourceModel, a *rms.AlertConfiguration) {
	m.ID = types.StringValue(a.ID.String())
	if a.Name != "" {
		m.Name = types.StringValue(a.Name)
	}
	if a.Type != "" {
		m.Type = types.StringValue(a.Type)
	}
	m.Enabled = types.BoolValue(a.Enabled)
	m.CreatedAt = timestampToString(a.CreatedAt)
	m.UpdatedAt = timestampToString(a.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*alertConfigurationResource)(nil)
