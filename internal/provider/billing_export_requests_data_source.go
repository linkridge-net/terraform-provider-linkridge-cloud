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
	_ datasource.DataSource              = &BillingExportRequestsDataSource{}
	_ datasource.DataSourceWithConfigure = &BillingExportRequestsDataSource{}
)

// BillingExportRequestsDataSource reads safe billing export review metadata.
type BillingExportRequestsDataSource struct {
	client *linkridgecloud.Client
}

// NewBillingExportRequestsDataSource returns a billing export requests data source.
func NewBillingExportRequestsDataSource() datasource.DataSource {
	return &BillingExportRequestsDataSource{}
}

type billingExportRequestsDataSourceModel struct {
	AccountID             types.String                     `tfsdk:"account_id"`
	AccountServiceID      types.String                     `tfsdk:"account_service_id"`
	ID                    types.String                     `tfsdk:"id"`
	Status                types.String                     `tfsdk:"status"`
	BillingExportRequests []billingExportRequestStateModel `tfsdk:"billing_export_requests"`
}

type billingExportRequestStateModel struct {
	ID                       types.String   `tfsdk:"id"`
	AccountID                types.String   `tfsdk:"account_id"`
	AccountServiceID         types.String   `tfsdk:"account_service_id"`
	ServiceID                types.String   `tfsdk:"service_id"`
	ScanEventIDs             []types.String `tfsdk:"scan_event_ids"`
	Quantity                 types.Int64    `tfsdk:"quantity"`
	Billable                 types.Bool     `tfsdk:"billable"`
	Status                   types.String   `tfsdk:"status"`
	RequestedByUserID        types.String   `tfsdk:"requested_by_user_id"`
	RequestedAt              types.String   `tfsdk:"requested_at"`
	ExportedAt               types.String   `tfsdk:"exported_at"`
	SourcePacketID           types.String   `tfsdk:"source_packet_id"`
	ExternalExportEnabled    types.Bool     `tfsdk:"external_export_enabled"`
	BillingCustomerID        types.String   `tfsdk:"billing_customer_id"`
	BillingSubscriptionID    types.String   `tfsdk:"billing_subscription_id"`
	ExternalUsageRecordID    types.String   `tfsdk:"external_usage_record_id"`
	ExportDestination        types.String   `tfsdk:"export_destination"`
	MetadataApprovalRequired types.String   `tfsdk:"metadata_approval_required"`
	MetadataBlockedReason    types.String   `tfsdk:"metadata_blocked_reason"`
}

func (d *BillingExportRequestsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_export_requests"
}

func (d *BillingExportRequestsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud billing export review metadata from the `/v1/accounts/{account_id}/services/{account_service_id}/billing-export-requests` control-plane API. This data source does not create billing customers, subscriptions, invoices, or metered usage records.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose billing export review records should be read.",
			},
			"account_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account service id whose billing export review records should be read.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional billing export request id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional billing export request status filter, for example `blocked`.",
			},
			"billing_export_requests": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe billing export review metadata returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Billing export request id."},
						"account_id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"account_service_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id."},
						"service_id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Service id."},
						"scan_event_ids":             schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Local scan usage events grouped by this review record."},
						"quantity":                   schema.Int64Attribute{Computed: true, MarkdownDescription: "Total local usage quantity represented by the review record."},
						"billable":                   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the usage is billable. Dev planning records should remain false until billing integration is approved."},
						"status":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Current billing export request status."},
						"requested_by_user_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "User id that prepared the local review request."},
						"requested_at":               schema.StringAttribute{Computed: true, MarkdownDescription: "Request timestamp, when the request is recorded beyond fixture evidence."},
						"exported_at":                schema.StringAttribute{Computed: true, MarkdownDescription: "Export timestamp. Dev planning records should remain null."},
						"source_packet_id":           schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
						"external_export_enabled":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether external billing export effects are enabled."},
						"billing_customer_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Billing customer id, when an approved billing integration exists."},
						"billing_subscription_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Billing subscription id, when an approved billing integration exists."},
						"external_usage_record_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "External usage record id, when export has been approved and performed."},
						"export_destination":         schema.StringAttribute{Computed: true, MarkdownDescription: "External export destination, when configured."},
						"metadata_approval_required": schema.StringAttribute{Computed: true, MarkdownDescription: "Operator approval required before external billing export."},
						"metadata_blocked_reason":    schema.StringAttribute{Computed: true, MarkdownDescription: "Reason external billing export remains blocked."},
					},
				},
			},
		},
	}
}

func (d *BillingExportRequestsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BillingExportRequestsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data billingExportRequestsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading billing export requests.")
		return
	}

	requests, err := d.client.ListBillingExportRequests(ctx, data.AccountID.ValueString(), data.AccountServiceID.ValueString(), stringFilters(map[string]types.String{
		"id":     data.ID,
		"status": data.Status,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Billing Export Requests", err.Error())
		return
	}

	data.BillingExportRequests = make([]billingExportRequestStateModel, 0, len(requests))
	for _, request := range requests {
		data.BillingExportRequests = append(data.BillingExportRequests, billingExportRequestStateModel{
			ID:                       types.StringValue(request.ID),
			AccountID:                types.StringValue(request.AccountID),
			AccountServiceID:         types.StringValue(request.AccountServiceID),
			ServiceID:                types.StringValue(request.ServiceID),
			ScanEventIDs:             stringSliceValues(request.ScanEventIDs),
			Quantity:                 types.Int64Value(request.Quantity),
			Billable:                 types.BoolValue(request.Billable),
			Status:                   types.StringValue(request.Status),
			RequestedByUserID:        stringValueOrNull(request.RequestedByUserID),
			RequestedAt:              stringValueOrNull(request.RequestedAt),
			ExportedAt:               stringValueOrNull(request.ExportedAt),
			SourcePacketID:           stringValueOrNull(request.SourcePacketID),
			ExternalExportEnabled:    types.BoolValue(request.ExternalExportEnabled),
			BillingCustomerID:        stringValueOrNull(request.Metadata.BillingCustomerID),
			BillingSubscriptionID:    stringValueOrNull(request.Metadata.BillingSubscriptionID),
			ExternalUsageRecordID:    stringValueOrNull(request.Metadata.ExternalUsageRecordID),
			ExportDestination:        stringValueOrNull(request.Metadata.ExportDestination),
			MetadataApprovalRequired: stringValueOrNull(request.Metadata.ApprovalRequired),
			MetadataBlockedReason:    stringValueOrNull(request.Metadata.BlockedReason),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
