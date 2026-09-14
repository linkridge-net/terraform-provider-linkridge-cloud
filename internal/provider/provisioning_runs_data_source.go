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
	_ datasource.DataSource              = &ProvisioningRunsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProvisioningRunsDataSource{}
)

// ProvisioningRunsDataSource reads safe provisioning rehearsal evidence.
type ProvisioningRunsDataSource struct {
	client *linkridgecloud.Client
}

// NewProvisioningRunsDataSource returns a provisioning runs data source.
func NewProvisioningRunsDataSource() datasource.DataSource {
	return &ProvisioningRunsDataSource{}
}

type provisioningRunsDataSourceModel struct {
	AccountID        types.String                `tfsdk:"account_id"`
	AccountServiceID types.String                `tfsdk:"account_service_id"`
	ServiceID        types.String                `tfsdk:"service_id"`
	Status           types.String                `tfsdk:"status"`
	SourcePacketID   types.String                `tfsdk:"source_packet_id"`
	ProvisioningRuns []provisioningRunStateModel `tfsdk:"provisioning_runs"`
}

type provisioningRunStateModel struct {
	ID                          types.String   `tfsdk:"id"`
	RunID                       types.String   `tfsdk:"run_id"`
	AccountID                   types.String   `tfsdk:"account_id"`
	AccountServiceID            types.String   `tfsdk:"account_service_id"`
	ServiceID                   types.String   `tfsdk:"service_id"`
	QRWorkspaceID               types.String   `tfsdk:"qr_workspace_id"`
	RequestedByUserID           types.String   `tfsdk:"requested_by_user_id"`
	ApprovedByUserID            types.String   `tfsdk:"approved_by_user_id"`
	Status                      types.String   `tfsdk:"status"`
	StartedAt                   types.String   `tfsdk:"started_at"`
	CompletedAt                 types.String   `tfsdk:"completed_at"`
	PlannedSteps                []types.String `tfsdk:"planned_steps"`
	LocalRecordsPrepared        []types.String `tfsdk:"local_records_prepared"`
	ExternalWritesBlocked       []types.String `tfsdk:"external_writes_blocked"`
	NextRequiredOperatorAction  types.String   `tfsdk:"next_required_operator_action"`
	Result                      types.String   `tfsdk:"result"`
	ResultExternalWritesBlocked []types.String `tfsdk:"result_external_writes_blocked"`
	SourcePacketID              types.String   `tfsdk:"source_packet_id"`
	ExternalEffectsEnabled      types.Bool     `tfsdk:"external_effects_enabled"`
}

// Metadata sets the data source type name.
func (d *ProvisioningRunsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provisioning_runs"
}

// Schema defines the data source schema.
func (d *ProvisioningRunsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud provisioning rehearsal evidence from the `/v1/provisioning-runs` control-plane API. This data source does not approve provisioning, run external writes, or enable customer-visible service access.",
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
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional provisioning run status filter, for example `rehearsal_only` or `blocked`.",
			},
			"source_packet_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional seed packet or draft source id filter.",
			},
			"provisioning_runs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe provisioning rehearsal metadata returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Provisioning run id.",
						},
						"run_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Legacy or seed-packet provisioning run id, when present.",
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
						"qr_workspace_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Related QR workspace id, when present.",
						},
						"requested_by_user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User id that requested the provisioning rehearsal, when present.",
						},
						"approved_by_user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User id that approved provisioning, when present.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current provisioning run status.",
						},
						"started_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Start timestamp, when execution has begun.",
						},
						"completed_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Completion timestamp, when execution has completed.",
						},
						"planned_steps": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Local planning steps persisted by the control plane.",
						},
						"local_records_prepared": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Local records prepared by the provisioning rehearsal.",
						},
						"external_writes_blocked": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "External writes explicitly blocked by the rehearsal.",
						},
						"next_required_operator_action": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Next operator action required before provisioning can advance.",
						},
						"result": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Provisioning result, for example `not_executed`.",
						},
						"result_external_writes_blocked": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "External writes blocked inside the persisted result payload.",
						},
						"source_packet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Seed packet or draft source id.",
						},
						"external_effects_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether external customer effects are enabled for this provisioning run.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *ProvisioningRunsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the provisioning rehearsal state.
func (d *ProvisioningRunsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data provisioningRunsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading provisioning runs.")
		return
	}

	provisioningRuns, err := d.client.ListProvisioningRuns(ctx, stringFilters(map[string]types.String{
		"account_id":         data.AccountID,
		"account_service_id": data.AccountServiceID,
		"service_id":         data.ServiceID,
		"status":             data.Status,
		"source_packet_id":   data.SourcePacketID,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Provisioning Runs", err.Error())
		return
	}

	data.ProvisioningRuns = make([]provisioningRunStateModel, 0, len(provisioningRuns))
	for _, provisioningRun := range provisioningRuns {
		result := provisioningRun.Result.Result
		if result == "" {
			result = "not_executed"
		}

		data.ProvisioningRuns = append(data.ProvisioningRuns, provisioningRunStateModel{
			ID:                          stringValueOrNull(firstNonEmpty(provisioningRun.ID, provisioningRun.RunID)),
			RunID:                       stringValueOrNull(provisioningRun.RunID),
			AccountID:                   types.StringValue(provisioningRun.AccountID),
			AccountServiceID:            stringValueOrNull(provisioningRun.AccountServiceID),
			ServiceID:                   stringValueOrNull(provisioningRun.ServiceID),
			QRWorkspaceID:               stringValueOrNull(provisioningRun.QRWorkspaceID),
			RequestedByUserID:           stringValueOrNull(provisioningRun.RequestedByUserID),
			ApprovedByUserID:            stringValueOrNull(provisioningRun.ApprovedByUserID),
			Status:                      types.StringValue(provisioningRun.Status),
			StartedAt:                   stringValueOrNull(provisioningRun.StartedAt),
			CompletedAt:                 stringValueOrNull(provisioningRun.CompletedAt),
			PlannedSteps:                stringSliceValues(provisioningRun.PlannedSteps),
			LocalRecordsPrepared:        stringSliceValues(provisioningRun.LocalRecordsPrepared),
			ExternalWritesBlocked:       stringSliceValues(provisioningRun.ExternalWritesBlocked),
			NextRequiredOperatorAction:  stringValueOrNull(provisioningRun.NextRequiredOperatorAction),
			Result:                      stringValueOrNull(result),
			ResultExternalWritesBlocked: stringSliceValues(provisioningRun.Result.ExternalWritesBlocked),
			SourcePacketID:              stringValueOrNull(provisioningRun.SourcePacketID),
			ExternalEffectsEnabled:      types.BoolValue(provisioningRun.ExternalEffectsEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
