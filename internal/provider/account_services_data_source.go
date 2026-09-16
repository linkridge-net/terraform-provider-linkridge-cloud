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
	_ datasource.DataSource              = &AccountServicesDataSource{}
	_ datasource.DataSourceWithConfigure = &AccountServicesDataSource{}
)

// AccountServicesDataSource reads service planning records for an account.
type AccountServicesDataSource struct {
	client *linkridgecloud.Client
}

// NewAccountServicesDataSource returns an account services data source.
func NewAccountServicesDataSource() datasource.DataSource {
	return &AccountServicesDataSource{}
}

type accountServicesDataSourceModel struct {
	AccountID       types.String               `tfsdk:"account_id"`
	ID              types.String               `tfsdk:"id"`
	Status          types.String               `tfsdk:"status"`
	PlanKey         types.String               `tfsdk:"plan_key"`
	AccountServices []accountServiceStateModel `tfsdk:"account_services"`
}

type accountServiceStateModel struct {
	ID                     types.String `tfsdk:"id"`
	AccountID              types.String `tfsdk:"account_id"`
	ServiceID              types.String `tfsdk:"service_id"`
	PlanID                 types.String `tfsdk:"plan_id"`
	PlanKey                types.String `tfsdk:"plan_key"`
	Status                 types.String `tfsdk:"status"`
	ApprovalRequired       types.Bool   `tfsdk:"approval_required"`
	SourcePacketID         types.String `tfsdk:"source_packet_id"`
	ExternalEffectsEnabled types.Bool   `tfsdk:"external_effects_enabled"`
}

// Metadata sets the data source type name.
func (d *AccountServicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_services"
}

// Schema defines the data source schema.
func (d *AccountServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads LinkRidge Cloud account service planning records from the `/v1/accounts/{account_id}/services` control-plane API.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose service planning records should be read.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account service id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account service status filter, for example `planned`.",
			},
			"plan_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional service plan key filter, for example `starter`.",
			},
			"account_services": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Account service planning records returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Account service id.",
						},
						"account_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning account id.",
						},
						"service_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service id.",
						},
						"plan_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service plan id.",
						},
						"plan_key": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service plan key.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current account service status.",
						},
						"approval_required": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether Matthew/operator approval is required before customer-visible activation.",
						},
						"source_packet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Seed packet or draft source id.",
						},
						"external_effects_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether external customer effects are enabled for this account service.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *AccountServicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the account service state.
func (d *AccountServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accountServicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading account services.")
		return
	}

	accountServices, err := d.client.ListAccountServices(ctx, data.AccountID.ValueString(), stringFilters(map[string]types.String{
		"id":       data.ID,
		"status":   data.Status,
		"plan_key": data.PlanKey,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Account Services", err.Error())
		return
	}

	data.AccountServices = make([]accountServiceStateModel, 0, len(accountServices))
	for _, accountService := range accountServices {
		data.AccountServices = append(data.AccountServices, accountServiceStateModel{
			ID:                     types.StringValue(accountService.ID),
			AccountID:              types.StringValue(accountService.AccountID),
			ServiceID:              types.StringValue(accountService.ServiceID),
			PlanID:                 stringValueOrNull(accountService.PlanID),
			PlanKey:                stringValueOrNull(accountService.PlanKey),
			Status:                 types.StringValue(accountService.Status),
			ApprovalRequired:       types.BoolValue(accountService.ApprovalRequired),
			SourcePacketID:         stringValueOrNull(accountService.SourcePacketID),
			ExternalEffectsEnabled: types.BoolValue(accountService.ExternalEffectsEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
