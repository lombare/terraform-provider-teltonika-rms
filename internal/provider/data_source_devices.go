package provider

import (
	"context"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

var deviceObjectAttrs = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"serial":       types.StringType,
	"mac":          types.StringType,
	"model":        types.StringType,
	"manufacturer": types.StringType,
	"fw_version":   types.StringType,
	"status":       types.StringType,
	"company_id":   types.Int64Type,
	"latitude":     types.Float64Type,
	"longitude":    types.Float64Type,
	"description":  types.StringType,
	"last_seen":    types.StringType,
	"created_at":   types.StringType,
	"updated_at":   types.StringType,
	"raw":          types.StringType,
}

func deviceObject(d rms.Device) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(deviceObjectAttrs, map[string]attr.Value{
		"id":           types.StringValue(d.ID.String()),
		"name":         types.StringValue(d.Name),
		"serial":       types.StringValue(d.Serial),
		"mac":          types.StringValue(d.MAC),
		"model":        types.StringValue(d.Model),
		"manufacturer": types.StringValue(d.Manufacturer),
		"fw_version":   types.StringValue(d.FWVersion),
		"status":       types.StringValue(d.Status),
		"company_id":   numberToInt64(d.CompanyID),
		"latitude":     numberToFloat64(d.Latitude),
		"longitude":    numberToFloat64(d.Longitude),
		"description":  types.StringValue(d.Description),
		"last_seen":    timestampToString(d.LastSeen),
		"created_at":   timestampToString(d.CreatedAt),
		"updated_at":   timestampToString(d.UpdatedAt),
		"raw":          rawToString(d.Raw),
	})
}

type deviceDataSource struct{ client *rms.Client }

type deviceDSModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	Serial       types.String  `tfsdk:"serial"`
	MAC          types.String  `tfsdk:"mac"`
	Model        types.String  `tfsdk:"model"`
	Manufacturer types.String  `tfsdk:"manufacturer"`
	FWVersion    types.String  `tfsdk:"fw_version"`
	Status       types.String  `tfsdk:"status"`
	CompanyID    types.Int64   `tfsdk:"company_id"`
	Latitude     types.Float64 `tfsdk:"latitude"`
	Longitude    types.Float64 `tfsdk:"longitude"`
	Description  types.String  `tfsdk:"description"`
	LastSeen     types.String  `tfsdk:"last_seen"`
	CreatedAt    types.String  `tfsdk:"created_at"`
	UpdatedAt    types.String  `tfsdk:"updated_at"`
	Raw          types.String  `tfsdk:"raw"`
}

func NewDeviceDataSource() datasource.DataSource { return &deviceDataSource{} }

func (d *deviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_device"
}

func (d *deviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *deviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single device by id (`/devices/{id}`). RMS device payloads are wide and vary by model — `raw` exposes the full JSON body for anything not surfaced explicitly.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Required: true},
			"name":         schema.StringAttribute{Computed: true},
			"serial":       schema.StringAttribute{Computed: true},
			"mac":          schema.StringAttribute{Computed: true},
			"model":        schema.StringAttribute{Computed: true},
			"manufacturer": schema.StringAttribute{Computed: true},
			"fw_version":   schema.StringAttribute{Computed: true},
			"status":       schema.StringAttribute{Computed: true},
			"company_id":   schema.Int64Attribute{Computed: true},
			"latitude":     schema.Float64Attribute{Computed: true},
			"longitude":    schema.Float64Attribute{Computed: true},
			"description":  schema.StringAttribute{Computed: true},
			"last_seen":    schema.StringAttribute{Computed: true},
			"created_at":   schema.StringAttribute{Computed: true},
			"updated_at":   schema.StringAttribute{Computed: true},
			"raw":          schema.StringAttribute{Computed: true, Description: "Full JSON body as returned by RMS."},
		},
	}
}

