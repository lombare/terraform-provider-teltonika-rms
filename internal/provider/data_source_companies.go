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

var companyObjectAttrs = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"parent_id":    types.Int64Type,
	"email":        types.StringType,
	"level":        types.Int64Type,
	"device_count": types.Int64Type,
	"created_at":   types.StringType,
	"updated_at":   types.StringType,
}

func companyObject(c rms.Company) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(companyObjectAttrs, map[string]attr.Value{
		"id":           types.StringValue(c.ID.String()),
		"name":         types.StringValue(c.Name),
		"parent_id":    numberToInt64(c.ParentID),
		"email":        types.StringValue(c.Email),
		"level":        numberToInt64(c.Level),
		"device_count": numberToInt64(c.DeviceCount),
		"created_at":   timestampToString(c.CreatedAt),
		"updated_at":   timestampToString(c.UpdatedAt),
	})
}

// ---- company (single) ----

type companyDataSource struct{ client *rms.Client }

type companyDSModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ParentID    types.Int64  `tfsdk:"parent_id"`
	Email       types.String `tfsdk:"email"`
	Level       types.Int64  `tfsdk:"level"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewCompanyDataSource() datasource.DataSource { return &companyDataSource{} }

func (d *companyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_company"
}

func (d *companyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *companyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lookup a company by id.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Required: true},
			"name":         schema.StringAttribute{Computed: true},
			"parent_id":    schema.Int64Attribute{Computed: true},
			"email":        schema.StringAttribute{Computed: true},
			"level":        schema.Int64Attribute{Computed: true},
			"device_count": schema.Int64Attribute{Computed: true},
			"created_at":   schema.StringAttribute{Computed: true},
			"updated_at":   schema.StringAttribute{Computed: true},
		},
	}
}

func (d *companyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg companyDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	c, err := d.client.GetCompany(ctx, cfg.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch company", err.Error())
		return
	}
	cfg.Name = types.StringValue(c.Name)
	cfg.ParentID = numberToInt64(c.ParentID)
	cfg.Email = types.StringValue(c.Email)
	cfg.Level = numberToInt64(c.Level)
	cfg.DeviceCount = numberToInt64(c.DeviceCount)
	cfg.CreatedAt = timestampToString(c.CreatedAt)
	cfg.UpdatedAt = timestampToString(c.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
}

// ---- companies (list) ----

type companiesDataSource struct{ client *rms.Client }

type companiesModel struct {
	Companies types.List `tfsdk:"companies"`
}

func NewCompaniesDataSource() datasource.DataSource { return &companiesDataSource{} }

func (d *companiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rms_companies"
}

func (d *companiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFrom(req.ProviderData, &resp.Diagnostics)
}

func (d *companiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All companies visible to the current token.",
		Attributes: map[string]schema.Attribute{
			"companies": schema.ListAttribute{
				Computed:    true,
				ElementType: types.ObjectType{AttrTypes: companyObjectAttrs},
			},
		},
	}
}

func (d *companiesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, err := d.client.ListCompanies(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list companies", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, c := range list {
		obj, diags := companyObject(c)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	l, diags := types.ListValue(types.ObjectType{AttrTypes: companyObjectAttrs}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, companiesModel{Companies: l})...)
}
