// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linkridge-net/terraform-provider-linkridge-cloud/pkg/linkridgecloud"
)

var (
	_ datasource.DataSource              = &ServicesDataSource{}
	_ datasource.DataSourceWithConfigure = &ServicesDataSource{}
)

// ServicesDataSource reads the LinkRidge Cloud service catalog.
type ServicesDataSource struct {
	client *linkridgecloud.Client
}

// NewServicesDataSource returns a services data source.
func NewServicesDataSource() datasource.DataSource {
	return &ServicesDataSource{}
}

type servicesDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Services []serviceStateModel `tfsdk:"services"`
}

type serviceStateModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
	Plans       []string     `tfsdk:"plans"`
}

// Metadata sets the data source type name.
func (d *ServicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_services"
}

// Schema defines the data source schema.
func (d *ServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads LinkRidge Cloud services from the `/v1/services` control-plane API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional service id filter, for example `qr-codes`.",
			},
			"services": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Services returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service id.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Human-readable service name.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current service status from the control plane.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service description.",
						},
						"plans": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Plan ids supported by the service.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *ServicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*linkridgecloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *linkridgecloud.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the service catalog state.
func (d *ServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data servicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading services.")
		return
	}

	id := ""
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		id = data.ID.ValueString()
	}

	services, err := d.client.ListServices(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Services", err.Error())
		return
	}

	data.Services = make([]serviceStateModel, 0, len(services))
	for _, service := range services {
		data.Services = append(data.Services, serviceStateModel{
			ID:          types.StringValue(service.ID),
			Name:        types.StringValue(service.Name),
			Status:      types.StringValue(service.Status),
			Description: types.StringValue(service.Description),
			Plans:       service.Plans,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
