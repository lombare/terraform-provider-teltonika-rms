package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

// ---- teltonika_rms_task_group (single) ----

type taskGroupDataSource struct{ client *rms.Client }

type taskGroupDSModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CompanyID     types.Int64  `tfsdk:"company_id"`
	SuccessTagIDs types.List   `tfsdk:"success_tag_ids"`
	FailedTagIDs  types.List   `tfsdk:"failed_tag_ids"`
	Tasks         types.List   `tfsdk:"tasks"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewTaskGroupDataSource() datasource.DataSource { return &taskGroupDataSource{} }

func (d *taskGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_task_group"
}

func (d *taskGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *taskGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single task group by id (`GET /devices/tasks/groups/{id}`) and its tasks " +
			"(`GET /devices/tasks?group_id={id}`).",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, Description: "RMS task group id."},
			"name":            schema.StringAttribute{Computed: true},
			"company_id":      schema.Int64Attribute{Computed: true},
			"success_tag_ids": schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"failed_tag_ids":  schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
			"tasks": schema.ListAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: TaskGroupTaskObjectAttrs},
				Description: "Tasks belonging to the group, in execution order.",
			},
			"created_at": schema.StringAttribute{Computed: true},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *taskGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg taskGroupDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	g, err := d.client.GetTaskGroup(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch task group", err.Error())
		return
	}
	cfg.Name = types.StringValue(g.Name)
	cfg.CompanyID = numberToInt64(g.CompanyID)
	cfg.SuccessTagIDs = listFromInt64sNumbers(g.SuccessTagID)
	cfg.FailedTagIDs = listFromInt64sNumbers(g.FailedTagID)
	cfg.Tasks = tasksToList(rms.SortTasks(g.Tasks), &resp.Diagnostics)
	cfg.CreatedAt = timestampToString(g.CreatedAt)
	cfg.UpdatedAt = timestampToString(g.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

// ---- teltonika_rms_task_groups (list) ----

type taskGroupsDataSource struct{ client *rms.Client }

type taskGroupsModel struct {
	TaskGroups types.List `tfsdk:"task_groups"`
}

var taskGroupObjectAttrs = map[string]attr.Type{
	"id":              types.StringType,
	"name":            types.StringType,
	"company_id":      types.Int64Type,
	"success_tag_ids": types.ListType{ElemType: types.Int64Type},
	"failed_tag_ids":  types.ListType{ElemType: types.Int64Type},
	"created_at":      types.StringType,
	"updated_at":      types.StringType,
}

func taskGroupObject(g rms.TaskGroup) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(taskGroupObjectAttrs, map[string]attr.Value{
		"id":              types.StringValue(g.ID.String()),
		"name":            types.StringValue(g.Name),
		"company_id":      numberToInt64(g.CompanyID),
		"success_tag_ids": listFromInt64sNumbers(g.SuccessTagID),
		"failed_tag_ids":  listFromInt64sNumbers(g.FailedTagID),
		"created_at":      timestampToString(g.CreatedAt),
		"updated_at":      timestampToString(g.UpdatedAt),
	})
}

func NewTaskGroupsDataSource() datasource.DataSource { return &taskGroupsDataSource{} }

func (d *taskGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_task_groups"
}

func (d *taskGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *taskGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every task group visible to the token (`GET /devices/tasks/groups`, paginated " +
			"transparently). Only group metadata is returned; use `teltonika_rms_task_group` (singular) to " +
			"fetch a group's task list as well.",
		Attributes: map[string]schema.Attribute{
			"task_groups": schema.ListAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: taskGroupObjectAttrs},
			},
		},
	}
}

func (d *taskGroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListTaskGroups(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list task groups", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, g := range items {
		obj, diags := taskGroupObject(g)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: taskGroupObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, taskGroupsModel{TaskGroups: l})...)
}
