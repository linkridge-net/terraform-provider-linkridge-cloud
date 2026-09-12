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
	_ datasource.DataSource              = &AccountsDataSource{}
	_ datasource.DataSourceWithConfigure = &AccountsDataSource{}
)

// AccountsDataSource reads LinkRidge Cloud account planning records.
type AccountsDataSource struct {
	client *linkridgecloud.Client
}

// NewAccountsDataSource returns an accounts data source.
func NewAccountsDataSource() datasource.DataSource {
	return &AccountsDataSource{}
}

type accountsDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Status   types.String        `tfsdk:"status"`
	Accounts []accountStateModel `tfsdk:"accounts"`
}

type accountStateModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Status                 types.String `tfsdk:"status"`
	PrimaryOwnerUserID     types.String `tfsdk:"primary_owner_user_id"`
	BillingCustomerID      types.String `tfsdk:"billing_customer_id"`
	SourcePacketID         types.String `tfsdk:"source_packet_id"`
	ExternalEffectsEnabled types.Bool   `tfsdk:"external_effects_enabled"`
}

// Metadata sets the data source type name.
func (d *AccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accounts"
}

// Schema defines the data source schema.
func (d *AccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads LinkRidge Cloud account planning records from the `/v1/accounts` control-plane API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account status filter, for example `draft`.",
			},
			"accounts": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Account planning records returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Account id.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Account display name.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current account status.",
						},
						"primary_owner_user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Primary owner user id when available.",
						},
						"billing_customer_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Billing customer id when an approved external billing record exists.",
						},
						"source_packet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Seed packet or draft source id.",
						},
						"external_effects_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether external customer effects are enabled for this account.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *AccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the account state.
func (d *AccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading accounts.")
		return
	}

	id := ""
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		id = data.ID.ValueString()
	}
	status := ""
	if !data.Status.IsNull() && !data.Status.IsUnknown() {
		status = data.Status.ValueString()
	}

	accounts, err := d.client.ListAccounts(ctx, id, status)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Accounts", err.Error())
		return
	}

	data.Accounts = make([]accountStateModel, 0, len(accounts))
	for _, account := range accounts {
		data.Accounts = append(data.Accounts, accountStateModel{
			ID:                     types.StringValue(account.ID),
			Name:                   types.StringValue(account.Name),
			Status:                 types.StringValue(account.Status),
			PrimaryOwnerUserID:     stringValueOrNull(account.PrimaryOwnerUserID),
			BillingCustomerID:      stringValueOrNull(account.BillingCustomerID),
			SourcePacketID:         stringValueOrNull(account.SourcePacketID),
			ExternalEffectsEnabled: types.BoolValue(account.ExternalEffectsEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
