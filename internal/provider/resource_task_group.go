package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

type taskGroupResource struct{ client *rms.Client }

type taskGroupResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CompanyID     types.Int64  `tfsdk:"company_id"`
	SuccessTagIDs types.List   `tfsdk:"success_tag_ids"`
	FailedTagIDs  types.List   `tfsdk:"failed_tag_ids"`
	Tasks         types.List   `tfsdk:"tasks"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewTaskGroupResource() resource.Resource { return &taskGroupResource{} }

func (r *taskGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_task_group"
}

func (r *taskGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (r *taskGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a device task group (`POST /devices/tasks/groups`, " +
			"`PUT /devices/tasks/groups/{id}`, `DELETE /devices/tasks/groups/{id}`). A task group bundles " +
			"an ordered set of tasks (shell commands, file uploads, file downloads) that RMS can execute " +
			"against a set of devices; the group also optionally tags devices with success/failure markers.\n\n" +
			"The `tasks` attribute is a list of task objects. Each entry can be written inline, or supplied " +
			"as a reference to a `teltonika_rms_task_group_task` state-only resource — both compile to the " +
			"same plan since Terraform treats the resource reference as an object with the matching attributes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "RMS-assigned identifier of the task group.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Task group name.",
			},
			"company_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Id of the company that owns the group. Defaults to the token's company when omitted.",
			},
			"success_tag_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "Tag ids attached to a device when the whole group succeeds against it.",
			},
			"failed_tag_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "Tag ids attached to a device when the group fails against it.",
			},
			"tasks": schema.ListNestedAttribute{
				Required: true,
				Description: "Ordered list of tasks in the group. Each entry mirrors the OpenAPI " +
					"`device_task_groups_create_tasks` shape.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "RMS-assigned task identifier (populated on read).",
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Task name.",
						},
						"order": schema.Int64Attribute{
							Required:    true,
							Description: "Execution order within the group; must be >= 0.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Task type. One of `command`, `upload_file`, `download_file`.",
						},
						"timeout": schema.Int64Attribute{
							Required:    true,
							Description: "Runtime cap in seconds (1–300).",
						},
						"stop_on_failure": schema.BoolAttribute{
							Required:    true,
							Description: "Abort the remaining tasks if this one fails.",
						},
						"data": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Type-specific payload. Only the fields matching `type` are used.",
							Attributes: map[string]schema.Attribute{
								"command":          schema.StringAttribute{Optional: true, Description: "Shell command (type `command`)."},
								"acceptable_codes": schema.ListAttribute{Optional: true, ElementType: types.Int64Type, Description: "Exit codes not treated as failure (type `command`)."},
								"file_id":          schema.Int64Attribute{Optional: true, Description: "Id of the file to push (type `upload_file`)."},
								"target_path":      schema.StringAttribute{Optional: true, Description: "Destination path on the device (type `upload_file`)."},
								"file":             schema.StringAttribute{Optional: true, Description: "Source path on the device (type `download_file`)."},
							},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of group creation.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent update.",
			},
		},
	}
}

func (r *taskGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan taskGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tasks := tasksFromModel(ctx, plan.Tasks, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.CreateTaskGroup(ctx, rms.TaskGroupCreate{
		Name:         plan.Name.ValueString(),
		CompanyID:    int64Or(plan.CompanyID, 0),
		Tasks:        tasks,
		SuccessTagID: int64SliceFromList(plan.SuccessTagIDs, &resp.Diagnostics),
		FailedTagID:  int64SliceFromList(plan.FailedTagIDs, &resp.Diagnostics),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create task group", err.Error())
		return
	}
	writeTaskGroup(&plan, group, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *taskGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state taskGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fresh, err := r.client.GetTaskGroup(ctx, state.ID.ValueString())
	if handleNotFound(err, &resp.Diagnostics, "task group", state.ID.ValueString()) {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	writeTaskGroup(&state, fresh, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *taskGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state taskGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tasks := tasksFromModel(ctx, plan.Tasks, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateTaskGroup(ctx, state.ID.ValueString(), rms.TaskGroupUpdate{
		Name:  plan.Name.ValueString(),
		Tasks: tasks,
	}); err != nil {
		resp.Diagnostics.AddError("Failed to update task group", err.Error())
		return
	}
	fresh, err := r.client.GetTaskGroup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to re-read task group after update", err.Error())
		return
	}
	plan.ID = state.ID
	writeTaskGroup(&plan, fresh, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *taskGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state taskGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTaskGroup(ctx, state.ID.ValueString()); err != nil && !rms.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete task group", err.Error())
	}
}

func (r *taskGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ resource.ResourceWithImportState = (*taskGroupResource)(nil)

// -------- helpers shared with the task-group data sources ---------------------------

// tasksFromModel converts a Terraform list-of-nested-objects into the RMS Task slice.
func tasksFromModel(ctx context.Context, list types.List, diags *diag.Diagnostics) []rms.Task {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := list.Elements()
	out := make([]rms.Task, 0, len(elems))
	for _, e := range elems {
		obj, ok := e.(types.Object)
		if !ok {
			diags.AddError("Type error", "expected task object")
			return nil
		}
		attrs := obj.Attributes()
		t := rms.Task{
			Name:          asString(attrs["name"]),
			Order:         int(asInt64(attrs["order"])),
			Type:          asString(attrs["type"]),
			Timeout:       int(asInt64(attrs["timeout"])),
			StopOnFailure: asBool(attrs["stop_on_failure"]),
			Data:          taskDataFromObject(attrs["data"]),
		}
		out = append(out, t)
	}
	return out
}

func taskDataFromObject(v attr.Value) rms.TaskData {
	obj, ok := v.(types.Object)
	if !ok || obj.IsNull() || obj.IsUnknown() {
		return rms.TaskData{}
	}
	attrs := obj.Attributes()
	data := rms.TaskData{
		Command:    asString(attrs["command"]),
		FileID:     asInt64(attrs["file_id"]),
		TargetPath: asString(attrs["target_path"]),
		File:       asString(attrs["file"]),
	}
	if v, ok := attrs["acceptable_codes"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		for _, e := range v.Elements() {
			if n, ok := e.(types.Int64); ok {
				data.AcceptableCodes = append(data.AcceptableCodes, int(n.ValueInt64()))
			}
		}
	}
	return data
}

func asString(v attr.Value) string {
	s, ok := v.(types.String)
	if !ok || s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func asInt64(v attr.Value) int64 {
	n, ok := v.(types.Int64)
	if !ok || n.IsNull() || n.IsUnknown() {
		return 0
	}
	return n.ValueInt64()
}

func asBool(v attr.Value) bool {
	b, ok := v.(types.Bool)
	if !ok || b.IsNull() || b.IsUnknown() {
		return false
	}
	return b.ValueBool()
}

func writeTaskGroup(m *taskGroupResourceModel, g *rms.TaskGroup, diags *diag.Diagnostics) {
	m.ID = types.StringValue(g.ID.String())
	m.Name = types.StringValue(g.Name)
	m.CompanyID = numberToInt64(g.CompanyID)
	m.SuccessTagIDs = listFromInt64sNumbers(g.SuccessTagID)
	m.FailedTagIDs = listFromInt64sNumbers(g.FailedTagID)
	m.CreatedAt = timestampToString(g.CreatedAt)
	m.UpdatedAt = timestampToString(g.UpdatedAt)
	m.Tasks = tasksToList(rms.SortTasks(g.Tasks), diags)
}

func tasksToList(tasks []rms.Task, diags *diag.Diagnostics) types.List {
	elems := make([]attr.Value, 0, len(tasks))
	for _, t := range tasks {
		obj, d := taskToObject(t)
		diags.Append(d...)
		elems = append(elems, obj)
	}
	l, d := types.ListValue(types.ObjectType{AttrTypes: TaskGroupTaskObjectAttrs}, elems)
	diags.Append(d...)
	return l
}

func taskToObject(t rms.Task) (types.Object, diag.Diagnostics) {
	dataObj, d := types.ObjectValue(
		TaskGroupTaskObjectAttrs["data"].(types.ObjectType).AttrTypes,
		map[string]attr.Value{
			"command":          types.StringValue(t.Data.Command),
			"acceptable_codes": intSliceToList(t.Data.AcceptableCodes),
			"file_id":          types.Int64Value(t.Data.FileID),
			"target_path":      types.StringValue(t.Data.TargetPath),
			"file":             types.StringValue(t.Data.File),
		},
	)
	if d.HasError() {
		return types.ObjectNull(TaskGroupTaskObjectAttrs), d
	}
	return types.ObjectValue(TaskGroupTaskObjectAttrs, map[string]attr.Value{
		"id":              types.StringValue(t.ID.String()),
		"name":            types.StringValue(t.Name),
		"order":           types.Int64Value(int64(t.Order)),
		"type":            types.StringValue(t.Type),
		"timeout":         types.Int64Value(int64(t.Timeout)),
		"stop_on_failure": types.BoolValue(t.StopOnFailure),
		"data":            dataObj,
	})
}

func intSliceToList(v []int) types.List {
	elems := make([]attr.Value, 0, len(v))
	for _, n := range v {
		elems = append(elems, types.Int64Value(int64(n)))
	}
	l, _ := types.ListValue(types.Int64Type, elems)
	return l
}
