// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func stringFilters(values map[string]types.String) map[string]string {
	filters := make(map[string]string, len(values))
	for key, value := range values {
		if !value.IsNull() && !value.IsUnknown() {
			filters[key] = value.ValueString()
		}
	}
	return filters
}
