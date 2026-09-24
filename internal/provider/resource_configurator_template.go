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

type configuratorTemplateResource struct{ client *rms.Client }

type configuratorTemplateResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Payload   types.String `tfsdk:"payload"`
	Name      types.String `tfsdk:"name"`
	Model     types.String `tfsdk:"model"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewConfiguratorTemplateResource() resource.Resource { return &configuratorTemplateResource{} }

func (r *configuratorTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_configurator_template"
}

func (r *configuratorTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *configuratorTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a device configurator template (`POST /devices/configurator/templates`, " +
			"`DELETE /devices/configurator/templates/{id}`). Templates capture a device configuration for later " +
			"replay via `POST /devices/configurator/templates/{id}/use`. Because template bodies mirror the " +
			"underlying device configuration grammar — which is model-specific and evolves per firmware — the " +
			"payload is passed through as raw JSON. RMS does not expose an update verb for templates, so any " +
			"change to `payload` forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned identifier of the configurator template.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"payload": schema.StringAttribute{
				Required:      true,
				Description:   "Raw JSON body sent to `POST /devices/configurator/templates`. Changes force replacement (no update verb exists).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Template name reported by RMS.",
			},
			"model": schema.StringAttribute{
				Computed:    true,
				Description: "Device model the template targets (e.g. `RUT950`).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of template creation.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent update.",
			},
		},
	}
}

func (r *configuratorTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configuratorTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var payload json.RawMessage
	if err := json.Unmarshal([]byte(plan.Payload.ValueString()), &payload); err != nil {
		resp.Diagnostics.AddError("Invalid `payload` JSON", err.Error())
		return
	}
	id, err := r.client.CreateConfiguratorTemplate(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create configurator template", err.Error())
		return
	}
	fresh, err := r.client.GetConfiguratorTemplate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read configurator template after create", err.Error())
		return
	}
	writeConfiguratorTemplate(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *configuratorTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configuratorTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetConfiguratorTemplate(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "configurator template", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeConfiguratorTemplate(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *configuratorTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// The templates endpoint does not expose an update verb — the resource requires replace
	// on payload change. This body is only reached when non-Payload attributes drift, so we
	// re-read state and return.
	var state configuratorTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetConfiguratorTemplate(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read configurator template", err.Error())
		return
	}
	writeConfiguratorTemplate(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *configuratorTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state configuratorTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteConfiguratorTemplate(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete configurator template", err.Error())
	}
}

func (r *configuratorTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeConfiguratorTemplate(m *configuratorTemplateResourceModel, c *rms.ConfiguratorTemplate) {
	m.ID = types.StringValue(c.ID.String())
	if c.Name != "" {
		m.Name = types.StringValue(c.Name)
	}
	if c.Model != "" {
		m.Model = types.StringValue(c.Model)
	}
	m.CreatedAt = timestampToString(c.CreatedAt)
	m.UpdatedAt = timestampToString(c.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*configuratorTemplateResource)(nil)
