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

type companyResource struct{ client *rms.Client }

type companyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ParentID    types.Int64  `tfsdk:"parent_id"`
	Email       types.String `tfsdk:"email"`
	Level       types.Int64  `tfsdk:"level"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewCompanyResource() resource.Resource { return &companyResource{} }

func (r *companyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_company"
}

func (r *companyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *companyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a subsidiary company under the account owning the RMS token. " +
			"Wraps the `POST /companies`, `PUT /companies/{id}`, and `DELETE /companies/{id}` endpoints. " +
			"See https://developers.rms.teltonika-networks.com/pages/api.html for the full API reference.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned identifier of the company.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable company name. Length must be between 3 and 200 characters.",
			},
			"parent_id": schema.Int64Attribute{
				Required:    true,
				Description: "Id of the parent company the new subsidiary is nested under. Required by RMS at creation time.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Contact email for the company. Required when updating an existing company; optional at creation (the API only accepts email via the update endpoint).",
			},
			"level": schema.Int64Attribute{
				Computed:    true,
				Description: "Depth of the company in the parent/subsidiary tree, reported by the API.",
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of devices currently registered under this company.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC-3339-ish timestamp of when the company was created (RMS format `Y-m-d H:i:s`).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent update, same format as `created_at`.",
			},
		},
	}
}

func (r *companyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan companyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateCompany(ctx, rms.CompanyCreate{
		Name:     plan.Name.ValueString(),
		ParentID: plan.ParentID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create company", err.Error())
		return
	}

	if plan.Email.ValueString() != "" {
		if err := r.client.UpdateCompany(ctx, created.ID.String(), rms.CompanyUpdate{
			Name:  created.Name,
			Email: plan.Email.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError("Failed to set company email after create", err.Error())
			return
		}
	}

	fresh, err := r.client.GetCompany(ctx, created.ID.String())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read company after create", err.Error())
		return
	}
	writeCompany(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *companyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state companyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetCompany(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "company", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeCompany(&state, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *companyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state companyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateCompany(ctx, state.ID.ValueString(), rms.CompanyUpdate{
		Name:  plan.Name.ValueString(),
		Email: plan.Email.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Failed to update company", err.Error())
		return
	}
	fresh, err := r.client.GetCompany(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read company after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeCompany(&plan, fresh)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *companyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state companyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCompany(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete company", err.Error())
	}
}

func (r *companyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func writeCompany(m *companyResourceModel, c *rms.Company) {
	m.ID = types.StringValue(c.ID.String())
	m.Name = types.StringValue(c.Name)
	m.ParentID = numberToInt64(c.ParentID)
	if c.Email != "" {
		m.Email = types.StringValue(c.Email)
	}
	m.Level = numberToInt64(c.Level)
	m.DeviceCount = numberToInt64(c.DeviceCount)
	m.CreatedAt = timestampToString(c.CreatedAt)
	m.UpdatedAt = timestampToString(c.UpdatedAt)
}

var _ resource.ResourceWithImportState = (*companyResource)(nil)
