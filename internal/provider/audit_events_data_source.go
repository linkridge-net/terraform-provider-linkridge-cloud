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
	_ datasource.DataSource              = &AuditEventsDataSource{}
	_ datasource.DataSourceWithConfigure = &AuditEventsDataSource{}
)

// AuditEventsDataSource reads append-only control-plane audit evidence.
type AuditEventsDataSource struct {
	client *linkridgecloud.Client
}

// NewAuditEventsDataSource returns an audit events data source.
func NewAuditEventsDataSource() datasource.DataSource {
	return &AuditEventsDataSource{}
}

type auditEventsDataSourceModel struct {
	ID               types.String           `tfsdk:"id"`
	AccountID        types.String           `tfsdk:"account_id"`
	AccountServiceID types.String           `tfsdk:"account_service_id"`
	ServiceID        types.String           `tfsdk:"service_id"`
	Action           types.String           `tfsdk:"action"`
	TargetType       types.String           `tfsdk:"target_type"`
	TargetID         types.String           `tfsdk:"target_id"`
	SourcePacketID   types.String           `tfsdk:"source_packet_id"`
	AuditEvents      []auditEventStateModel `tfsdk:"audit_events"`
}

type auditEventStateModel struct {
	ID               types.String `tfsdk:"id"`
	AccountID        types.String `tfsdk:"account_id"`
	AccountServiceID types.String `tfsdk:"account_service_id"`
	ServiceID        types.String `tfsdk:"service_id"`
	QRWorkspaceID    types.String `tfsdk:"qr_workspace_id"`
	QRCodeID         types.String `tfsdk:"qr_code_id"`
	Action           types.String `tfsdk:"action"`
	TargetType       types.String `tfsdk:"target_type"`
	TargetID         types.String `tfsdk:"target_id"`
	ActorUserID      types.String `tfsdk:"actor_user_id"`
	OccurredAt       types.String `tfsdk:"occurred_at"`
	MetadataJSON     types.String `tfsdk:"metadata_json"`
	SourcePacketID   types.String `tfsdk:"source_packet_id"`
}

func (d *AuditEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_events"
}

func (d *AuditEventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads append-only LinkRidge Cloud audit events from `/v1/audit-events`. This data source exposes persisted evidence only; it does not replay events, approve work, issue secrets, execute imports, export billing usage, notify customers, or enable customer-visible effects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional audit event id filter.",
			},
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
			"action": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional audit action filter, for example `service_token.prepare` or `qr_import_job.prepare`.",
			},
			"target_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional target type filter.",
			},
			"target_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional target id filter.",
			},
			"source_packet_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional seed packet or draft source id filter.",
			},
			"audit_events": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Safe append-only audit evidence returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Audit event id, when present."},
						"account_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"account_service_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account service id, when present."},
						"service_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Service id, when present."},
						"qr_workspace_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Related QR workspace id, when present."},
						"qr_code_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Related QR code id, when present."},
						"action":             schema.StringAttribute{Computed: true, MarkdownDescription: "Audit action."},
						"target_type":        schema.StringAttribute{Computed: true, MarkdownDescription: "Audited target type."},
						"target_id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Audited target id."},
						"actor_user_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "User id associated with the action, when present."},
						"occurred_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Timestamp for the event, when present."},
						"metadata_json":      schema.StringAttribute{Computed: true, MarkdownDescription: "Event metadata as compact JSON for operator inspection."},
						"source_packet_id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
					},
				},
			},
		},
	}
}

func (d *AuditEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AuditEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data auditEventsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading audit events.")
		return
	}

	auditEvents, err := d.client.ListAuditEvents(ctx, stringFilters(map[string]types.String{
		"id":                 data.ID,
		"account_id":         data.AccountID,
		"account_service_id": data.AccountServiceID,
		"service_id":         data.ServiceID,
		"action":             data.Action,
		"target_type":        data.TargetType,
		"target_id":          data.TargetID,
		"source_packet_id":   data.SourcePacketID,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Audit Events", err.Error())
		return
	}

	data.AuditEvents = make([]auditEventStateModel, 0, len(auditEvents))
	for _, event := range auditEvents {
		data.AuditEvents = append(data.AuditEvents, auditEventStateModel{
			ID:               stringValueOrNull(event.ID),
			AccountID:        stringValueOrNull(event.AccountID),
			AccountServiceID: stringValueOrNull(event.AccountServiceID),
			ServiceID:        stringValueOrNull(event.ServiceID),
			QRWorkspaceID:    stringValueOrNull(event.QRWorkspaceID),
			QRCodeID:         stringValueOrNull(event.QRCodeID),
			Action:           types.StringValue(event.Action),
			TargetType:       stringValueOrNull(event.TargetType),
			TargetID:         stringValueOrNull(event.TargetID),
			ActorUserID:      stringValueOrNull(event.ActorUserID),
			OccurredAt:       stringValueOrNull(event.OccurredAt),
			MetadataJSON:     rawJSONValueOrNull(event.Metadata),
			SourcePacketID:   stringValueOrNull(event.SourcePacketID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
