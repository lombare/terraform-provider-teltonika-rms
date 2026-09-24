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

type automationResource struct{ client *rms.Client }

type automationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Payload     types.String `tfsdk:"payload"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewAutomationResource() resource.Resource { return &automationResource{} }

func (r *automationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_automation"
}

func (r *automationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *automationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages an RMS automation (`POST /automations`, `PUT /automations/{id}`, " +
			"`DELETE /automations/{id}`). Automations pair triggers (device state changes, alerts, schedules) " +
			"with conditions and actions (device reboot, firmware push, notification, …); the request grammar " +
			"is broad enough that the entire payload is passed through as raw JSON. Refer to the RMS OpenAPI " +
			"at https://api.rms.teltonika-networks.com/openapi/compiled.yaml for the current body shape.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned identifier of the automation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"payload": schema.StringAttribute{
				Required:    true,
				Description: "Raw JSON body sent to `POST /automations` on create and `PUT /automations/{id}` on update.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Automation name reported by RMS.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Automation description reported by RMS.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the automation is currently active.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of automation creation.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent update.",
			},
		},
	}
}

func (r *automationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan automationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var payload json.RawMessage
	if err := json.Unmarshal([]byte(plan.Payload.ValueString()), &payload); err != nil {
		resp.Diagnostics.AddError("Invalid `payload` JSON", err.Error())
		return
	}
	id, err := r.client.CreateAutomation(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create automation", err.Error())
		return
	}
	fresh, err := r.client.GetAutomation(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read automation after create", err.Error())
		return
	}
	writeAutomation(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *automationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state automationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetAutomation(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "automation", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeAutomation(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *automationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state automationResourceModel
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
	if err := r.client.UpdateAutomation(ctx, state.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Failed to update automation", err.Error())
		return
	}
	fresh, err := r.client.GetAutomation(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read automation after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeAutomation(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *automationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state automationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAutomation(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete automation", err.Error())
	}
}

func (r *automationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeAutomation(m *automationResourceModel, a *rms.Automation) {
	m.ID = types.StringValue(a.ID.String())
	if a.Name != "" {
		m.Name = types.StringValue(a.Name)
	}
	if a.Description != "" {
		m.Description = types.StringValue(a.Description)
	}
	m.Enabled = types.BoolValue(a.Enabled)
	m.CreatedAt = timestampToString(a.CreatedAt)
	m.UpdatedAt = timestampToString(a.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*automationResource)(nil)
