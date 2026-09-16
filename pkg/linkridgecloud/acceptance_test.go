// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

package linkridgecloud

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestAccClientReadOnlyDevSurface(t *testing.T) {
	if os.Getenv("LINKRIDGE_CLOUD_ACC") != "1" {
		t.Skip("set LINKRIDGE_CLOUD_ACC=1 to run read-only dev API acceptance checks")
	}

	client, err := NewClient(ClientConfig{
		BaseURL:  os.Getenv("LINKRIDGE_CLOUD_BASE_URL"),
		APIToken: os.Getenv("LINKRIDGE_CLOUD_API_TOKEN"),
	})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	services, err := client.ListServices(ctx, "qr-codes")
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	if len(services) != 1 || services[0].ID != "qr-codes" {
		t.Fatalf("expected qr-codes service, got %#v", services)
	}

	accounts, err := client.ListAccounts(ctx, "acct_local_qr_demo", "draft")
	if err != nil {
		t.Fatalf("list accounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected one draft account, got %#v", accounts)
	}
	if accounts[0].ExternalEffectsEnabled {
		t.Fatal("expected draft account external effects to stay disabled")
	}

	memberships, err := client.ListAccountMemberships(ctx, "acct_local_qr_demo", map[string]string{
		"user_id": "usr_local_qr_owner",
		"role":    "owner",
		"status":  "planned",
	})
	if err != nil {
		t.Fatalf("list account memberships: %v", err)
	}
	if len(memberships) != 1 {
		t.Fatalf("expected one planned membership, got %#v", memberships)
	}
	if memberships[0].InviteDeliveryStatus != "not_sent" {
		t.Fatalf("expected membership invite delivery to stay not_sent, got %q", memberships[0].InviteDeliveryStatus)
	}

	invites, err := client.ListAccountInvites(ctx, "acct_local_qr_demo", map[string]string{
		"role":                   "editor",
		"status":                 "draft",
		"invite_delivery_status": "not_sent",
	})
	if err != nil {
		t.Fatalf("list account invites: %v", err)
	}
	for _, invite := range invites {
		if invite.ExternalEffectsEnabled || invite.Delivery.ExternalEffectPerformed {
			t.Fatalf("expected invite %q external delivery to stay disabled", invite.ID)
		}
	}

	accountServices, err := client.ListAccountServices(ctx, "acct_local_qr_demo", map[string]string{
		"id":     "asvc_local_qr_demo",
		"status": "planned",
	})
	if err != nil {
		t.Fatalf("list account services: %v", err)
	}
	if len(accountServices) != 1 {
		t.Fatalf("expected one planned account service, got %#v", accountServices)
	}
	if accountServices[0].ExternalEffectsEnabled {
		t.Fatal("expected planned account service external effects to stay disabled")
	}

	entitlements, err := client.ListEntitlements(ctx, "acct_local_qr_demo", "asvc_local_qr_demo", map[string]string{
		"entitlement_key": "active_qr_codes",
	})
	if err != nil {
		t.Fatalf("list entitlements: %v", err)
	}
	if len(entitlements) != 1 || entitlements[0].EntitlementKey != "active_qr_codes" {
		t.Fatalf("expected active_qr_codes entitlement, got %#v", entitlements)
	}

	workspaces, err := client.ListQRWorkspaces(ctx, map[string]string{
		"id":         "qrw_local_demo",
		"account_id": "acct_local_qr_demo",
		"status":     "planned",
	})
	if err != nil {
		t.Fatalf("list QR workspaces: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("expected one planned QR workspace, got %#v", workspaces)
	}
	if workspaces[0].ExternalRedirectsEnabled {
		t.Fatal("expected QR workspace external redirects to stay disabled")
	}

	importJobs, err := client.ListQRImportJobs(ctx, "qrw_local_demo", map[string]string{
		"status":           "planned",
		"import_performed": "false",
	})
	if err != nil {
		t.Fatalf("list QR import jobs: %v", err)
	}
	for _, job := range importJobs {
		if job.ImportPerformed {
			t.Fatalf("expected import job %q to remain unperformed", job.ID)
		}
		if job.ExternalRedirectsEnabled || job.ImportReviewPacket.ExternalEffectPerformed {
			t.Fatalf("expected import job %q external effects to stay disabled", job.ID)
		}
	}

	qrMutationRequirements, err := client.GetQRMutationRehearsalRequirements(ctx, "qrw_local_growth_demo", "qrm_local_growth_demo_menu_archive")
	if err != nil {
		t.Fatalf("get QR mutation rehearsal requirements: %v", err)
	}
	if qrMutationRequirements.AccountID != "acct_local_qr_growth_demo" {
		t.Fatalf("unexpected QR mutation rehearsal account: %q", qrMutationRequirements.AccountID)
	}
	if qrMutationRequirements.RehearsalAllowed {
		t.Fatal("expected QR mutation rehearsal to stay blocked until local approval evidence exists")
	}
	if len(qrMutationRequirements.BlockedExternalActions) == 0 {
		t.Fatal("expected QR mutation rehearsal requirements to include blocked external actions")
	}

	serviceTokens, err := client.ListServiceTokens(ctx, "acct_local_qr_demo", map[string]string{
		"status":                 "planned",
		"secret_material_issued": "false",
	})
	if err != nil {
		t.Fatalf("list service tokens: %v", err)
	}
	for _, token := range serviceTokens {
		if token.SecretMaterialIssued {
			t.Fatalf("expected service token %q to have no secret material", token.ID)
		}
		if token.ExternalEffectsEnabled {
			t.Fatalf("expected service token %q external effects to stay disabled", token.ID)
		}
	}

	activationPackets, err := client.ListActivationPackets(ctx, "acct_local_qr_demo", "asvc_local_qr_demo", map[string]string{
		"status": "blocked_pending_matthew_approval",
	})
	if err != nil {
		t.Fatalf("list activation packets: %v", err)
	}
	for _, packet := range activationPackets {
		if packet.ActivationPerformed {
			t.Fatalf("expected activation packet %q to remain unperformed", packet.ID)
		}
		if packet.ExternalEffectsEnabled {
			t.Fatalf("expected activation packet %q external effects to stay disabled", packet.ID)
		}
	}

	provisioningRuns, err := client.ListProvisioningRuns(ctx, map[string]string{
		"account_id":         "acct_local_qr_demo",
		"account_service_id": "asvc_local_qr_demo",
	})
	if err != nil {
		t.Fatalf("list provisioning runs: %v", err)
	}
	for _, run := range provisioningRuns {
		if run.ExternalEffectsEnabled {
			t.Fatalf("expected provisioning run %q external effects to stay disabled", run.ID)
		}
	}

	billingRequests, err := client.ListBillingExportRequests(ctx, "acct_local_qr_demo", "asvc_local_qr_demo", map[string]string{
		"status": "blocked",
	})
	if err != nil {
		t.Fatalf("list billing export requests: %v", err)
	}
	for _, request := range billingRequests {
		if request.ExternalExportEnabled {
			t.Fatalf("expected billing export request %q external export to stay disabled", request.ID)
		}
	}

	supportCases, err := client.ListSupportCases(ctx, "acct_local_qr_demo", "asvc_local_qr_demo", map[string]string{
		"status": "blocked",
	})
	if err != nil {
		t.Fatalf("list support cases: %v", err)
	}
	for _, supportCase := range supportCases {
		if supportCase.CustomerVisible || supportCase.ExternalNotificationSent || supportCase.ExternalTicketCreated {
			t.Fatalf("expected support case %q external support effects to stay disabled", supportCase.ID)
		}
	}

	reviewPackets, err := client.ListReviewPackets(ctx, map[string]string{
		"account_id": "acct_local_qr_demo",
		"status":     "blocked_pending_matthew_approval",
	})
	if err != nil {
		t.Fatalf("list review packets: %v", err)
	}
	if len(reviewPackets) == 0 {
		t.Fatal("expected blocked review packet evidence")
	}

	operatorApprovals, err := client.ListOperatorApprovals(ctx, map[string]string{
		"account_id":        "acct_local_qr_demo",
		"decision_state":    "blocked",
		"required_approval": "matthew",
	})
	if err != nil {
		t.Fatalf("list operator approvals: %v", err)
	}
	if len(operatorApprovals) == 0 {
		t.Fatal("expected blocked operator approval evidence")
	}
	for _, approval := range operatorApprovals {
		if approval.DecisionState != "blocked" {
			t.Fatalf("expected operator approval %q to stay blocked, got %q", approval.DecisionID, approval.DecisionState)
		}
		if approval.Result != "not_executed" {
			t.Fatalf("expected operator approval %q to remain unexecuted, got %q", approval.DecisionID, approval.Result)
		}
	}

	auditEvents, err := client.ListAuditEvents(ctx, map[string]string{
		"account_id":       "acct_local_qr_demo",
		"source_packet_id": "local-qr-starter-demo",
	})
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(auditEvents) == 0 {
		t.Fatal("expected local QR seed audit evidence")
	}
}
