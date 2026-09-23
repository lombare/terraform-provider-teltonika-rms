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

type roleResource struct{ client *rms.Client }

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	CompanyIDs  types.List   `tfsdk:"company_ids"`
	Permissions types.List   `tfsdk:"permission_ids"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewRoleResource() resource.Resource { return &roleResource{} }

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_role"
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "RBAC role (`/roles`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"title":       schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true, Computed: true},
			"company_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "Company ids the role applies to.",
			},
			"permission_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "Permission ids that make up the role.",
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateRole(ctx, rms.RoleCreate{
		Title:        plan.Title.ValueString(),
		Description:  plan.Description.ValueString(),
		CompanyIDs:   int64SliceFromList(plan.CompanyIDs, &resp.Diagnostics),
		PermissionID: int64SliceFromList(plan.Permissions, &resp.Diagnostics),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create role", err.Error())
		return
	}
	writeRole(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetRole(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "role", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeRole(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := parseInt64(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid role id", err.Error())
		return
	}
	if err := r.client.UpdateRole(ctx, rms.RoleUpdate{
		ID:           id,
		Title:        plan.Title.ValueString(),
		Description:  plan.Description.ValueString(),
		PermissionID: int64SliceFromList(plan.Permissions, &resp.Diagnostics),
	}); err != nil {
		resp.Diagnostics.AddError("Failed to update role", err.Error())
		return
	}
	fresh, err := r.client.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read role after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeRole(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRole(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete role", err.Error())
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeRole(m *roleResourceModel, role *rms.Role) {
	m.ID = types.StringValue(role.ID.String())
	m.Title = types.StringValue(role.Title)
	if role.Description != "" {
		m.Description = types.StringValue(role.Description)
	}
	m.CompanyIDs = listFromInt64sNumbers(role.CompanyID)
	m.Permissions = listFromInt64sNumbers(role.Permissions)
	m.CreatedAt = timestampToString(role.CreatedAt)
	m.UpdatedAt = timestampToString(role.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*roleResource)(nil)
