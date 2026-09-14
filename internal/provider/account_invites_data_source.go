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
	_ datasource.DataSource              = &AccountInvitesDataSource{}
	_ datasource.DataSourceWithConfigure = &AccountInvitesDataSource{}
)

// AccountInvitesDataSource reads draft account invite evidence.
type AccountInvitesDataSource struct {
	client *linkridgecloud.Client
}

// NewAccountInvitesDataSource returns an account invites data source.
func NewAccountInvitesDataSource() datasource.DataSource {
	return &AccountInvitesDataSource{}
}

type accountInvitesDataSourceModel struct {
	AccountID            types.String              `tfsdk:"account_id"`
	Role                 types.String              `tfsdk:"role"`
	Status               types.String              `tfsdk:"status"`
	InviteDeliveryStatus types.String              `tfsdk:"invite_delivery_status"`
	SourcePacketID       types.String              `tfsdk:"source_packet_id"`
	AccountInvites       []accountInviteStateModel `tfsdk:"account_invites"`
}

type accountInviteStateModel struct {
	ID                              types.String   `tfsdk:"id"`
	AccountID                       types.String   `tfsdk:"account_id"`
	Email                           types.String   `tfsdk:"email"`
	Role                            types.String   `tfsdk:"role"`
	ServiceScope                    []types.String `tfsdk:"service_scope"`
	Status                          types.String   `tfsdk:"status"`
	ExpiresAt                       types.String   `tfsdk:"expires_at"`
	CreatedByUserID                 types.String   `tfsdk:"created_by_user_id"`
	CreatedAt                       types.String   `tfsdk:"created_at"`
	UpdatedAt                       types.String   `tfsdk:"updated_at"`
	InviteDeliveryStatus            types.String   `tfsdk:"invite_delivery_status"`
	ApprovalRequired                types.String   `tfsdk:"approval_required"`
	SourcePacketID                  types.String   `tfsdk:"source_packet_id"`
	ExternalEffectsEnabled          types.Bool     `tfsdk:"external_effects_enabled"`
	DeliveryStatus                  types.String   `tfsdk:"delivery_status"`
	DeliverySentAt                  types.String   `tfsdk:"delivery_sent_at"`
	DeliveryAcceptedAt              types.String   `tfsdk:"delivery_accepted_at"`
	DeliveryExternalEffectPerformed types.Bool     `tfsdk:"delivery_external_effect_performed"`
	DeliveryApprovalRequired        types.String   `tfsdk:"delivery_approval_required"`
	DeliveryBlockedReason           types.String   `tfsdk:"delivery_blocked_reason"`
}

// Metadata sets the data source type name.
func (d *AccountInvitesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_invites"
}

// Schema defines the data source schema.
func (d *AccountInvitesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads draft LinkRidge Cloud account invite evidence from `/v1/accounts/{account_id}/invites`. This data source does not send invites, create external identity access, notify customers, or approve invite delivery.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose draft invites should be read.",
			},
			"role": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional invite role filter, for example `editor`.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional invite status filter, for example `draft`.",
			},
			"invite_delivery_status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional invite delivery status filter, for example `not_sent`.",
			},
			"source_packet_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional seed packet or draft source id filter.",
			},
			"account_invites": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Draft account invite evidence returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                                 schema.StringAttribute{Computed: true, MarkdownDescription: "Account invite id."},
						"account_id":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"email":                              schema.StringAttribute{Computed: true, MarkdownDescription: "Invited email address."},
						"role":                               schema.StringAttribute{Computed: true, MarkdownDescription: "Requested account role."},
						"service_scope":                      schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Service ids scoped to the invite."},
						"status":                             schema.StringAttribute{Computed: true, MarkdownDescription: "Current invite status."},
						"expires_at":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Invite expiry timestamp, when planned."},
						"created_by_user_id":                 schema.StringAttribute{Computed: true, MarkdownDescription: "User id that prepared the draft invite."},
						"created_at":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Draft creation timestamp, when available."},
						"updated_at":                         schema.StringAttribute{Computed: true, MarkdownDescription: "Draft update timestamp, when available."},
						"invite_delivery_status":             schema.StringAttribute{Computed: true, MarkdownDescription: "Invite delivery status."},
						"approval_required":                  schema.StringAttribute{Computed: true, MarkdownDescription: "Operator approval required before invite delivery."},
						"source_packet_id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
						"external_effects_enabled":           schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether external invite or identity effects are enabled."},
						"delivery_status":                    schema.StringAttribute{Computed: true, MarkdownDescription: "Delivery sub-state status, when present."},
						"delivery_sent_at":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Timestamp when an invite was sent, when present."},
						"delivery_accepted_at":               schema.StringAttribute{Computed: true, MarkdownDescription: "Timestamp when an invite was accepted, when present."},
						"delivery_external_effect_performed": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether invite delivery caused an external effect."},
						"delivery_approval_required":         schema.StringAttribute{Computed: true, MarkdownDescription: "Approval required by the delivery sub-state."},
						"delivery_blocked_reason":            schema.StringAttribute{Computed: true, MarkdownDescription: "Reason invite delivery remains blocked."},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *AccountInvitesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes draft account invite evidence.
func (d *AccountInvitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accountInvitesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading account invites.")
		return
	}

	invites, err := d.client.ListAccountInvites(ctx, data.AccountID.ValueString(), stringFilters(map[string]types.String{
		"role":                   data.Role,
		"status":                 data.Status,
		"invite_delivery_status": data.InviteDeliveryStatus,
		"source_packet_id":       data.SourcePacketID,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Account Invites", err.Error())
		return
	}

	data.AccountInvites = make([]accountInviteStateModel, 0, len(invites))
	for _, invite := range invites {
		data.AccountInvites = append(data.AccountInvites, accountInviteStateModel{
			ID:                              types.StringValue(invite.ID),
			AccountID:                       types.StringValue(invite.AccountID),
			Email:                           stringValueOrNull(invite.Email),
			Role:                            stringValueOrNull(invite.Role),
			ServiceScope:                    stringSliceValues(invite.ServiceScope),
			Status:                          types.StringValue(invite.Status),
			ExpiresAt:                       stringValueOrNull(invite.ExpiresAt),
			CreatedByUserID:                 stringValueOrNull(invite.CreatedByUserID),
			CreatedAt:                       stringValueOrNull(invite.CreatedAt),
			UpdatedAt:                       stringValueOrNull(invite.UpdatedAt),
			InviteDeliveryStatus:            stringValueOrNull(invite.InviteDeliveryStatus),
			ApprovalRequired:                stringValueOrNull(firstNonEmpty(invite.ApprovalRequired, invite.Delivery.ApprovalRequired)),
			SourcePacketID:                  stringValueOrNull(invite.SourcePacketID),
			ExternalEffectsEnabled:          types.BoolValue(invite.ExternalEffectsEnabled),
			DeliveryStatus:                  stringValueOrNull(invite.Delivery.Status),
			DeliverySentAt:                  stringValueOrNull(invite.Delivery.SentAt),
			DeliveryAcceptedAt:              stringValueOrNull(invite.Delivery.AcceptedAt),
			DeliveryExternalEffectPerformed: types.BoolValue(invite.Delivery.ExternalEffectPerformed),
			DeliveryApprovalRequired:        stringValueOrNull(invite.Delivery.ApprovalRequired),
			DeliveryBlockedReason:           stringValueOrNull(invite.Delivery.BlockedReason),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
