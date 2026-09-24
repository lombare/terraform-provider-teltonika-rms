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
	resp.TypeName = req.ProviderTypeName + "_rms_alert_configuration"
}

func (r *alertConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *alertConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages an alert configuration (`POST /alerts-configurations`, " +
			"`PUT /alerts-configurations/{id}`, `DELETE /alerts-configurations/{id}`). " +
			"RMS alerts span dozens of trigger types (SIM lifecycle, data usage, signal strength, " +
			"connection state, hotspot events, custom rules, …) each with its own condition and action " +
			"grammar. Rather than pinning the resource to one shape, the entire request body is passed " +
			"through as `payload`. The expected schema is `alert_config_add` in the RMS OpenAPI at " +
			"https://api.rms.teltonika-networks.com/openapi/compiled.yaml.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned identifier of the alert configuration.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"payload": schema.StringAttribute{
				Required: true,
				Description: "Raw JSON body of the alert configuration matching the OpenAPI `alert_config_add` " +
					"shape (the object that would sit under `data[0]` in the RMS request envelope — the provider " +
					"wraps it into the required `{data:[…]}` shape when POSTing). Round-tripped verbatim on update.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Alert configuration name reported by RMS.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Alert type reported by RMS (e.g. `device_offline`, `sim_swap`, `data_usage`).",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the alert is currently active.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of alert configuration creation.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent update.",
			},
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
