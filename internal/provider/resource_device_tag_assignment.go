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
	resp.TypeName = req.ProviderTypeName + "_rms_device_tag_assignment"
}

func (r *deviceTagAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *deviceTagAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Declares the exclusive tag set attached to a device. Every apply calls " +
			"`POST /devices/tags/overwrite` with the declared list, so drift (tags added or removed outside " +
			"Terraform) is corrected in place. On destroy, `POST /devices/tags/unassign` removes only the tags " +
			"this resource introduced. Note: RMS does not expose a single-endpoint read that returns just the " +
			"tag set for one device, so drift is reconciled at apply time rather than surfaced in `terraform plan`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Stable hash of the target `device_id`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"device_id": schema.Int64Attribute{
				Required:    true,
				Description: "RMS device id whose tag set is being managed.",
			},
			"tag_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "Tag ids that make up the exclusive set for the device. Any tag currently attached to the device but not listed here is removed on apply.",
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
