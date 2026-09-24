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

// userInvitationResource wraps /users/invite + DELETE /users/invitations/{id}. Invitations
// are ephemeral (accepted → become users), so this resource treats all attributes as
// replace-only.
type userInvitationResource struct{ client *rms.Client }

type userInvitationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CompanyID types.Int64  `tfsdk:"company_id"`
}

func NewUserInvitationResource() resource.Resource { return &userInvitationResource{} }

func (r *userInvitationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_user_invitation"
}

func (r *userInvitationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *userInvitationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Sends an invitation email to `email` for the given `company_id` with the given `role` " +
			"(`POST /users/invite`). RMS stores the invitation until the invitee accepts, at which point they " +
			"become an ordinary user. `email` and `role` are immutable — changing them replaces the invitation. " +
			"On destroy, the invitation is revoked via `DELETE /users/invitations/{id}`; if the invitee has " +
			"already accepted the revoke will 404, at which point the resource is dropped from state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned invitation identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"email": schema.StringAttribute{
				Required:      true,
				PlanModifiers: replace,
				Description:   "Email address the invitation is sent to. Changes force a new invitation.",
			},
			"role": schema.StringAttribute{
				Required:      true,
				PlanModifiers: replace,
				Description:   "Role granted on acceptance. One of `admin`, `end_user`, `read_only`. Changes force a new invitation.",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "Id of the company the invitee will be attached to.",
			},
		},
	}
}

func (r *userInvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userInvitationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.InviteUser(ctx, rms.InviteUser{
		Role:      plan.Role.ValueString(),
		Email:     plan.Email.ValueString(),
		CompanyID: plan.CompanyID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to invite user", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userInvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// The RMS API does not expose a GET-single endpoint for invitations. We accept the
	// state as-is; if the invitation has been accepted or revoked the DELETE will 404 and
	// the resource will be recreated on the next apply.
	var state userInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userInvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// company_id is the only mutable attribute and RMS lacks an update verb; we simply pass
	// the plan back to state. Any real change to role/email triggers replacement via
	// RequiresReplace.
	var plan userInvitationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userInvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteInvitation(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to revoke invitation", err.Error())
	}
}

func (r *userInvitationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ resource.ResourceWithImportState = (*userInvitationResource)(nil)
