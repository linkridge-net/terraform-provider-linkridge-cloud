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
	_ datasource.DataSource              = &ActivationPacketsDataSource{}
	_ datasource.DataSourceWithConfigure = &ActivationPacketsDataSource{}
)

// ActivationPacketsDataSource reads safe activation review metadata for an account service.
type ActivationPacketsDataSource struct {
	client *linkridgecloud.Client
}

// NewActivationPacketsDataSource returns an activation packets data source.
func NewActivationPacketsDataSource() datasource.DataSource {
	return &ActivationPacketsDataSource{}
}

type activationPacketsDataSourceModel struct {
	AccountID         types.String                 `tfsdk:"account_id"`
	AccountServiceID  types.String                 `tfsdk:"account_service_id"`
	ID                types.String                 `tfsdk:"id"`
	Status            types.String                 `tfsdk:"status"`
	ActivationPackets []activationPacketStateModel `tfsdk:"activation_packets"`
}

type activationPacketStateModel struct {
	ID                                types.String   `tfsdk:"id"`
	SourcePacketID                    types.String   `tfsdk:"source_packet_id"`
	Status                            types.String   `tfsdk:"status"`
	AccountID                         types.String   `tfsdk:"account_id"`
	AccountName                       types.String   `tfsdk:"account_name"`
	AccountStatus                     types.String   `tfsdk:"account_status"`
	PrimaryOwnerUserID                types.String   `tfsdk:"primary_owner_user_id"`
	AccountServiceID                  types.String   `tfsdk:"account_service_id"`
	ServiceID                         types.String   `tfsdk:"service_id"`
	PlanKey                           types.String   `tfsdk:"plan_key"`
	AccountServiceStatus              types.String   `tfsdk:"account_service_status"`
	ServiceWorkspaceID                types.String   `tfsdk:"service_workspace_id"`
	ServiceWorkspaceStatus            types.String   `tfsdk:"service_workspace_status"`
	ActivationPerformed               types.Bool     `tfsdk:"activation_performed"`
	ApprovalRequired                  types.String   `tfsdk:"approval_required"`
	ExternalEffectsEnabled            types.Bool     `tfsdk:"external_effects_enabled"`
	OperatorDecisionID                types.String   `tfsdk:"operator_decision_id"`
	OperatorDecisionState             types.String   `tfsdk:"operator_decision_state"`
	OperatorDecisionRequiredApproval  types.String   `tfsdk:"operator_decision_required_approval"`
	ApprovedActions                   []types.String `tfsdk:"approved_actions"`
	BlockedActions                    []types.String `tfsdk:"blocked_actions"`
	OperatorDecisionResult            types.String   `tfsdk:"operator_decision_result"`
	ProvisioningRunID                 types.String   `tfsdk:"provisioning_run_id"`
	ProvisioningRunStatus             types.String   `tfsdk:"provisioning_run_status"`
	ProvisioningPlannedSteps          []types.String `tfsdk:"provisioning_planned_steps"`
	ProvisioningLocalRecordsPrepared  []types.String `tfsdk:"provisioning_local_records_prepared"`
	ProvisioningExternalWritesBlocked []types.String `tfsdk:"provisioning_external_writes_blocked"`
	ProvisioningNextOperatorAction    types.String   `tfsdk:"provisioning_next_operator_action"`
}

// Metadata sets the data source type name.
func (d *ActivationPacketsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activation_packets"
}

// Schema defines the data source schema.
func (d *ActivationPacketsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud activation review metadata from the `/v1/accounts/{account_id}/services/{account_service_id}/activation-packets` control-plane API. This data source does not approve activation, create billing/customer access, issue secrets, send invites, or run external provisioning.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose activation packets should be read.",
			},
			"account_service_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account service id whose activation packets should be read.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional activation packet id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional activation packet status filter, for example `blocked_pending_matthew_approval`.",
			},
			"activation_packets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe activation review metadata returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: activationPacketAttributes(),
				},
			},
		},
	}
}

func activationPacketAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                                   schema.StringAttribute{Computed: true, MarkdownDescription: "Activation packet id."},
		"source_packet_id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
		"status":                               schema.StringAttribute{Computed: true, MarkdownDescription: "Current activation packet status."},
		"account_id":                           schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
		"account_name":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account display name."},
		"account_status":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account status."},
		"primary_owner_user_id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Primary owner user id."},
		"account_service_id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id."},
		"service_id":                           schema.StringAttribute{Computed: true, MarkdownDescription: "Service id."},
		"plan_key":                             schema.StringAttribute{Computed: true, MarkdownDescription: "Requested service plan key."},
		"account_service_status":               schema.StringAttribute{Computed: true, MarkdownDescription: "Account service planning status."},
		"service_workspace_id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Prepared service workspace id, when present."},
		"service_workspace_status":             schema.StringAttribute{Computed: true, MarkdownDescription: "Prepared service workspace status, when present."},
		"activation_performed":                 schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether activation was performed. Dev review packets should remain false until explicitly approved outside Terraform."},
		"approval_required":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Operator approval required before activation."},
		"external_effects_enabled":             schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether external customer effects are enabled for this activation packet."},
		"operator_decision_id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Operator decision id, when present."},
		"operator_decision_state":              schema.StringAttribute{Computed: true, MarkdownDescription: "Operator decision state, for example `blocked`."},
		"operator_decision_required_approval":  schema.StringAttribute{Computed: true, MarkdownDescription: "Approval required by the operator decision."},
		"approved_actions":                     schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Actions approved by the operator decision."},
		"blocked_actions":                      schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "External actions blocked by the operator decision."},
		"operator_decision_result":             schema.StringAttribute{Computed: true, MarkdownDescription: "Operator decision result, for example `not_executed`."},
		"provisioning_run_id":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Local provisioning rehearsal id, when present."},
		"provisioning_run_status":              schema.StringAttribute{Computed: true, MarkdownDescription: "Local provisioning rehearsal status."},
		"provisioning_planned_steps":           schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Local provisioning steps planned by the control plane."},
		"provisioning_local_records_prepared":  schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Local records prepared by the provisioning rehearsal."},
		"provisioning_external_writes_blocked": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "External writes blocked by the provisioning rehearsal."},
		"provisioning_next_operator_action":    schema.StringAttribute{Computed: true, MarkdownDescription: "Next operator action required before activation can advance."},
	}
}

// Configure stores the configured API client.
func (d *ActivationPacketsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes activation review metadata.
func (d *ActivationPacketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data activationPacketsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading activation packets.")
		return
	}

	activationPackets, err := d.client.ListActivationPackets(ctx, data.AccountID.ValueString(), data.AccountServiceID.ValueString(), stringFilters(map[string]types.String{
		"id":     data.ID,
		"status": data.Status,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Activation Packets", err.Error())
		return
	}

	data.ActivationPackets = make([]activationPacketStateModel, 0, len(activationPackets))
	for _, activationPacket := range activationPackets {
		data.ActivationPackets = append(data.ActivationPackets, activationPacketStateModel{
			ID:                                types.StringValue(activationPacket.ID),
			SourcePacketID:                    stringValueOrNull(activationPacket.SourcePacketID),
			Status:                            types.StringValue(activationPacket.Status),
			AccountID:                         types.StringValue(activationPacket.Account.ID),
			AccountName:                       stringValueOrNull(activationPacket.Account.Name),
			AccountStatus:                     stringValueOrNull(activationPacket.Account.Status),
			PrimaryOwnerUserID:                stringValueOrNull(activationPacket.Account.PrimaryOwnerUserID),
			AccountServiceID:                  types.StringValue(activationPacket.AccountService.ID),
			ServiceID:                         stringValueOrNull(activationPacket.AccountService.ServiceID),
			PlanKey:                           stringValueOrNull(activationPacket.AccountService.PlanKey),
			AccountServiceStatus:              stringValueOrNull(activationPacket.AccountService.Status),
			ServiceWorkspaceID:                stringValueOrNull(activationPacket.ServiceWorkspace.ID),
			ServiceWorkspaceStatus:            stringValueOrNull(activationPacket.ServiceWorkspace.Status),
			ActivationPerformed:               types.BoolValue(activationPacket.ActivationPerformed),
			ApprovalRequired:                  stringValueOrNull(activationPacket.ApprovalRequired),
			ExternalEffectsEnabled:            types.BoolValue(activationPacket.ExternalEffectsEnabled),
			OperatorDecisionID:                stringValueOrNull(activationPacket.OperatorApprovalDecision.DecisionID),
			OperatorDecisionState:             stringValueOrNull(activationPacket.OperatorApprovalDecision.DecisionState),
			OperatorDecisionRequiredApproval:  stringValueOrNull(activationPacket.OperatorApprovalDecision.RequiredApproval),
			ApprovedActions:                   stringSliceValues(activationPacket.OperatorApprovalDecision.ApprovedActions),
			BlockedActions:                    stringSliceValues(activationPacket.OperatorApprovalDecision.BlockedActions),
			OperatorDecisionResult:            stringValueOrNull(activationPacket.OperatorApprovalDecision.Result),
			ProvisioningRunID:                 stringValueOrNull(activationPacket.LocalProvisioningRun.RunID),
			ProvisioningRunStatus:             stringValueOrNull(activationPacket.LocalProvisioningRun.Status),
			ProvisioningPlannedSteps:          stringSliceValues(activationPacket.LocalProvisioningRun.PlannedSteps),
			ProvisioningLocalRecordsPrepared:  stringSliceValues(activationPacket.LocalProvisioningRun.LocalRecordsPrepared),
			ProvisioningExternalWritesBlocked: stringSliceValues(activationPacket.LocalProvisioningRun.ExternalWritesBlocked),
			ProvisioningNextOperatorAction:    stringValueOrNull(activationPacket.LocalProvisioningRun.NextRequiredOperatorAction),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
