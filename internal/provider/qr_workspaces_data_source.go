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
	_ datasource.DataSource              = &QRWorkspacesDataSource{}
	_ datasource.DataSourceWithConfigure = &QRWorkspacesDataSource{}
)

// QRWorkspacesDataSource reads LinkRidge Cloud QR workspaces.
type QRWorkspacesDataSource struct {
	client *linkridgecloud.Client
}

// NewQRWorkspacesDataSource returns a QR workspaces data source.
func NewQRWorkspacesDataSource() datasource.DataSource {
	return &QRWorkspacesDataSource{}
}

type qrWorkspacesDataSourceModel struct {
	ID               types.String            `tfsdk:"id"`
	AccountID        types.String            `tfsdk:"account_id"`
	AccountServiceID types.String            `tfsdk:"account_service_id"`
	Status           types.String            `tfsdk:"status"`
	Workspaces       []qrWorkspaceStateModel `tfsdk:"workspaces"`
}

type qrWorkspaceStateModel struct {
	ID                       types.String `tfsdk:"id"`
	AccountID                types.String `tfsdk:"account_id"`
	AccountServiceID         types.String `tfsdk:"account_service_id"`
	ServiceID                types.String `tfsdk:"service_id"`
	PlanKey                  types.String `tfsdk:"plan_key"`
	Status                   types.String `tfsdk:"status"`
	SourcePacketID           types.String `tfsdk:"source_packet_id"`
	ExternalRedirectsEnabled types.Bool   `tfsdk:"external_redirects_enabled"`
}

// Metadata sets the data source type name.
func (d *QRWorkspacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qr_workspaces"
}

// Schema defines the data source schema.
func (d *QRWorkspacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads LinkRidge Cloud QR workspaces from the `/v1/qr/workspaces` control-plane API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional QR workspace id filter.",
			},
			"account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account id filter.",
			},
			"account_service_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account service id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional QR workspace status filter, for example `planned`.",
			},
			"workspaces": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "QR workspaces returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "QR workspace id.",
						},
						"account_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning account id.",
						},
						"account_service_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning account service id.",
						},
						"service_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service id.",
						},
						"plan_key": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service plan key.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current QR workspace status.",
						},
						"source_packet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Seed packet or draft source id.",
						},
						"external_redirects_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether hosted QR redirects are enabled for this workspace.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *QRWorkspacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the QR workspace state.
func (d *QRWorkspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data qrWorkspacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading QR workspaces.")
		return
	}

	workspaces, err := d.client.ListQRWorkspaces(ctx, stringFilters(map[string]types.String{
		"id":                 data.ID,
		"account_id":         data.AccountID,
		"account_service_id": data.AccountServiceID,
		"status":             data.Status,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud QR Workspaces", err.Error())
		return
	}

	data.Workspaces = make([]qrWorkspaceStateModel, 0, len(workspaces))
	for _, workspace := range workspaces {
		data.Workspaces = append(data.Workspaces, qrWorkspaceStateModel{
			ID:                       types.StringValue(workspace.ID),
			AccountID:                types.StringValue(workspace.AccountID),
			AccountServiceID:         types.StringValue(workspace.AccountServiceID),
			ServiceID:                types.StringValue(workspace.ServiceID),
			PlanKey:                  types.StringValue(workspace.PlanKey),
			Status:                   types.StringValue(workspace.Status),
			SourcePacketID:           stringValueOrNull(workspace.SourcePacketID),
			ExternalRedirectsEnabled: types.BoolValue(workspace.ExternalRedirectsEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
