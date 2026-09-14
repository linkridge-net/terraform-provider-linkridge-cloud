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
	_ datasource.DataSource              = &QRMutationRehearsalRequirementsDataSource{}
	_ datasource.DataSourceWithConfigure = &QRMutationRehearsalRequirementsDataSource{}
)

// QRMutationRehearsalRequirementsDataSource reads a safe local/dev QR mutation rehearsal checklist.
type QRMutationRehearsalRequirementsDataSource struct {
	client *linkridgecloud.Client
}

// NewQRMutationRehearsalRequirementsDataSource returns a QR mutation rehearsal requirements data source.
func NewQRMutationRehearsalRequirementsDataSource() datasource.DataSource {
	return &QRMutationRehearsalRequirementsDataSource{}
}

type qrMutationRehearsalRequirementsDataSourceModel struct {
	WorkspaceID                      types.String   `tfsdk:"workspace_id"`
	MutationRequestID                types.String   `tfsdk:"mutation_request_id"`
	AccountID                        types.String   `tfsdk:"account_id"`
	ServiceID                        types.String   `tfsdk:"service_id"`
	RehearsalAllowed                 types.Bool     `tfsdk:"rehearsal_allowed"`
	RehearsalBlockedReason           types.String   `tfsdk:"rehearsal_blocked_reason"`
	RehearsalRequirementsFingerprint types.String   `tfsdk:"rehearsal_requirements_fingerprint"`
	ActorBindingFingerprint          types.String   `tfsdk:"actor_binding_fingerprint"`
	RequiredBodyFields               []types.String `tfsdk:"required_body_fields"`
	BlockedExternalActions           []types.String `tfsdk:"blocked_external_actions"`
	RehearsalTemplateJSON            types.String   `tfsdk:"rehearsal_template_json"`
	CurrentRehearsalJSON             types.String   `tfsdk:"current_rehearsal_json"`
	GuardrailsJSON                   types.String   `tfsdk:"guardrails_json"`
}

func (d *QRMutationRehearsalRequirementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_qr_mutation_rehearsal_requirements"
}

func (d *QRMutationRehearsalRequirementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the internal/dev QR mutation execution rehearsal checklist from `/v1/qr/workspaces/{workspace_id}/mutation-requests/{mutation_request_id}/execution-rehearsal-requirements`. This data source is read-only: it does not approve work, execute QR mutations, write tenant QR records, enable hosted redirects, record billing usage, or create customer-visible effects.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "QR workspace id that owns the mutation request.",
			},
			"mutation_request_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "QR mutation request id whose rehearsal checklist should be inspected.",
			},
			"account_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Owning account id.",
			},
			"service_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service id, usually `qr-codes`.",
			},
			"rehearsal_allowed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the mutation request currently has the required approved_local_only review evidence for a local/dev rehearsal.",
			},
			"rehearsal_blocked_reason": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Reason rehearsal is blocked, when blocked.",
			},
			"rehearsal_requirements_fingerprint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fingerprint binding the current requirements response to the mutation request, approval state, skipped actions, and actor binding.",
			},
			"actor_binding_fingerprint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Non-secret actor binding fingerprint returned by the control plane.",
			},
			"required_body_fields": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Fields required by the local/dev execution-rehearsal request, empty while rehearsal is blocked.",
			},
			"blocked_external_actions": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "External or customer-visible actions that must remain acknowledged as skipped.",
			},
			"rehearsal_template_json": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Safe request template as compact JSON, when rehearsal is allowed.",
			},
			"current_rehearsal_json": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Existing local rehearsal evidence as compact JSON, when present.",
			},
			"guardrails_json": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Local-only guardrail evidence as compact JSON.",
			},
		},
	}
}

func (d *QRMutationRehearsalRequirementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *QRMutationRehearsalRequirementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data qrMutationRehearsalRequirementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading QR mutation rehearsal requirements.")
		return
	}

	requirements, err := d.client.GetQRMutationRehearsalRequirements(ctx, data.WorkspaceID.ValueString(), data.MutationRequestID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud QR Mutation Rehearsal Requirements", err.Error())
		return
	}

	data.AccountID = types.StringValue(requirements.AccountID)
	data.ServiceID = stringValueOrNull(requirements.ServiceID)
	data.RehearsalAllowed = types.BoolValue(requirements.RehearsalAllowed)
	data.RehearsalBlockedReason = stringValueOrNull(requirements.RehearsalBlockedReason)
	data.RehearsalRequirementsFingerprint = stringValueOrNull(requirements.RehearsalRequirementsFingerprint)
	data.ActorBindingFingerprint = stringValueOrNull(requirements.ActorBindingFingerprint)
	data.RequiredBodyFields = stringSliceValues(requirements.RequiredBodyFields)
	data.BlockedExternalActions = stringSliceValues(requirements.BlockedExternalActions)
	data.RehearsalTemplateJSON = rawJSONValueOrNull(requirements.RehearsalTemplate)
	data.CurrentRehearsalJSON = rawJSONValueOrNull(requirements.CurrentRehearsal)
	data.GuardrailsJSON = rawJSONValueOrNull(requirements.Guardrails)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
