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

type vpnHubResource struct{ client *rms.Client }

type vpnHubResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	Description types.String `tfsdk:"description"`
	HubZone     types.String `tfsdk:"hub_zone"`
	VPNType     types.String `tfsdk:"vpn_type"`
	TagIDs      types.List   `tfsdk:"tag_ids"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewVPNHubResource() resource.Resource { return &vpnHubResource{} }

func (r *vpnHubResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_vpn_hub"
}

func (r *vpnHubResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *vpnHubResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a VPN hub (`POST /vpn/hubs`, `PUT /vpn/hubs/{id}`, `DELETE /vpn/hubs`). " +
			"A hub is the server-side termination point for RMS-managed VPN tunnels; devices connect to it " +
			"either by tag or by explicit assignment. `enabled` toggles the hub via `PUT /vpn/hubs/{id}/toggle`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned identifier of the VPN hub.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable VPN hub name.",
			},
			"company_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Id of the company that owns the hub. Defaults to the token's company when omitted.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Free-form hub description.",
			},
			"hub_zone": schema.StringAttribute{
				Required:    true,
				Description: "RMS server region hosting the hub. One of `frankfurt-1`, `bahrain-1`.",
			},
			"vpn_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "OpenVPN topology. One of `tap` (layer-2 bridged) or `tun` (layer-3 routed).",
			},
			"tag_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "Ids of tags whose devices are auto-attached to this hub.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the hub is currently accepting connections. Changes are applied via `PUT /vpn/hubs/{id}/toggle`.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of hub creation.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent update.",
			},
		},
	}
}

func (r *vpnHubResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpnHubResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateVPNHub(ctx, rms.VPNHubCreate{
		Name:        plan.Name.ValueString(),
		CompanyID:   int64Or(plan.CompanyID, 0),
		Description: stringOr(plan.Description, ""),
		HubZone:     plan.HubZone.ValueString(),
		VPNType:     stringOr(plan.VPNType, ""),
		TagID:       int64SliceFromList(plan.TagIDs, &resp.Diagnostics),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VPN hub", err.Error())
		return
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		if err := r.client.ToggleVPNHub(ctx, created.ID.String(), plan.Enabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Failed to toggle VPN hub state on create", err.Error())
			return
		}
	}
	fresh, err := r.client.GetVPNHub(ctx, created.ID.String())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read VPN hub after create", err.Error())
		return
	}
	writeVPNHub(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vpnHubResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpnHubResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetVPNHub(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "VPN hub", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeVPNHub(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vpnHubResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vpnHubResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateVPNHub(ctx, state.ID.ValueString(), rms.VPNHubUpdate{
		Name:        plan.Name.ValueString(),
		Description: stringOr(plan.Description, ""),
		VPNType:     stringOr(plan.VPNType, ""),
	}); err != nil {
		resp.Diagnostics.AddError("Failed to update VPN hub", err.Error())
		return
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		if err := r.client.ToggleVPNHub(ctx, state.ID.ValueString(), plan.Enabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Failed to toggle VPN hub on update", err.Error())
			return
		}
	}
	fresh, err := r.client.GetVPNHub(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read VPN hub after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeVPNHub(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vpnHubResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpnHubResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVPNHub(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete VPN hub", err.Error())
	}
}

func (r *vpnHubResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeVPNHub(m *vpnHubResourceModel, h *rms.VPNHub) {
	m.ID = types.StringValue(h.ID.String())
	m.Name = types.StringValue(h.Name)
	m.CompanyID = numberToInt64(h.CompanyID)
	if h.Description != "" {
		m.Description = types.StringValue(h.Description)
	}
	if h.HubZone != "" {
		m.HubZone = types.StringValue(h.HubZone)
	}
	if h.VPNType != "" {
		m.VPNType = types.StringValue(h.VPNType)
	}
	m.TagIDs = listFromInt64sNumbers(h.TagIDs)
	m.Enabled = types.BoolValue(h.Enabled)
	m.CreatedAt = timestampToString(h.CreatedAt)
	m.UpdatedAt = timestampToString(h.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*vpnHubResource)(nil)
