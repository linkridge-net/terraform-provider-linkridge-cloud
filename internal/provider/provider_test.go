// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderSchemaHasExpectedAttributes(t *testing.T) {
	p := New("test")()
	resp := &frameworkprovider.SchemaResponse{}
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	if attrs == nil {
		t.Fatal("expected schema attributes")
	}

	baseURL, ok := attrs["base_url"]
	if !ok {
		t.Fatal("expected base_url attribute")
	}
	if !baseURL.IsOptional() {
		t.Fatal("expected base_url to be optional")
	}

	apiToken, ok := attrs["api_token"]
	if !ok {
		t.Fatal("expected api_token attribute")
	}
	if !apiToken.IsOptional() {
		t.Fatal("expected api_token to be optional")
	}
	if !apiToken.IsSensitive() {
		t.Fatal("expected api_token to be sensitive")
	}
}

func TestResolveProviderConfigEnvFallback(t *testing.T) {
	t.Setenv("LINKRIDGE_CLOUD_BASE_URL", "https://dev.cloud.linkridge.net")
	t.Setenv("LINKRIDGE_CLOUD_API_TOKEN", "env-token")

	cfg := resolveProviderConfig(ProviderModel{
		BaseURL:  types.StringNull(),
		APIToken: types.StringNull(),
	})

	if cfg.baseURL != "https://dev.cloud.linkridge.net" {
		t.Fatalf("unexpected base URL: %q", cfg.baseURL)
	}
	if cfg.apiToken != "env-token" {
		t.Fatalf("unexpected API token: %q", cfg.apiToken)
	}
}

func TestResolveProviderConfigHCLPriority(t *testing.T) {
	t.Setenv("LINKRIDGE_CLOUD_BASE_URL", "https://env.example.com")
	t.Setenv("LINKRIDGE_CLOUD_API_TOKEN", "env-token")

	cfg := resolveProviderConfig(ProviderModel{
		BaseURL:  types.StringValue("https://hcl.example.com"),
		APIToken: types.StringValue("hcl-token"),
	})

	if cfg.baseURL != "https://hcl.example.com" {
		t.Fatalf("unexpected base URL: %q", cfg.baseURL)
	}
	if cfg.apiToken != "hcl-token" {
		t.Fatalf("unexpected API token: %q", cfg.apiToken)
	}
}

func TestValidateRequiredConfig(t *testing.T) {
	var diags diag.Diagnostics
	validateRequiredConfig(resolvedConfig{}, &diags)

	if !diags.HasError() {
		t.Fatal("expected missing configuration errors")
	}
	if len(diags.Errors()) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(diags.Errors()))
	}
}

func TestProviderRegistersServicesDataSource(t *testing.T) {
	dataSources := (&LinkRidgeCloudProvider{}).DataSources(context.Background())
	if len(dataSources) != 16 {
		t.Fatalf("expected 16 data sources, got %d", len(dataSources))
	}
}

func TestStringValueOrNull(t *testing.T) {
	if !stringValueOrNull("").IsNull() {
		t.Fatal("expected empty string to become null")
	}
	if got := stringValueOrNull("value"); got.ValueString() != "value" {
		t.Fatalf("unexpected string value: %q", got.ValueString())
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "run-id", "fallback"); got != "run-id" {
		t.Fatalf("unexpected first non-empty value: %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("expected empty fallback, got %q", got)
	}
}

func TestStringFiltersDropsNullAndUnknownValues(t *testing.T) {
	filters := stringFilters(map[string]types.String{
		"id":     types.StringValue("acct_local_qr_demo"),
		"status": types.StringNull(),
	})

	if filters["id"] != "acct_local_qr_demo" {
		t.Fatalf("unexpected id filter: %q", filters["id"])
	}
	if _, ok := filters["status"]; ok {
		t.Fatal("expected null status filter to be omitted")
	}
}

func TestRawJSONValueOrNull(t *testing.T) {
	if !rawJSONValueOrNull(nil).IsNull() {
		t.Fatal("expected nil JSON to become null")
	}
	if !rawJSONValueOrNull([]byte("null")).IsNull() {
		t.Fatal("expected null JSON to become null")
	}
	if got := rawJSONValueOrNull([]byte(`{"z": true, "a": 1}`)); got.ValueString() != `{"a":1,"z":true}` {
		t.Fatalf("unexpected compact JSON value: %q", got.ValueString())
	}
}
