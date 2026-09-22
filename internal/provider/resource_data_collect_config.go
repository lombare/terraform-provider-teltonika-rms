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

type dataCollectConfigResource struct{ client *rms.Client }

type dataCollectConfigResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Payload     types.String `tfsdk:"payload"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Interval    types.Int64  `tfsdk:"interval"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewDataCollectConfigResource() resource.Resource { return &dataCollectConfigResource{} }

func (r *dataCollectConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_collect_config"
}

func (r *dataCollectConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *dataCollectConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data-collect configuration (`/data-collect/configs`). Payload is passed through verbatim.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"payload":     schema.StringAttribute{Required: true, Description: "Raw JSON body sent to POST/PUT `/data-collect/configs`."},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"enabled":     schema.BoolAttribute{Computed: true},
			"interval":    schema.Int64Attribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (r *dataCollectConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataCollectConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var payload json.RawMessage
	if err := json.Unmarshal([]byte(plan.Payload.ValueString()), &payload); err != nil {
		resp.Diagnostics.AddError("Invalid `payload` JSON", err.Error())
		return
	}
	id, err := r.client.CreateDataCollectConfig(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create data-collect config", err.Error())
		return
	}
	fresh, err := r.client.GetDataCollectConfig(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read data-collect config after create", err.Error())
		return
	}
	writeDataCollectConfig(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dataCollectConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataCollectConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetDataCollectConfig(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "data-collect config", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeDataCollectConfig(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *dataCollectConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dataCollectConfigResourceModel
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
	if err := r.client.UpdateDataCollectConfig(ctx, state.ID.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError("Failed to update data-collect config", err.Error())
		return
	}
	fresh, err := r.client.GetDataCollectConfig(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read data-collect config after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeDataCollectConfig(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dataCollectConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataCollectConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteDataCollectConfig(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete data-collect config", err.Error())
	}
}

func (r *dataCollectConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeDataCollectConfig(m *dataCollectConfigResourceModel, d *rms.DataCollectConfig) {
	m.ID = types.StringValue(d.ID.String())
	if d.Name != "" {
		m.Name = types.StringValue(d.Name)
	}
	if d.Description != "" {
		m.Description = types.StringValue(d.Description)
	}
	m.Enabled = types.BoolValue(d.Enabled)
	m.Interval = numberToInt64(d.Interval)
	m.CreatedAt = timestampToString(d.CreatedAt)
	m.UpdatedAt = timestampToString(d.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*dataCollectConfigResource)(nil)
