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
	_ datasource.DataSource              = &QRImportJobsDataSource{}
	_ datasource.DataSourceWithConfigure = &QRImportJobsDataSource{}
)

// QRImportJobsDataSource reads safe QR import planning metadata for a workspace.
type QRImportJobsDataSource struct {
	client *linkridgecloud.Client
}

// NewQRImportJobsDataSource returns a QR import jobs data source.
func NewQRImportJobsDataSource() datasource.DataSource {
	return &QRImportJobsDataSource{}
}

type qrImportJobsDataSourceModel struct {
	WorkspaceID     types.String            `tfsdk:"workspace_id"`
	ID              types.String            `tfsdk:"id"`
	Status          types.String            `tfsdk:"status"`
	ImportPerformed types.Bool              `tfsdk:"import_performed"`
	ImportJobs      []qrImportJobStateModel `tfsdk:"import_jobs"`
}

type qrImportJobStateModel struct {
	ID                            types.String `tfsdk:"id"`
	AccountID                     types.String `tfsdk:"account_id"`
	AccountServiceID              types.String `tfsdk:"account_service_id"`
	WorkspaceID                   types.String `tfsdk:"workspace_id"`
	ServiceID                     types.String `tfsdk:"service_id"`
	Source                        types.String `tfsdk:"source"`
	SourceRepository              types.String `tfsdk:"source_repository"`
	RequestedByUserID             types.String `tfsdk:"requested_by_user_id"`
	Status                        types.String `tfsdk:"status"`
	ImportPerformed               types.Bool   `tfsdk:"import_performed"`
	RecordsExamined               types.Int64  `tfsdk:"records_examined"`
	RecordsPlanned                types.Int64  `tfsdk:"records_planned"`
	RecordsRejected               types.Int64  `tfsdk:"records_rejected"`
	ReviewStatus                  types.String `tfsdk:"review_status"`
	ReviewApprovalRequired        types.String `tfsdk:"review_approval_required"`
	ReviewArtifactReviewed        types.Bool   `tfsdk:"review_artifact_reviewed"`
	ReviewParserContractChecked   types.Bool   `tfsdk:"review_parser_contract_checked"`
	ReviewExternalEffectPerformed types.Bool   `tfsdk:"review_external_effect_performed"`
	SourcePacketID                types.String `tfsdk:"source_packet_id"`
	ExternalRedirectsEnabled      types.Bool   `tfsdk:"external_redirects_enabled"`
}

// Metadata sets the data source type name.
func (d *QRImportJobsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qr_import_jobs"
}

// Schema defines the data source schema.
func (d *QRImportJobsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads safe LinkRidge Cloud QR import planning metadata from the `/v1/qr/workspaces/{workspace_id}/import-jobs` control-plane API. Import execution and hosted redirects are not triggered by this data source.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "QR workspace id whose import plans should be read.",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional QR import job id filter.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional QR import job status filter, for example `planned`.",
			},
			"import_performed": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Optional filter for whether import execution has happened. Dev planning records should remain `false`.",
			},
			"import_jobs": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe QR import planning metadata returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "QR import job id.",
						},
						"account_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning account id.",
						},
						"account_service_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning account service id.",
						},
						"workspace_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Owning QR workspace id.",
						},
						"service_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Service id.",
						},
						"source": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Import source type.",
						},
						"source_repository": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Repository associated with the import source.",
						},
						"requested_by_user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User id that requested the import plan, when present.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Current import job status.",
						},
						"import_performed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether import execution has happened. Planning records should remain false until explicitly approved outside Terraform.",
						},
						"records_examined": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of source records examined by the import plan.",
						},
						"records_planned": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of records planned for future import.",
						},
						"records_rejected": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of source records rejected by validation.",
						},
						"review_status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Operator review status for the import plan.",
						},
						"review_approval_required": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Who must approve the import before external effects are allowed.",
						},
						"review_artifact_reviewed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the import artifact has been reviewed.",
						},
						"review_parser_contract_checked": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the service parser contract has been checked.",
						},
						"review_external_effect_performed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether review caused any external effect.",
						},
						"source_packet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Seed packet or draft source id.",
						},
						"external_redirects_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether hosted QR redirects are enabled for this import plan.",
						},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *QRImportJobsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the QR import planning state.
func (d *QRImportJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data qrImportJobsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading QR import jobs.")
		return
	}

	filters := stringFilters(map[string]types.String{
		"id":     data.ID,
		"status": data.Status,
	})
	if !data.ImportPerformed.IsNull() && !data.ImportPerformed.IsUnknown() {
		filters["import_performed"] = fmt.Sprintf("%t", data.ImportPerformed.ValueBool())
	}

	importJobs, err := d.client.ListQRImportJobs(ctx, data.WorkspaceID.ValueString(), filters)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud QR Import Jobs", err.Error())
		return
	}

	data.ImportJobs = make([]qrImportJobStateModel, 0, len(importJobs))
	for _, importJob := range importJobs {
		data.ImportJobs = append(data.ImportJobs, qrImportJobStateModel{
			ID:                            types.StringValue(importJob.ID),
			AccountID:                     types.StringValue(importJob.AccountID),
			AccountServiceID:              types.StringValue(importJob.AccountServiceID),
			WorkspaceID:                   types.StringValue(importJob.QRWorkspaceID),
			ServiceID:                     types.StringValue(importJob.ServiceID),
			Source:                        types.StringValue(importJob.Source),
			SourceRepository:              stringValueOrNull(importJob.SourceRepository),
			RequestedByUserID:             stringValueOrNull(importJob.RequestedByUserID),
			Status:                        types.StringValue(importJob.Status),
			ImportPerformed:               types.BoolValue(importJob.ImportPerformed),
			RecordsExamined:               types.Int64Value(importJob.RecordsExamined),
			RecordsPlanned:                types.Int64Value(importJob.RecordsPlanned),
			RecordsRejected:               types.Int64Value(importJob.RecordsRejected),
			ReviewStatus:                  stringValueOrNull(importJob.ImportReviewPacket.Status),
			ReviewApprovalRequired:        stringValueOrNull(importJob.ImportReviewPacket.ApprovalRequired),
			ReviewArtifactReviewed:        types.BoolValue(importJob.ImportReviewPacket.ArtifactReviewed),
			ReviewParserContractChecked:   types.BoolValue(importJob.ImportReviewPacket.ParserContractChecked),
			ReviewExternalEffectPerformed: types.BoolValue(importJob.ImportReviewPacket.ExternalEffectPerformed),
			SourcePacketID:                stringValueOrNull(importJob.SourcePacketID),
			ExternalRedirectsEnabled:      types.BoolValue(importJob.ExternalRedirectsEnabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
