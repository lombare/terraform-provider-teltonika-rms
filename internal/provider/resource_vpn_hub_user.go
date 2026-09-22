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

type vpnHubUserResource struct{ client *rms.Client }

type vpnHubUserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	HubID     types.Int64  `tfsdk:"hub_id"`
	Username  types.String `tfsdk:"username"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func NewVPNHubUserResource() resource.Resource { return &vpnHubUserResource{} }

func (r *vpnHubUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_hub_user"
}

func (r *vpnHubUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *vpnHubUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "VPN hub user (`/vpn/hubs/users`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":       schema.StringAttribute{Required: true},
			"hub_id":     schema.Int64Attribute{Required: true},
			"username":   schema.StringAttribute{Computed: true},
			"enabled":    schema.BoolAttribute{Optional: true, Computed: true},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *vpnHubUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpnHubUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateVPNHubUser(ctx, rms.VPNHubUserCreate{
		Name:  plan.Name.ValueString(),
		HubID: plan.HubID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VPN hub user", err.Error())
		return
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		if err := r.client.ToggleVPNHubUser(ctx, created.ID.String(), plan.Enabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Failed to toggle VPN hub user state on create", err.Error())
			return
		}
	}
	fresh, err := r.client.GetVPNHubUser(ctx, created.ID.String())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read VPN hub user after create", err.Error())
		return
	}
	writeVPNHubUser(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vpnHubUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpnHubUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetVPNHubUser(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "VPN hub user", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeVPNHubUser(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vpnHubUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vpnHubUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		if err := r.client.ToggleVPNHubUser(ctx, state.ID.ValueString(), plan.Enabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Failed to toggle VPN hub user on update", err.Error())
			return
		}
	}
	fresh, err := r.client.GetVPNHubUser(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read VPN hub user after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeVPNHubUser(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *vpnHubUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpnHubUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVPNHubUser(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete VPN hub user", err.Error())
	}
}

func (r *vpnHubUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeVPNHubUser(m *vpnHubUserResourceModel, u *rms.VPNHubUser) {
	m.ID = types.StringValue(u.ID.String())
	m.Name = types.StringValue(u.Name)
	m.HubID = numberToInt64(u.HubID)
	if u.Username != "" {
		m.Username = types.StringValue(u.Username)
	}
	m.Enabled = types.BoolValue(u.Enabled)
	m.CreatedAt = timestampToString(u.CreatedAt)
	m.UpdatedAt = timestampToString(u.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*vpnHubUserResource)(nil)
