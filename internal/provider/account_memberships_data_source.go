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
	_ datasource.DataSource              = &AccountMembershipsDataSource{}
	_ datasource.DataSourceWithConfigure = &AccountMembershipsDataSource{}
)

// AccountMembershipsDataSource reads account-scoped role assignment evidence.
type AccountMembershipsDataSource struct {
	client *linkridgecloud.Client
}

// NewAccountMembershipsDataSource returns an account memberships data source.
func NewAccountMembershipsDataSource() datasource.DataSource {
	return &AccountMembershipsDataSource{}
}

type accountMembershipsDataSourceModel struct {
	AccountID          types.String                  `tfsdk:"account_id"`
	UserID             types.String                  `tfsdk:"user_id"`
	Role               types.String                  `tfsdk:"role"`
	Status             types.String                  `tfsdk:"status"`
	SourcePacketID     types.String                  `tfsdk:"source_packet_id"`
	AccountMemberships []accountMembershipStateModel `tfsdk:"account_memberships"`
}

type accountMembershipStateModel struct {
	ID                   types.String `tfsdk:"id"`
	AccountID            types.String `tfsdk:"account_id"`
	UserID               types.String `tfsdk:"user_id"`
	Role                 types.String `tfsdk:"role"`
	Status               types.String `tfsdk:"status"`
	InviteDeliveryStatus types.String `tfsdk:"invite_delivery_status"`
	SourcePacketID       types.String `tfsdk:"source_packet_id"`
}

// Metadata sets the data source type name.
func (d *AccountMembershipsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_memberships"
}

// Schema defines the data source schema.
func (d *AccountMembershipsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads account-scoped LinkRidge Cloud role assignment evidence from `/v1/accounts/{account_id}/memberships`. This data source does not create users, send invites, grant external access, or change roles.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Account id whose memberships should be read.",
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional user id filter.",
			},
			"role": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional account role filter, for example `owner`.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional membership status filter, for example `planned`.",
			},
			"source_packet_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional seed packet or draft source id filter.",
			},
			"account_memberships": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Account membership evidence returned by LinkRidge Cloud.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                     schema.StringAttribute{Computed: true, MarkdownDescription: "Account membership id."},
						"account_id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Owning account id."},
						"user_id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Member user id."},
						"role":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Account role."},
						"status":                 schema.StringAttribute{Computed: true, MarkdownDescription: "Current membership status."},
						"invite_delivery_status": schema.StringAttribute{Computed: true, MarkdownDescription: "Invite delivery state associated with the membership, when present."},
						"source_packet_id":       schema.StringAttribute{Computed: true, MarkdownDescription: "Seed packet or draft source id."},
					},
				},
			},
		},
	}
}

// Configure stores the configured API client.
func (d *AccountMembershipsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes account membership evidence.
func (d *AccountMembershipsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data accountMembershipsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Missing LinkRidge Cloud Client", "The provider was not configured before reading account memberships.")
		return
	}

	memberships, err := d.client.ListAccountMemberships(ctx, data.AccountID.ValueString(), stringFilters(map[string]types.String{
		"user_id":          data.UserID,
		"role":             data.Role,
		"status":           data.Status,
		"source_packet_id": data.SourcePacketID,
	}))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read LinkRidge Cloud Account Memberships", err.Error())
		return
	}

	data.AccountMemberships = make([]accountMembershipStateModel, 0, len(memberships))
	for _, membership := range memberships {
		data.AccountMemberships = append(data.AccountMemberships, accountMembershipStateModel{
			ID:                   types.StringValue(membership.ID),
			AccountID:            types.StringValue(membership.AccountID),
			UserID:               types.StringValue(membership.UserID),
			Role:                 types.StringValue(membership.Role),
			Status:               types.StringValue(membership.Status),
			InviteDeliveryStatus: stringValueOrNull(membership.InviteDeliveryStatus),
			SourcePacketID:       stringValueOrNull(membership.SourcePacketID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
