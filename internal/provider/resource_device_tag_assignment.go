package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

// deviceTagAssignmentResource declares an exclusive set of tag ids on a device — every
// apply calls the overwrite endpoint so drift is corrected in place. The id is a stable
// hash of the device id so plans stay legible.
type deviceTagAssignmentResource struct{ client *rms.Client }

type deviceTagAssignmentResourceModel struct {
	ID       types.String `tfsdk:"id"`
	DeviceID types.Int64  `tfsdk:"device_id"`
	TagIDs   types.List   `tfsdk:"tag_ids"`
}

func NewDeviceTagAssignmentResource() resource.Resource { return &deviceTagAssignmentResource{} }

func (r *deviceTagAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_tag_assignment"
}

func (r *deviceTagAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *deviceTagAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Exclusive tag set on a device (`/devices/tags/overwrite`). Every apply replaces the device's tag set with the declared value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"device_id": schema.Int64Attribute{
				Required: true,
			},
			"tag_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (r *deviceTagAssignmentResource) apply(ctx context.Context, deviceID int64, tagIDs []int64) error {
	return r.client.OverwriteDeviceTags(ctx, []int64{deviceID}, tagIDs)
}

func (r *deviceTagAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceTagAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagIDs := int64SliceFromList(plan.TagIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.DeviceID.ValueInt64(), tagIDs); err != nil {
		resp.Diagnostics.AddError("Failed to set device tag assignment", err.Error())
		return
	}
	plan.ID = types.StringValue(deviceTagAssignmentID(plan.DeviceID.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *deviceTagAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No dedicated endpoint returns just this projection, and the device shape varies by
	// model — we trust local state until the next apply reconciles.
	var state deviceTagAssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *deviceTagAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceTagAssignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagIDs := int64SliceFromList(plan.TagIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.DeviceID.ValueInt64(), tagIDs); err != nil {
		resp.Diagnostics.AddError("Failed to update device tag assignment", err.Error())
		return
	}
	plan.ID = types.StringValue(deviceTagAssignmentID(plan.DeviceID.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *deviceTagAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceTagAssignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagIDs := int64SliceFromList(state.TagIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UnassignDeviceTags(ctx, []int64{state.DeviceID.ValueInt64()}, tagIDs); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to remove device tag assignment", err.Error())
	}
}

func (r *deviceTagAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func deviceTagAssignmentID(deviceID int64) string {
	sum := sha256.Sum256([]byte(strconv.FormatInt(deviceID, 10)))
	return hex.EncodeToString(sum[:8])
}

var _ resource.ResourceWithImportState = (*deviceTagAssignmentResource)(nil)
