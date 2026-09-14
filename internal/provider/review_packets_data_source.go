// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linkridge-net/terraform-provider-linkridge-cloud/pkg/linkridgecloud"
)

var (
	_ datasource.DataSource              = &ReviewPacketsDataSource{}
	_ datasource.DataSourceWithConfigure = &ReviewPacketsDataSource{}
)

// ReviewPacketsDataSource reads normalized internal review evidence.
type ReviewPacketsDataSource struct {
	client *linkridgecloud.Client
}

// NewReviewPacketsDataSource returns a review packets data source.
func NewReviewPacketsDataSource() datasource.DataSource {
	return &ReviewPacketsDataSource{}
}

type reviewPacketsDataSourceModel struct {
	ID               types.String             `tfsdk:"id"`
	AccountID        types.String             `tfsdk:"account_id"`
	AccountServiceID types.String             `tfsdk:"account_service_id"`
	PacketType       types.String             `tfsdk:"packet_type"`
	Status           types.String             `tfsdk:"status"`
	ReviewPackets    []reviewPacketStateModel `tfsdk:"review_packets"`
}

type reviewPacketStateModel struct {
	ID                     types.String   `tfsdk:"id"`
	AccountID              types.String   `tfsdk:"account_id"`
	AccountServiceID       types.String   `tfsdk:"account_service_id"`
	ServiceID              types.String   `tfsdk:"service_id"`
	PacketType             types.String   `tfsdk:"packet_type"`
	Status                 types.String   `tfsdk:"status"`
	ApprovalRequired       types.String   `tfsdk:"approval_required"`
	RequestedBySubject     types.String   `tfsdk:"requested_by_subject"`
	RequestedByUserID      types.String   `tfsdk:"requested_by_user_id"`
	ReviewChecksJSON       types.String   `tfsdk:"review_checks_json"`
	BlockedExternalActions []types.String `tfsdk:"blocked_external_actions"`
	SourcePacketID         types.String   `tfsdk:"source_packet_id"`
}

func (d *ReviewPacketsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_review_packets"
}

func (d *ReviewPacketsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads normalized internal LinkRidge Cloud review packets from `/v1/review-packets`. This data source exposes approval-gate evidence only; it does not approve work, execute imports, issue secrets, export billing usage, create support tickets, notify customers, or enable customer-visible effects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional review packet id filter.",
			},
			"account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account id filter.",
			},
			"account_service_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account service id filter.",
			},
			"packet_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional packet type filter, for example `service_token`, `import`, `activation`, `billing_export`, or `support`.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional review packet status filter, for example `blocked_pending_matthew_approval`.",
			},
			"review_packets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe normalized review packet evidence returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Review packet id."},
						"account_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"account_service_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id."},
						"service_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Service id."},
						"packet_type":              schema.StringAttribute{Computed: true, MarkdownDescription: "Review packet type."},
						"status":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Current review packet status."},
						"approval_required":        schema.StringAttribute{Computed: true, MarkdownDescription: "Operator approval required before external effects."},
						"requested_by_subject":     schema.StringAttribute{Computed: true, MarkdownDescription: "Authenticated actor subject that prepared the packet, when present."},
						"requested_by_user_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "User id that requested or prepared the packet, when present."},
						"review_checks_json":       schema.StringAttribute{Computed: true, MarkdownDescription: "Packet-specific review checks as compact JSON for operator inspection."},
						"blocked_external_actions": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "External actions blocked by the approval gate."},
						"source_packet_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
					},
				},
			},
		},
	}
}

func (d *ReviewPacketsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ReviewPacketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data reviewPacketsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading review packets.")
		return
	}

	reviewPackets, err := d.client.ListReviewPackets(ctx, stringFilters(map[string]types.String{
		"id":                 data.ID,
		"account_id":         data.AccountID,
		"account_service_id": data.AccountServiceID,
		"packet_type":        data.PacketType,
		"status":             data.Status,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Review Packets", err.Error())
		return
	}

	data.ReviewPackets = make([]reviewPacketStateModel, 0, len(reviewPackets))
	for _, packet := range reviewPackets {
		data.ReviewPackets = append(data.ReviewPackets, reviewPacketStateModel{
			ID:                     types.StringValue(packet.ID),
			AccountID:              types.StringValue(packet.AccountID),
			AccountServiceID:       stringValueOrNull(packet.AccountServiceID),
			ServiceID:              stringValueOrNull(packet.ServiceID),
			PacketType:             types.StringValue(packet.PacketType),
			Status:                 types.StringValue(packet.Status),
			ApprovalRequired:       stringValueOrNull(packet.ApprovalRequired),
			RequestedBySubject:     stringValueOrNull(packet.RequestedByActor.Subject),
			RequestedByUserID:      stringValueOrNull(packet.RequestedByActor.UserID),
			ReviewChecksJSON:       rawJSONValueOrNull(packet.ReviewChecks),
			BlockedExternalActions: stringSliceValues(packet.BlockedExternalActions),
			SourcePacketID:         stringValueOrNull(packet.SourcePacketID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func rawJSONValueOrNull(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		return types.StringNull()
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return types.StringValue(string(raw))
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return types.StringValue(string(raw))
	}
	return types.StringValue(string(encoded))
}
