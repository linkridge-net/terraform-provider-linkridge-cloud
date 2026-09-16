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
	_ datasource.DataSource              = &SupportCasesDataSource{}
	_ datasource.DataSourceWithConfigure = &SupportCasesDataSource{}
)

// SupportCasesDataSource reads safe support handoff metadata.
type SupportCasesDataSource struct {
	client *linkridgecloud.Client
}

// NewSupportCasesDataSource returns a support cases data source.
func NewSupportCasesDataSource() datasource.DataSource {
	return &SupportCasesDataSource{}
}

type supportCasesDataSourceModel struct {
	AccountID        types.String            `tfsdk:"account_id"`
	AccountServiceID types.String            `tfsdk:"account_service_id"`
	ID               types.String            `tfsdk:"id"`
	Category         types.String            `tfsdk:"category"`
	Severity         types.String            `tfsdk:"severity"`
	Status           types.String            `tfsdk:"status"`
	SupportCases     []supportCaseStateModel `tfsdk:"support_cases"`
}

type supportCaseStateModel struct {
	ID                       types.String   `tfsdk:"id"`
	AccountID                types.String   `tfsdk:"account_id"`
	AccountServiceID         types.String   `tfsdk:"account_service_id"`
	QRWorkspaceID            types.String   `tfsdk:"qr_workspace_id"`
	ServiceID                types.String   `tfsdk:"service_id"`
	TargetType               types.String   `tfsdk:"target_type"`
	TargetID                 types.String   `tfsdk:"target_id"`
	Category                 types.String   `tfsdk:"category"`
	Severity                 types.String   `tfsdk:"severity"`
	Status                   types.String   `tfsdk:"status"`
	Subject                  types.String   `tfsdk:"subject"`
	CreatedByUserID          types.String   `tfsdk:"created_by_user_id"`
	CreatedAt                types.String   `tfsdk:"created_at"`
	ResolvedAt               types.String   `tfsdk:"resolved_at"`
	SourcePacketID           types.String   `tfsdk:"source_packet_id"`
	CustomerVisible          types.Bool     `tfsdk:"customer_visible"`
	ExternalNotificationSent types.Bool     `tfsdk:"external_notification_sent"`
	ExternalTicketCreated    types.Bool     `tfsdk:"external_ticket_created"`
	MetadataQRCodeID         types.String   `tfsdk:"metadata_qr_code_id"`
	MetadataSlug             types.String   `tfsdk:"metadata_slug"`
	MetadataCustomerVisible  types.Bool     `tfsdk:"metadata_customer_visible"`
	ExternalTicketID         types.String   `tfsdk:"external_ticket_id"`
	MetadataNotificationSent types.Bool     `tfsdk:"metadata_external_notification_sent"`
	EscalationPerformed      types.Bool     `tfsdk:"escalation_performed"`
	MetadataApprovalRequired types.String   `tfsdk:"metadata_approval_required"`
	MetadataBlockedReason    types.String   `tfsdk:"metadata_blocked_reason"`
	BlockedExternalActions   []types.String `tfsdk:"blocked_external_actions"`
}

func (d *SupportCasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_support_cases"
}

