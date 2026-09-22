package provider

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/lombare/terraform-provider-teltonika-rms/internal/rms"
)

func clientFrom(providerData any, diags *diag.Diagnostics) *rms.Client {
	if providerData == nil {
		return nil
	}
	client, ok := providerData.(*rms.Client)
	if !ok {
		diags.AddError(
			"Unexpected provider data",
			fmt.Sprintf("expected *rms.Client, got %T", providerData),
		)
		return nil
	}
	return client
}

func handleNotFound(err error, diags *diag.Diagnostics, kind, id string) bool {
	if rms.IsNotFound(err) {
		return true
	}
	if err != nil {
		diags.AddError(fmt.Sprintf("Failed to read %s %s", kind, id), err.Error())
	}
	return false
}

func numberToInt64(n json.Number) types.Int64 {
	if n.String() == "" {
		return types.Int64Null()
	}
	v, err := n.Int64()
	if err != nil {
		f, err := n.Float64()
		if err != nil {
			return types.Int64Null()
		}
		return types.Int64Value(int64(f))
	}
	return types.Int64Value(v)
}

func numberToFloat64(n json.Number) types.Float64 {
	if n.String() == "" {
		return types.Float64Null()
	}
	v, err := n.Float64()
	if err != nil {
		return types.Float64Null()
	}
	return types.Float64Value(v)
}

func timestampToString(t rms.Timestamp) types.String {
	if t.Value == "" {
		return types.StringNull()
	}
	return types.StringValue(t.Value)
}

func rawToString(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func stringOr(v types.String, fallback string) string {
	if v.IsNull() || v.IsUnknown() {
		return fallback
	}
	return v.ValueString()
}

func int64Or(v types.Int64, fallback int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return fallback
	}
	return v.ValueInt64()
}

func int64SliceFromList(list types.List, diags *diag.Diagnostics) []int64 {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := list.Elements()
	out := make([]int64, 0, len(elems))
	for _, e := range elems {
		n, ok := e.(types.Int64)
		if !ok {
			diags.AddError("Type error", fmt.Sprintf("expected list of ints, got %T", e))
			return nil
		}
		out = append(out, n.ValueInt64())
	}
	return out
}

func listFromInt64sNumbers(values []json.Number) types.List {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		i, err := v.Int64()
		if err != nil {
			continue
		}
		elems = append(elems, types.Int64Value(i))
	}
	l, _ := types.ListValue(types.Int64Type, elems)
	return l
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
