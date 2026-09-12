// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the LinkRidge Cloud Terraform provider.
package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/linkridge-net/terraform-provider-linkridge-cloud/pkg/linkridgecloud"
)

var _ provider.Provider = &LinkRidgeCloudProvider{}

// LinkRidgeCloudProvider defines the provider implementation.
type LinkRidgeCloudProvider struct {
	version string
}

// ProviderModel describes provider configuration.
type ProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	APIToken types.String `tfsdk:"api_token"`
}

type resolvedConfig struct {
	baseURL  string
	apiToken string
}

// Metadata sets the provider type name and version.
func (p *LinkRidgeCloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "linkridgecloud"
	resp.Version = p.version
}

// Schema defines the provider configuration schema.
func (p *LinkRidgeCloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The LinkRidge Cloud provider manages draft and read-only LinkRidge Cloud control-plane resources.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "LinkRidge Cloud API base URL. Can also be set with `LINKRIDGE_CLOUD_BASE_URL`.",
			},
			"api_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "LinkRidge Cloud API bearer token. Can also be set with `LINKRIDGE_CLOUD_API_TOKEN`.",
			},
		},
	}
}

// Configure prepares the LinkRidge Cloud API client.
func (p *LinkRidgeCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := resolveProviderConfig(data)
	validateRequiredConfig(cfg, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := linkridgecloud.NewClient(linkridgecloud.ClientConfig{
		BaseURL:  cfg.baseURL,
		APIToken: cfg.apiToken,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create LinkRidge Cloud API Client",
			fmt.Sprintf("Provider configuration could not create a client: %s", err),
		)
		return
	}

	tflog.Info(ctx, "LinkRidge Cloud provider configured", map[string]interface{}{"base_url": client.BaseURL()})
	resp.DataSourceData = client
	resp.ResourceData = client
}

func resolveProviderConfig(data ProviderModel) resolvedConfig {
	return resolvedConfig{
		baseURL:  envOrValue(data.BaseURL, "LINKRIDGE_CLOUD_BASE_URL"),
		apiToken: envOrValue(data.APIToken, "LINKRIDGE_CLOUD_API_TOKEN"),
	}
}

func validateRequiredConfig(cfg resolvedConfig, diags *diag.Diagnostics) {
	if cfg.baseURL == "" {
		diags.AddError(
			"Missing LinkRidge Cloud Base URL",
			"Set the `base_url` provider attribute or the `LINKRIDGE_CLOUD_BASE_URL` environment variable.",
		)
	}
	if cfg.apiToken == "" {
		diags.AddError(
			"Missing LinkRidge Cloud API Token",
			"Set the `api_token` provider attribute or the `LINKRIDGE_CLOUD_API_TOKEN` environment variable.",
		)
	}
}

func envOrValue(val types.String, envVar string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envVar)
}

// Resources returns the list of resource types supported by this provider.
func (p *LinkRidgeCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// DataSources returns the list of data source types supported by this provider.
func (p *LinkRidgeCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewActivationPacketsDataSource,
		NewAccountServicesDataSource,
		NewAccountsDataSource,
		NewProvisioningRunsDataSource,
		NewQRImportJobsDataSource,
		NewQRWorkspacesDataSource,
		NewServiceTokensDataSource,
		NewServicesDataSource,
	}
}

// New returns a provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LinkRidgeCloudProvider{version: version}
	}
}