func (d *SupportCasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud support handoff metadata from the `/v1/accounts/{account_id}/services/{account_service_id}/support-cases` control-plane API. This data source does not create external support tickets, notify customers, escalate incidents, publish customer timelines, or change service status.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose support handoff records should be read.",
			},
			"account_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account service id whose support handoff records should be read.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional support case id filter.",
			},
			"category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional support category filter, for example `import_review`.",
			},
			"severity": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional support severity filter, for example `low` or `medium`.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional support case status filter, for example `blocked`.",
			},
			"support_cases": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe support handoff metadata returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                                  schema.StringAttribute{Computed: true, MarkdownDescription: "Support case id."},
						"account_id":                          schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"account_service_id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id."},
						"qr_workspace_id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "QR workspace associated with the case."},
						"service_id":                          schema.StringAttribute{Computed: true, MarkdownDescription: "Service id."},
						"target_type":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Service object type under review."},
						"target_id":                           schema.StringAttribute{Computed: true, MarkdownDescription: "Service object id under review, when present."},
						"category":                            schema.StringAttribute{Computed: true, MarkdownDescription: "Support category."},
						"severity":                            schema.StringAttribute{Computed: true, MarkdownDescription: "Local triage severity."},
						"status":                              schema.StringAttribute{Computed: true, MarkdownDescription: "Current support case status."},
						"subject":                             schema.StringAttribute{Computed: true, MarkdownDescription: "Operator-readable support case summary."},
						"created_by_user_id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "User id that prepared the local case."},
						"created_at":                          schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp, when the case is recorded beyond fixture evidence."},
						"resolved_at":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Resolution timestamp, when the case is resolved."},
						"source_packet_id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
						"customer_visible":                    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the case is visible to customers."},
						"external_notification_sent":          schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether an external customer notification has been sent."},
						"external_ticket_created":             schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether an external support ticket has been created."},
						"metadata_qr_code_id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Related QR code id from metadata, when present."},
						"metadata_slug":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Related QR slug from metadata, when present."},
						"metadata_customer_visible":           schema.BoolAttribute{Computed: true, MarkdownDescription: "Metadata-level customer visibility flag."},
						"external_ticket_id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "External support ticket id, when one exists."},
						"metadata_external_notification_sent": schema.BoolAttribute{Computed: true, MarkdownDescription: "Metadata-level external notification flag."},
						"escalation_performed":                schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether escalation has been performed."},
						"metadata_approval_required":          schema.StringAttribute{Computed: true, MarkdownDescription: "Operator approval required before external support effects."},
						"metadata_blocked_reason":             schema.StringAttribute{Computed: true, MarkdownDescription: "Reason external support effects remain blocked."},
						"blocked_external_actions":            schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "External support actions blocked by the approval gate."},
					},
				},
			},
		},
	}
}

func (d *SupportCasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SupportCasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data supportCasesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading support cases.")
		return
	}

	supportCases, err := d.client.ListSupportCases(ctx, data.AccountID.ValueString(), data.AccountServiceID.ValueString(), stringFilters(map[string]types.String{
		"id":       data.ID,
		"category": data.Category,
		"severity": data.Severity,
		"status":   data.Status,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Support Cases", err.Error())
		return
	}

	data.SupportCases = make([]supportCaseStateModel, 0, len(supportCases))
	for _, supportCase := range supportCases {
		data.SupportCases = append(data.SupportCases, supportCaseStateModel{
			ID:                       types.StringValue(supportCase.ID),
			AccountID:                types.StringValue(supportCase.AccountID),
			AccountServiceID:         types.StringValue(supportCase.AccountServiceID),
			QRWorkspaceID:            stringValueOrNull(supportCase.QRWorkspaceID),
			ServiceID:                stringValueOrNull(supportCase.ServiceID),
			TargetType:               stringValueOrNull(supportCase.TargetType),
			TargetID:                 stringValueOrNull(supportCase.TargetID),
			Category:                 stringValueOrNull(supportCase.Category),
			Severity:                 stringValueOrNull(supportCase.Severity),
			Status:                   types.StringValue(supportCase.Status),
			Subject:                  stringValueOrNull(supportCase.Subject),
			CreatedByUserID:          stringValueOrNull(supportCase.CreatedByUserID),
			CreatedAt:                stringValueOrNull(supportCase.CreatedAt),
			ResolvedAt:               stringValueOrNull(supportCase.ResolvedAt),
			SourcePacketID:           stringValueOrNull(supportCase.SourcePacketID),
			CustomerVisible:          types.BoolValue(supportCase.CustomerVisible),
			ExternalNotificationSent: types.BoolValue(supportCase.ExternalNotificationSent),
			ExternalTicketCreated:    types.BoolValue(supportCase.ExternalTicketCreated),
			MetadataQRCodeID:         stringValueOrNull(supportCase.Metadata.QRCodeID),
			MetadataSlug:             stringValueOrNull(supportCase.Metadata.Slug),
			MetadataCustomerVisible:  types.BoolValue(supportCase.Metadata.CustomerVisible),
			ExternalTicketID:         stringValueOrNull(supportCase.Metadata.ExternalTicketID),
			MetadataNotificationSent: types.BoolValue(supportCase.Metadata.ExternalNotificationSent),
			EscalationPerformed:      types.BoolValue(supportCase.Metadata.EscalationPerformed),
			MetadataApprovalRequired: stringValueOrNull(supportCase.Metadata.ApprovalRequired),
			MetadataBlockedReason:    stringValueOrNull(supportCase.Metadata.BlockedReason),
			BlockedExternalActions:   stringSliceValues(supportCase.Metadata.BlockedExternalActions),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
