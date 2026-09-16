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

func stringSliceValues(values []string) []types.String {
	result := make([]types.String, 0, len(values))
	for _, value := range values {
		result = append(result, types.StringValue(value))
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
