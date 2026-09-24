package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// taskGroupTaskResource is a **state-only** resource. It performs no RMS API calls; it
// exists to give users a named, reusable handle for a task definition that can be spliced
// into the `tasks` attribute of one or more `teltonika_rms_task_group` resources. The RMS
// API creates tasks nested inside a group (`POST /devices/tasks/groups`), so this record
// carries the fields the task-group resource assembles into that request body.
type taskGroupTaskResource struct{}

func NewTaskGroupTaskResource() resource.Resource { return &taskGroupTaskResource{} }

// TaskGroupTaskObjectAttrs / TaskGroupTaskObjectSchema are shared with resource_task_group
// so the group's `tasks` list-of-objects has exactly the same shape a task record exposes.
// Terraform then accepts either an inline object literal or a resource reference.
var TaskGroupTaskObjectAttrs = map[string]attr.Type{
	"id":              types.StringType,
	"name":            types.StringType,
	"order":           types.Int64Type,
	"type":            types.StringType,
	"timeout":         types.Int64Type,
	"stop_on_failure": types.BoolType,
	"data": types.ObjectType{AttrTypes: map[string]attr.Type{
		"command":          types.StringType,
		"acceptable_codes": types.ListType{ElemType: types.Int64Type},
		"file_id":          types.Int64Type,
		"target_path":      types.StringType,
		"file":             types.StringType,
	}},
}

type taskGroupTaskResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Order         types.Int64  `tfsdk:"order"`
	Type          types.String `tfsdk:"type"`
	Timeout       types.Int64  `tfsdk:"timeout"`
	StopOnFailure types.Bool   `tfsdk:"stop_on_failure"`
	Data          types.Object `tfsdk:"data"`
}

func (r *taskGroupTaskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_task_group_task"
}

func (r *taskGroupTaskResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func (r *taskGroupTaskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "**State-only helper resource.** Declares a single task definition (name, order, type, " +
			"timeout, stop-on-failure, data) so it can be referenced from one or more " +
			"`teltonika_rms_task_group` resources through the group's `tasks` attribute. " +
			"\n\n" +
			"This resource performs **no RMS API calls** on any lifecycle event — Create/Read/Update/Delete " +
			"just move the plan into state. RMS creates and destroys tasks as a side effect of the group's " +
			"POST/PUT (`POST /devices/tasks/groups`), so a task record only makes sense when it is spliced " +
			"into a group's `tasks` list. The task-group resource is what actually talks to RMS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Stable client-side identifier derived from `name` + `order`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Task name as displayed in the RMS UI.",
			},
			"order": schema.Int64Attribute{
				Required:    true,
				Description: "Position of the task inside the group. RMS uses this to order execution; must be >= 0.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Task type. One of `command`, `upload_file`, `download_file`.",
			},
			"timeout": schema.Int64Attribute{
				Required:    true,
				Description: "Maximum runtime of the task in seconds. Must be between 1 and 300.",
			},
			"stop_on_failure": schema.BoolAttribute{
				Required:    true,
				Description: "If true, a failure aborts the remaining tasks in the group.",
			},
			"data": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Type-specific payload sent to the device. Only the fields matching `type` are used; the rest are ignored server-side.",
				Attributes: map[string]schema.Attribute{
					"command": schema.StringAttribute{
						Optional:    true,
						Description: "Shell command to execute. Required when `type = command`.",
					},
					"acceptable_codes": schema.ListAttribute{
						Optional:    true,
						ElementType: types.Int64Type,
						Description: "Exit codes that should NOT be treated as failure. Optional; used only when `type = command`.",
					},
					"file_id": schema.Int64Attribute{
						Optional:    true,
						Description: "Id of the file to push to the device. Required when `type = upload_file`.",
					},
					"target_path": schema.StringAttribute{
						Optional:    true,
						Description: "Absolute path on the device where the file will be written. Required when `type = upload_file`.",
					},
					"file": schema.StringAttribute{
						Optional:    true,
						Description: "Absolute path on the device of the file to fetch. Required when `type = download_file`.",
					},
				},
			},
		},
	}
}

func (r *taskGroupTaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan taskGroupTaskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(taskGroupTaskID(plan.Name.ValueString(), plan.Order.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *taskGroupTaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state taskGroupTaskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *taskGroupTaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan taskGroupTaskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(taskGroupTaskID(plan.Name.ValueString(), plan.Order.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *taskGroupTaskResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *taskGroupTaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func taskGroupTaskID(name string, order int64) string {
	sum := sha256.Sum256([]byte(name + "|" + strconv.FormatInt(order, 10)))
	return hex.EncodeToString(sum[:8])
}

var _ resource.ResourceWithImportState = (*taskGroupTaskResource)(nil)
