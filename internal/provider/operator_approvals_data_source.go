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
	_ datasource.DataSource              = &OperatorApprovalsDataSource{}
	_ datasource.DataSourceWithConfigure = &OperatorApprovalsDataSource{}
)

// OperatorApprovalsDataSource reads safe operator decision evidence.
type OperatorApprovalsDataSource struct {
	client *linkridgecloud.Client
}

// NewOperatorApprovalsDataSource returns an operator approvals data source.
func NewOperatorApprovalsDataSource() datasource.DataSource {
	return &OperatorApprovalsDataSource{}
}

type operatorApprovalsDataSourceModel struct {
	AccountID         types.String                 `tfsdk:"account_id"`
	AccountServiceID  types.String                 `tfsdk:"account_service_id"`
	ServiceID         types.String                 `tfsdk:"service_id"`
	DecisionID        types.String                 `tfsdk:"decision_id"`
	DecisionState     types.String                 `tfsdk:"decision_state"`
	RequiredApproval  types.String                 `tfsdk:"required_approval"`
	SourcePacketID    types.String                 `tfsdk:"source_packet_id"`
	OperatorApprovals []operatorApprovalStateModel `tfsdk:"operator_approvals"`
}

type operatorApprovalStateModel struct {
	Status           types.String   `tfsdk:"status"`
	DecisionID       types.String   `tfsdk:"decision_id"`
	SourcePacketID   types.String   `tfsdk:"source_packet_id"`
	AccountID        types.String   `tfsdk:"account_id"`
	AccountServiceID types.String   `tfsdk:"account_service_id"`
	ServiceID        types.String   `tfsdk:"service_id"`
	QRWorkspaceID    types.String   `tfsdk:"qr_workspace_id"`
	ReviewerUserID   types.String   `tfsdk:"reviewer_user_id"`
	DecidedAt        types.String   `tfsdk:"decided_at"`
	DecisionState    types.String   `tfsdk:"decision_state"`
	RequiredApproval types.String   `tfsdk:"required_approval"`
	ApprovedActions  []types.String `tfsdk:"approved_actions"`
	BlockedActions   []types.String `tfsdk:"blocked_actions"`
	AuditEventAction types.String   `tfsdk:"audit_event_action"`
	Result           types.String   `tfsdk:"result"`
}

func (d *OperatorApprovalsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operator_approvals"
}

func (d *OperatorApprovalsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud operator approval decisions from `/v1/operator-approvals`. This data source exposes approval-gate evidence only; it does not approve, reject, execute provisioning, issue secrets, send invites, export billing usage, create support tickets, configure DNS, deploy production changes, or enable customer-visible effects.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account id filter.",
			},
			"account_service_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account service id filter.",
			},
			"service_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional service id filter, for example `qr-codes`.",
			},
			"decision_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional operator decision id filter.",
			},
			"decision_state": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional decision state filter, for example `blocked`.",
			},
			"required_approval": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional required approval filter, for example `matthew`.",
			},
			"source_packet_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional seed packet or draft source id filter.",
			},
			"operator_approvals": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe operator approval decision evidence returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"status":             schema.StringAttribute{Computed: true, MarkdownDescription: "Current approval packet status."},
						"decision_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Operator decision id."},
						"source_packet_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
						"account_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"account_service_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id."},
						"service_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Service id."},
						"qr_workspace_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Related QR workspace id, when present."},
						"reviewer_user_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Reviewer user id, when a decision has been made."},
						"decided_at":         schema.StringAttribute{Computed: true, MarkdownDescription: "Decision timestamp, when a decision has been made."},
						"decision_state":     schema.StringAttribute{Computed: true, MarkdownDescription: "Decision state, for example `blocked`."},
						"required_approval":  schema.StringAttribute{Computed: true, MarkdownDescription: "Human approval required before advancing."},
						"approved_actions":   schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Actions approved by this decision, when any."},
						"blocked_actions":    schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Actions blocked by this decision."},
						"audit_event_action": schema.StringAttribute{Computed: true, MarkdownDescription: "Audit action associated with the decision evidence."},
						"result":             schema.StringAttribute{Computed: true, MarkdownDescription: "Decision result, for example `not_executed`."},
					},
				},
			},
		},
	}
}

func (d *OperatorApprovalsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OperatorApprovalsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data operatorApprovalsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading operator approvals.")
		return
	}

	operatorApprovals, err := d.client.ListOperatorApprovals(ctx, stringFilters(map[string]types.String{
		"account_id":         data.AccountID,
		"account_service_id": data.AccountServiceID,
		"service_id":         data.ServiceID,
		"decision_id":        data.DecisionID,
		"decision_state":     data.DecisionState,
		"required_approval":  data.RequiredApproval,
		"source_packet_id":   data.SourcePacketID,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Operator Approvals", err.Error())
		return
	}

	data.OperatorApprovals = make([]operatorApprovalStateModel, 0, len(operatorApprovals))
	for _, approval := range operatorApprovals {
		data.OperatorApprovals = append(data.OperatorApprovals, operatorApprovalStateModel{
			Status:           types.StringValue(approval.Status),
			DecisionID:       types.StringValue(approval.DecisionID),
			SourcePacketID:   stringValueOrNull(approval.SourcePacketID),
			AccountID:        types.StringValue(approval.AccountID),
			AccountServiceID: stringValueOrNull(approval.AccountServiceID),
			ServiceID:        stringValueOrNull(approval.ServiceID),
			QRWorkspaceID:    stringValueOrNull(approval.QRWorkspaceID),
			ReviewerUserID:   stringValueOrNull(approval.ReviewerUserID),
			DecidedAt:        stringValueOrNull(approval.DecidedAt),
			DecisionState:    types.StringValue(approval.DecisionState),
			RequiredApproval: stringValueOrNull(approval.RequiredApproval),
			ApprovedActions:  stringSliceValues(approval.ApprovedActions),
			BlockedActions:   stringSliceValues(approval.BlockedActions),
			AuditEventAction: stringValueOrNull(approval.AuditEventAction),
			Result:           stringValueOrNull(approval.Result),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