func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg deviceDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	dev, err := d.client.GetDevice(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch device", err.Error())
		return
	}
	cfg.Name = types.StringValue(dev.Name)
	cfg.Serial = types.StringValue(dev.Serial)
	cfg.MAC = types.StringValue(dev.MAC)
	cfg.Model = types.StringValue(dev.Model)
	cfg.Manufacturer = types.StringValue(dev.Manufacturer)
	cfg.FWVersion = types.StringValue(dev.FWVersion)
	cfg.Status = types.StringValue(dev.Status)
	cfg.CompanyID = numberToInt64(dev.CompanyID)
	cfg.Latitude = numberToFloat64(dev.Latitude)
	cfg.Longitude = numberToFloat64(dev.Longitude)
	cfg.Description = types.StringValue(dev.Description)
	cfg.LastSeen = timestampToString(dev.LastSeen)
	cfg.CreatedAt = timestampToString(dev.CreatedAt)
	cfg.UpdatedAt = timestampToString(dev.UpdatedAt)
	cfg.Raw = rawToString(dev.Raw)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

type devicesDataSource struct{ client *rms.Client }

type devicesModel struct {
	CompanyID types.Int64  `tfsdk:"company_id"`
	Search    types.String `tfsdk:"search"`
	Devices   types.List   `tfsdk:"devices"`
}

func NewDevicesDataSource() datasource.DataSource { return &devicesDataSource{} }

func (d *devicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_devices"
}

func (d *devicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *devicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All devices visible to the token, optionally filtered.",
		Attributes: map[string]schema.Attribute{
			"company_id": schema.Int64Attribute{Optional: true},
			"search":     schema.StringAttribute{Optional: true},
			"devices":    schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: deviceObjectAttrs}},
		},
	}
}

func (d *devicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg devicesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	q := url.Values{}
	if !cfg.CompanyID.IsNull() && !cfg.CompanyID.IsUnknown() {
		q.Set("company_id", strconv.FormatInt(cfg.CompanyID.ValueInt64(), 10))
	}
	if s := stringOr(cfg.Search, ""); s != "" {
		q.Set("q", s)
	}
	items, err := d.client.ListDevices(ctx, q)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list devices", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, dev := range items {
		obj, diags := deviceObject(dev)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: deviceObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	cfg.Devices = l
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

// ---- devices monitoring (raw entries) ----

var deviceMonitoringObjectAttrs = map[string]attr.Type{
	"id":     types.StringType,
	"name":   types.StringType,
	"status": types.StringType,
	"raw":    types.StringType,
}

func deviceMonitoringObject(m rms.DeviceMonitoring) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(deviceMonitoringObjectAttrs, map[string]attr.Value{
		"id":     types.StringValue(m.ID.String()),
		"name":   types.StringValue(m.Name),
		"status": types.StringValue(m.Status),
		"raw":    rawToString(m.Raw),
	})
}

type devicesMonitoringDataSource struct{ client *rms.Client }

type devicesMonitoringModel struct {
	Devices types.List `tfsdk:"devices"`
}

func NewDevicesMonitoringDataSource() datasource.DataSource { return &devicesMonitoringDataSource{} }

func (d *devicesMonitoringDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_devices_monitoring"
}

func (d *devicesMonitoringDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *devicesMonitoringDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Live monitoring entries (`/devices/monitoring`). Each entry surfaces the raw JSON RMS returns.",
		Attributes: map[string]schema.Attribute{
			"devices": schema.ListAttribute{Computed: true, ElementType: types.ObjectType{AttrTypes: deviceMonitoringObjectAttrs}},
		},
	}
}

func (d *devicesMonitoringDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	items, err := d.client.ListDeviceMonitoring(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list device monitoring", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(items))
	for _, m := range items {
		obj, diags := deviceMonitoringObject(m)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: deviceMonitoringObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, devicesMonitoringModel{Devices: l})...)
}

// ---- device statistics ----

type deviceStatisticsDataSource struct{ client *rms.Client }

type deviceStatisticsModel struct {
	JSON types.String `tfsdk:"json"`
}

func NewDeviceStatisticsDataSource() datasource.DataSource { return &deviceStatisticsDataSource{} }

func (d *deviceStatisticsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_device_statistics"
}

func (d *deviceStatisticsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *deviceStatisticsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Aggregate device statistics (`/devices/statistics`) returned as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"json": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *deviceStatisticsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	raw, err := d.client.DeviceStatistics(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch device statistics", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, deviceStatisticsModel{JSON: rawToString(raw)})...)
}
