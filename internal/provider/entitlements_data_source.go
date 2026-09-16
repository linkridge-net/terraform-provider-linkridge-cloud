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
	_ datasource.DataSource              = &EntitlementsDataSource{}
	_ datasource.DataSourceWithConfigure = &EntitlementsDataSource{}
)

// EntitlementsDataSource reads effective entitlement evidence for an account service.
type EntitlementsDataSource struct {
	client *linkridgecloud.Client
}

// NewEntitlementsDataSource returns an entitlements data source.
func NewEntitlementsDataSource() datasource.DataSource {
	return &EntitlementsDataSource{}
}

type entitlementsDataSourceModel struct {
	AccountID        types.String            `tfsdk:"account_id"`
	AccountServiceID types.String            `tfsdk:"account_service_id"`
	EntitlementKey   types.String            `tfsdk:"entitlement_key"`
	Entitlements     []entitlementStateModel `tfsdk:"entitlements"`
}

type entitlementStateModel struct {
	ID                types.String `tfsdk:"id"`
	AccountID         types.String `tfsdk:"account_id"`
	AccountServiceID  types.String `tfsdk:"account_service_id"`
	ServiceID         types.String `tfsdk:"service_id"`
	EntitlementKey    types.String `tfsdk:"entitlement_key"`
	Kind              types.String `tfsdk:"kind"`
	ValueJSON         types.String `tfsdk:"value_json"`
	BillingSyncStatus types.String `tfsdk:"billing_sync_status"`
	SourcePacketID    types.String `tfsdk:"source_packet_id"`
}

// Metadata sets the data source type name.
func (d *EntitlementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entitlements"
}

// Schema defines the data source schema.
func (d *EntitlementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads effective LinkRidge Cloud entitlement evidence from the `/v1/accounts/{account_id}/services/{account_service_id}/entitlements` control-plane API. This data source is read-only and does not sync billing, change feature gates, or enable customer-visible service access.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose account-service entitlements should be read.",
			},
			"account_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account service id whose effective entitlements should be read.",
			},
			"entitlement_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional entitlement key filter, for example `active_qr_codes`.",
			},
			"entitlements": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Effective entitlement records returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Entitlement id when the backing store provides one."},
						"account_id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"account_service_id":  schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id."},
						"service_id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Service id."},
						"entitlement_key":     schema.StringAttribute{Computed: true, MarkdownDescription: "Entitlement key."},
						"kind":                schema.StringAttribute{Computed: true, MarkdownDescription: "Entitlement kind, such as `limit`, `metered-limit`, or `feature`."},
						"value_json":          schema.StringAttribute{Computed: true, MarkdownDescription: "Entitlement value as compact JSON so numbers, booleans, strings, and future structured values round-trip safely."},
						"billing_sync_status": schema.StringAttribute{Computed: true, MarkdownDescription: "Current local billing-sync status for the entitlement."},
						"source_packet_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *EntitlementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the entitlement state.
func (d *EntitlementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data entitlementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading entitlements.")
		return
	}

	entitlements, err := d.client.ListEntitlements(ctx, data.AccountID.ValueString(), data.AccountServiceID.ValueString(), stringFilters(map[string]types.String{
		"entitlement_key": data.EntitlementKey,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Entitlements", err.Error())
		return
	}

	data.Entitlements = make([]entitlementStateModel, 0, len(entitlements))
	for _, entitlement := range entitlements {
		data.Entitlements = append(data.Entitlements, entitlementStateModel{
			ID:                stringValueOrNull(entitlement.ID),
			AccountID:         types.StringValue(entitlement.AccountID),
			AccountServiceID:  types.StringValue(entitlement.AccountServiceID),
			ServiceID:         types.StringValue(entitlement.ServiceID),
			EntitlementKey:    types.StringValue(entitlement.EntitlementKey),
			Kind:              types.StringValue(entitlement.Kind),
			ValueJSON:         rawJSONValueOrNull(entitlement.Value),
			BillingSyncStatus: stringValueOrNull(entitlement.BillingSyncStatus),
			SourcePacketID:    stringValueOrNull(entitlement.SourcePacketID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
