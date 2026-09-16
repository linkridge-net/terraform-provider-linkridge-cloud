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
	_ datasource.DataSource              = &ServiceTokensDataSource{}
	_ datasource.DataSourceWithConfigure = &ServiceTokensDataSource{}
)

// ServiceTokensDataSource reads safe service-token metadata for an account.
type ServiceTokensDataSource struct {
	client *linkridgecloud.Client
}

// NewServiceTokensDataSource returns a service tokens data source.
func NewServiceTokensDataSource() datasource.DataSource {
	return &ServiceTokensDataSource{}
}

type serviceTokensDataSourceModel struct {
	AccountID            types.String             `tfsdk:"account_id"`
	ID                   types.String             `tfsdk:"id"`
	Status               types.String             `tfsdk:"status"`
	SecretMaterialIssued types.Bool               `tfsdk:"secret_material_issued"`
	ServiceTokens        []serviceTokenStateModel `tfsdk:"service_tokens"`
}

type serviceTokenStateModel struct {
	ID                     types.String   `tfsdk:"id"`
	AccountID              types.String   `tfsdk:"account_id"`
	AccountServiceID       types.String   `tfsdk:"account_service_id"`
	ServiceID              types.String   `tfsdk:"service_id"`
	CreatedByUserID        types.String   `tfsdk:"created_by_user_id"`
	Name                   types.String   `tfsdk:"name"`
	Scopes                 []types.String `tfsdk:"scopes"`
	Status                 types.String   `tfsdk:"status"`
	SecretMaterialIssued   types.Bool     `tfsdk:"secret_material_issued"`
	ApprovalRequired       types.String   `tfsdk:"approval_required"`
	SourcePacketID         types.String   `tfsdk:"source_packet_id"`
	ExternalEffectsEnabled types.Bool     `tfsdk:"external_effects_enabled"`
}

// Metadata sets the data source type name.
func (d *ServiceTokensDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_tokens"
}

// Schema defines the data source schema.
func (d *ServiceTokensDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud service-token metadata from the `/v1/accounts/{account_id}/service-tokens` control-plane API. Secret material is never returned by this data source.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose service-token metadata should be read.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional service-token metadata id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional service-token status filter, for example `planned`.",
			},
			"secret_material_issued": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Optional filter for whether secret material has been issued. Dev planning records should remain `false`.",
			},
			"service_tokens": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe service-token metadata returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service-token metadata id.",
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
						"created_by_user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User id that requested the token metadata.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Display name for the service-token request.",
						},
						"scopes": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Approved or requested token scopes.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current service-token metadata status.",
						},
						"secret_material_issued": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether secret material has been issued. Planning records should remain false until explicitly approved outside Terraform.",
						},
						"approval_required": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Operator approval required before secret issuance.",
						},
						"source_packet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Seed packet or draft source id.",
						},
						"external_effects_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether external customer effects are enabled for this token request.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *ServiceTokensDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the service-token metadata state.
func (d *ServiceTokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceTokensDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading service tokens.")
		return
	}

	filters := stringFilters(map[string]types.String{
		"id":     data.ID,
		"status": data.Status,
	})
	if !data.SecretMaterialIssued.IsNull() && !data.SecretMaterialIssued.IsUnknown() {
		filters["secret_material_issued"] = fmt.Sprintf("%t", data.SecretMaterialIssued.ValueBool())
	}

	serviceTokens, err := d.client.ListServiceTokens(ctx, data.AccountID.ValueString(), filters)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Service Tokens", err.Error())
		return
	}

	data.ServiceTokens = make([]serviceTokenStateModel, 0, len(serviceTokens))
	for _, serviceToken := range serviceTokens {
		data.ServiceTokens = append(data.ServiceTokens, serviceTokenStateModel{
			ID:                     types.StringValue(serviceToken.ID),
			AccountID:              types.StringValue(serviceToken.AccountID),
			AccountServiceID:       types.StringValue(serviceToken.AccountServiceID),
			ServiceID:              types.StringValue(serviceToken.ServiceID),
			CreatedByUserID:        stringValueOrNull(serviceToken.CreatedByUserID),
			Name:                   types.StringValue(serviceToken.Name),
			Scopes:                 stringSliceValues(serviceToken.Scopes),
			Status:                 types.StringValue(serviceToken.Status),
			SecretMaterialIssued:   types.BoolValue(serviceToken.SecretMaterialIssued),
			ApprovalRequired:       stringValueOrNull(serviceToken.ApprovalRequired),
			SourcePacketID:         stringValueOrNull(serviceToken.SourcePacketID),
			ExternalEffectsEnabled: types.BoolValue(serviceToken.ExternalEffectsEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
