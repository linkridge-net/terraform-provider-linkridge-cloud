// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

package linkridgecloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientValidatesRequiredConfig(t *testing.T) {
	if _, err := NewClient(ClientConfig{APIToken: "token"}); err == nil {
		t.Fatal("expected missing base URL error")
	}
	if _, err := NewClient(ClientConfig{BaseURL: "https://dev.cloud.linkridge.net"}); err == nil {
		t.Fatal("expected missing API token error")
	}
	if _, err := NewClient(ClientConfig{BaseURL: "not-a-url", APIToken: "token"}); err == nil {
		t.Fatal("expected invalid base URL error")
	}
}

func TestListServicesSendsBearerTokenAndFilter(t *testing.T) {
	var gotAuth string
	var gotID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotID = r.URL.Query().Get("id")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servicesResponse{
			Data: []Service{
				{
					ID:          "qr-codes",
					Name:        "QR Codes",
					Status:      "backing_store",
					Description: "Tenant-aware QR service previews.",
					Plans:       []string{"starter", "growth"},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL + "/", APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	services, err := client.ListServices(context.Background(), "qr-codes")
	if err != nil {
		t.Fatalf("expected services, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotID != "qr-codes" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if len(services) != 1 || services[0].ID != "qr-codes" {
		t.Fatalf("unexpected services: %#v", services)
	}
}

func TestListAccountsSendsFilters(t *testing.T) {
	var gotAuth string
	var gotID string
	var gotStatus string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotID = r.URL.Query().Get("id")
		gotStatus = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(accountsResponse{
			Data: []Account{
				{
					ID:                     "acct_local_qr_demo",
					Name:                   "Local QR Demo",
					Status:                 "draft",
					PrimaryOwnerUserID:     "usr_local_qr_owner",
					SourcePacketID:         "local-qr-starter-demo",
					ExternalEffectsEnabled: false,
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	accounts, err := client.ListAccounts(context.Background(), "acct_local_qr_demo", "draft")
	if err != nil {
		t.Fatalf("expected accounts, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotID != "acct_local_qr_demo" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotStatus != "draft" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if len(accounts) != 1 || accounts[0].ID != "acct_local_qr_demo" {
		t.Fatalf("unexpected accounts: %#v", accounts)
	}
}

func TestListAccountServicesRequiresAccountID(t *testing.T) {
	client, err := NewClient(ClientConfig{BaseURL: "https://dev.cloud.linkridge.net", APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	if _, err := client.ListAccountServices(context.Background(), "", nil); err == nil {
		t.Fatal("expected missing account ID error")
	}
}

func TestListAccountServicesSendsPathAndFilters(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotID string
	var gotStatus string
	var gotPlanKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.EscapedPath()
		gotID = r.URL.Query().Get("id")
		gotStatus = r.URL.Query().Get("status")
		gotPlanKey = r.URL.Query().Get("plan_key")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(accountServicesResponse{
			Data: []AccountService{
				{
					ID:                     "asvc_local_qr_demo",
					AccountID:              "acct local/qr demo",
					ServiceID:              "qr-codes",
					PlanID:                 "starter",
					PlanKey:                "starter",
					Status:                 "planned",
					ApprovalRequired:       true,
					SourcePacketID:         "local-qr-starter-demo",
					ExternalEffectsEnabled: false,
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	accountServices, err := client.ListAccountServices(context.Background(), "acct local/qr demo", map[string]string{
		"id":       "asvc_local_qr_demo",
		"status":   "planned",
		"plan_key": "starter",
	})
	if err != nil {
		t.Fatalf("expected account services, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotPath != "/v1/accounts/acct%20local%2Fqr%20demo/services" {
		t.Fatalf("expected escaped account services path, got %q", gotPath)
	}
	if gotID != "asvc_local_qr_demo" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotStatus != "planned" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if gotPlanKey != "starter" {
		t.Fatalf("expected plan_key filter, got %q", gotPlanKey)
	}
	if len(accountServices) != 1 || accountServices[0].ID != "asvc_local_qr_demo" {
		t.Fatalf("unexpected account services: %#v", accountServices)
	}
}

func TestListQRWorkspacesSendsFilters(t *testing.T) {
	var gotAccountID string
	var gotAccountServiceID string
	var gotStatus string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccountID = r.URL.Query().Get("account_id")
		gotAccountServiceID = r.URL.Query().Get("account_service_id")
		gotStatus = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(qrWorkspacesResponse{
			Data: []QRWorkspace{
				{
					ID:                       "qrw_local_demo",
					AccountID:                "acct_local_qr_demo",
					AccountServiceID:         "asvc_local_qr_demo",
					ServiceID:                "qr-codes",
					PlanKey:                  "starter",
					Status:                   "planned",
					SourcePacketID:           "local-qr-starter-demo",
					ExternalRedirectsEnabled: false,
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	workspaces, err := client.ListQRWorkspaces(context.Background(), map[string]string{
		"account_id":         "acct_local_qr_demo",
		"account_service_id": "asvc_local_qr_demo",
		"status":             "planned",
	})
	if err != nil {
		t.Fatalf("expected QR workspaces, got error: %v", err)
	}
	if gotAccountID != "acct_local_qr_demo" {
		t.Fatalf("expected account_id filter, got %q", gotAccountID)
	}
	if gotAccountServiceID != "asvc_local_qr_demo" {
		t.Fatalf("expected account_service_id filter, got %q", gotAccountServiceID)
	}
	if gotStatus != "planned" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if len(workspaces) != 1 || workspaces[0].ID != "qrw_local_demo" {
		t.Fatalf("unexpected QR workspaces: %#v", workspaces)
	}
}

func TestListServiceTokensRequiresAccountID(t *testing.T) {
	client, err := NewClient(ClientConfig{BaseURL: "https://dev.cloud.linkridge.net", APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	if _, err := client.ListServiceTokens(context.Background(), "", nil); err == nil {
		t.Fatal("expected missing account ID error")
	}
}

func TestListQRImportJobsRequiresWorkspaceID(t *testing.T) {
	client, err := NewClient(ClientConfig{BaseURL: "https://dev.cloud.linkridge.net", APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	if _, err := client.ListQRImportJobs(context.Background(), "", nil); err == nil {
		t.Fatal("expected missing workspace ID error")
	}
}

func TestListQRImportJobsSendsPathAndFilters(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotID string
	var gotStatus string
	var gotImportPerformed string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.EscapedPath()
		gotID = r.URL.Query().Get("id")
		gotStatus = r.URL.Query().Get("status")
		gotImportPerformed = r.URL.Query().Get("import_performed")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(qrImportJobsResponse{
			Data: []QRImportJob{
				{
					ID:                       "qrimp_local_demo_open_qr_links",
					AccountID:                "acct_local_qr_demo",
					AccountServiceID:         "asvc_local_qr_demo",
					QRWorkspaceID:            "qrw local/qr demo",
					ServiceID:                "qr-codes",
					Source:                   "open_qr_links_json",
					SourceRepository:         "https://gitlab.mfsoho.linkridge.net/makersridge/open-qr",
					RequestedByUserID:        "usr_local_qr_owner",
					Status:                   "planned",
					ImportPerformed:          false,
					RecordsExamined:          2,
					RecordsPlanned:           1,
					RecordsRejected:          1,
					SourcePacketID:           "local-qr-starter-demo",
					ExternalRedirectsEnabled: false,
					ImportReviewPacket: struct {
						Status                  string `json:"status"`
						ApprovalRequired        string `json:"approval_required"`
						ArtifactReviewed        bool   `json:"artifact_reviewed"`
						ParserContractChecked   bool   `json:"parser_contract_checked"`
						ExternalEffectPerformed bool   `json:"external_effect_performed"`
					}{
						Status:                  "blocked_pending_matthew_approval",
						ApprovalRequired:        "matthew",
						ArtifactReviewed:        false,
						ParserContractChecked:   true,
						ExternalEffectPerformed: false,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	importJobs, err := client.ListQRImportJobs(context.Background(), "qrw local/qr demo", map[string]string{
		"id":               "qrimp_local_demo_open_qr_links",
		"status":           "planned",
		"import_performed": "false",
	})
	if err != nil {
		t.Fatalf("expected QR import jobs, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotPath != "/v1/qr/workspaces/qrw%20local%2Fqr%20demo/import-jobs" {
		t.Fatalf("expected escaped QR import jobs path, got %q", gotPath)
	}
	if gotID != "qrimp_local_demo_open_qr_links" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotStatus != "planned" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if gotImportPerformed != "false" {
		t.Fatalf("expected import_performed filter, got %q", gotImportPerformed)
	}
	if len(importJobs) != 1 || importJobs[0].ID != "qrimp_local_demo_open_qr_links" {
		t.Fatalf("unexpected QR import jobs: %#v", importJobs)
	}
	if importJobs[0].ImportPerformed {
		t.Fatal("expected import planning metadata to report no performed import")
	}
	if importJobs[0].ImportReviewPacket.ExternalEffectPerformed {
		t.Fatal("expected import review packet to report no external effect")
	}
}

func TestListServiceTokensSendsPathAndFilters(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotID string
	var gotStatus string
	var gotSecretMaterialIssued string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.EscapedPath()
		gotID = r.URL.Query().Get("id")
		gotStatus = r.URL.Query().Get("status")
		gotSecretMaterialIssued = r.URL.Query().Get("secret_material_issued")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(serviceTokensResponse{
			Data: []ServiceToken{
				{
					ID:                     "stok_local_qr_demo_agent_preview",
					AccountID:              "acct local/qr demo",
					AccountServiceID:       "asvc_local_qr_demo",
					ServiceID:              "qr-codes",
					CreatedByUserID:        "usr_local_qr_owner",
					Name:                   "Local QR agent preview",
					Scopes:                 []string{"qr:read", "qr:plan_import"},
					Status:                 "planned",
					SecretMaterialIssued:   false,
					ApprovalRequired:       "matthew",
					SourcePacketID:         "local-qr-starter-demo",
					ExternalEffectsEnabled: false,
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	serviceTokens, err := client.ListServiceTokens(context.Background(), "acct local/qr demo", map[string]string{
		"id":                     "stok_local_qr_demo_agent_preview",
		"status":                 "planned",
		"secret_material_issued": "false",
	})
	if err != nil {
		t.Fatalf("expected service tokens, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotPath != "/v1/accounts/acct%20local%2Fqr%20demo/service-tokens" {
		t.Fatalf("expected escaped service tokens path, got %q", gotPath)
	}
	if gotID != "stok_local_qr_demo_agent_preview" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotStatus != "planned" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if gotSecretMaterialIssued != "false" {
		t.Fatalf("expected secret_material_issued filter, got %q", gotSecretMaterialIssued)
	}
	if len(serviceTokens) != 1 || serviceTokens[0].ID != "stok_local_qr_demo_agent_preview" {
		t.Fatalf("unexpected service tokens: %#v", serviceTokens)
	}
	if serviceTokens[0].SecretMaterialIssued {
		t.Fatal("expected service token metadata to report no issued secret material")
	}
}

func TestListActivationPacketsSendsPathAndFilters(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotID string
	var gotStatus string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.EscapedPath()
		gotID = r.URL.Query().Get("id")
		gotStatus = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(activationPacketsResponse{
			Data: []ActivationPacket{
				{
					ID:                     "local-qr-starter-demo",
					SourcePacketID:         "local-qr-starter-demo",
					Status:                 "blocked_pending_matthew_approval",
					ActivationPerformed:    false,
					ApprovalRequired:       "matthew",
					ExternalEffectsEnabled: false,
					Account: struct {
						ID                 string `json:"id"`
						Name               string `json:"name"`
						Status             string `json:"status"`
						PrimaryOwnerUserID string `json:"primary_owner_user_id"`
					}{
						ID:                 "acct local/qr demo",
						Name:               "Local QR Demo",
						Status:             "draft",
						PrimaryOwnerUserID: "usr_local_qr_owner",
					},
					AccountService: struct {
						ID        string `json:"id"`
						AccountID string `json:"account_id"`
						ServiceID string `json:"service_id"`
						PlanKey   string `json:"plan_key"`
						Status    string `json:"status"`
					}{
						ID:        "asvc local/qr demo",
						AccountID: "acct local/qr demo",
						ServiceID: "qr-codes",
						PlanKey:   "starter",
						Status:    "modeled",
					},
					ServiceWorkspace: struct {
						ID        string `json:"id"`
						AccountID string `json:"account_id"`
						ServiceID string `json:"service_id"`
						PlanKey   string `json:"plan_key"`
						Status    string `json:"status"`
					}{
						ID:        "qrw_local_demo",
						AccountID: "acct local/qr demo",
						ServiceID: "qr-codes",
						PlanKey:   "starter",
						Status:    "planned",
					},
					OperatorApprovalDecision: struct {
						DecisionID       string   `json:"decision_id"`
						DecisionState    string   `json:"decision_state"`
						RequiredApproval string   `json:"required_approval"`
						ApprovedActions  []string `json:"approved_actions"`
						BlockedActions   []string `json:"blocked_actions"`
						Result           string   `json:"result"`
					}{
						DecisionID:       "opd_local_qr_demo_001",
						DecisionState:    "blocked",
						RequiredApproval: "matthew",
						BlockedActions:   []string{"create_billing_customer", "run_production_deploy"},
						Result:           "not_executed",
					},
					LocalProvisioningRun: struct {
						RunID                      string   `json:"run_id"`
						Status                     string   `json:"status"`
						PlannedSteps               []string `json:"planned_steps"`
						LocalRecordsPrepared       []string `json:"local_records_prepared"`
						ExternalWritesBlocked      []string `json:"external_writes_blocked"`
						NextRequiredOperatorAction string   `json:"next_required_operator_action"`
					}{
						RunID:                      "lpv_local_qr_demo_001",
						Status:                     "rehearsal_only",
						PlannedSteps:               []string{"accounts", "qr_workspaces"},
						LocalRecordsPrepared:       []string{"accounts", "qr_workspaces"},
						ExternalWritesBlocked:      []string{"billing_customer", "external_qr_redirect"},
						NextRequiredOperatorAction: "review_activation_packet",
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	activationPackets, err := client.ListActivationPackets(context.Background(), "acct local/qr demo", "asvc local/qr demo", map[string]string{
		"id":     "local-qr-starter-demo",
		"status": "blocked_pending_matthew_approval",
	})
	if err != nil {
		t.Fatalf("expected activation packets, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotPath != "/v1/accounts/acct%20local%2Fqr%20demo/services/asvc%20local%2Fqr%20demo/activation-packets" {
		t.Fatalf("expected escaped activation packets path, got %q", gotPath)
	}
	if gotID != "local-qr-starter-demo" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotStatus != "blocked_pending_matthew_approval" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if len(activationPackets) != 1 || activationPackets[0].ID != "local-qr-starter-demo" {
		t.Fatalf("unexpected activation packets: %#v", activationPackets)
	}
	if activationPackets[0].ActivationPerformed {
		t.Fatal("expected activation packet to report no performed activation")
	}
	if activationPackets[0].ExternalEffectsEnabled {
		t.Fatal("expected activation packet to report external effects disabled")
	}
}

func TestListProvisioningRunsSendsFilters(t *testing.T) {
	var gotAuth string
	var gotAccountID string
	var gotAccountServiceID string
	var gotServiceID string
	var gotStatus string
	var gotSourcePacketID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccountID = r.URL.Query().Get("account_id")
		gotAccountServiceID = r.URL.Query().Get("account_service_id")
		gotServiceID = r.URL.Query().Get("service_id")
		gotStatus = r.URL.Query().Get("status")
		gotSourcePacketID = r.URL.Query().Get("source_packet_id")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(provisioningRunsResponse{
			Data: []ProvisioningRun{
				{
					ID:                         "lpv_local_qr_demo_001",
					AccountID:                  "acct_local_qr_demo",
					AccountServiceID:           "asvc_local_qr_demo",
					ServiceID:                  "qr-codes",
					QRWorkspaceID:              "qrw_local_demo",
					RequestedByUserID:          "usr_local_qr_owner",
					Status:                     "rehearsal_only",
					PlannedSteps:               []string{"accounts", "qr_workspaces"},
					LocalRecordsPrepared:       []string{"accounts", "qr_workspaces"},
					ExternalWritesBlocked:      []string{"billing_customer", "external_qr_redirect"},
					NextRequiredOperatorAction: "review_local_seed_packet",
					Result: ProvisioningRunResult{
						Result:                "not_executed",
						ExternalWritesBlocked: []string{"billing_customer", "external_qr_redirect"},
					},
					SourcePacketID:         "local-qr-starter-demo",
					ExternalEffectsEnabled: false,
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	provisioningRuns, err := client.ListProvisioningRuns(context.Background(), map[string]string{
		"account_id":         "acct_local_qr_demo",
		"account_service_id": "asvc_local_qr_demo",
		"service_id":         "qr-codes",
		"status":             "rehearsal_only",
		"source_packet_id":   "local-qr-starter-demo",
	})
	if err != nil {
		t.Fatalf("expected provisioning runs, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotAccountID != "acct_local_qr_demo" {
		t.Fatalf("expected account_id filter, got %q", gotAccountID)
	}
	if gotAccountServiceID != "asvc_local_qr_demo" {
		t.Fatalf("expected account_service_id filter, got %q", gotAccountServiceID)
	}
	if gotServiceID != "qr-codes" {
		t.Fatalf("expected service_id filter, got %q", gotServiceID)
	}
	if gotStatus != "rehearsal_only" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if gotSourcePacketID != "local-qr-starter-demo" {
		t.Fatalf("expected source_packet_id filter, got %q", gotSourcePacketID)
	}
	if len(provisioningRuns) != 1 || provisioningRuns[0].ID != "lpv_local_qr_demo_001" {
		t.Fatalf("unexpected provisioning runs: %#v", provisioningRuns)
	}
	if provisioningRuns[0].ExternalEffectsEnabled {
		t.Fatal("expected provisioning run to report external effects disabled")
	}
}

func TestProvisioningRunResultDecodesStringShape(t *testing.T) {
	var provisioningRun ProvisioningRun
	if err := json.Unmarshal([]byte(`{"result":"not_executed"}`), &provisioningRun); err != nil {
		t.Fatalf("expected string result shape to decode: %v", err)
	}
	if provisioningRun.Result.Result != "not_executed" {
		t.Fatalf("unexpected result: %q", provisioningRun.Result.Result)
	}
}

func TestListBillingExportRequestsSendsPathAndFilters(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotID string
	var gotStatus string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.EscapedPath()
		gotID = r.URL.Query().Get("id")
		gotStatus = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(billingExportRequestsResponse{
			Data: []BillingExportRequest{
				{
					ID:                    "qrbill_local_demo_scan_review",
					AccountID:             "acct local/qr demo",
					AccountServiceID:      "asvc local/qr demo",
					ServiceID:             "qr-codes",
					ScanEventIDs:          []string{"qrs_local_demo_welcome_planned"},
					Quantity:              1,
					Billable:              false,
					Status:                "blocked",
					RequestedByUserID:     "usr_local_qr_owner",
					SourcePacketID:        "local-qr-starter-demo",
					ExternalExportEnabled: false,
					Metadata: struct {
						BillingCustomerID     string `json:"billing_customer_id"`
						BillingSubscriptionID string `json:"billing_subscription_id"`
						ExternalUsageRecordID string `json:"external_usage_record_id"`
						ExportDestination     string `json:"export_destination"`
						ApprovalRequired      string `json:"approval_required"`
						BlockedReason         string `json:"blocked_reason"`
					}{
						ApprovalRequired: "matthew",
						BlockedReason:    "billing_export_not_approved",
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	requests, err := client.ListBillingExportRequests(context.Background(), "acct local/qr demo", "asvc local/qr demo", map[string]string{
		"id":     "qrbill_local_demo_scan_review",
		"status": "blocked",
	})
	if err != nil {
		t.Fatalf("expected billing export requests, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotPath != "/v1/accounts/acct%20local%2Fqr%20demo/services/asvc%20local%2Fqr%20demo/billing-export-requests" {
		t.Fatalf("expected escaped billing export requests path, got %q", gotPath)
	}
	if gotID != "qrbill_local_demo_scan_review" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotStatus != "blocked" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if len(requests) != 1 || requests[0].ID != "qrbill_local_demo_scan_review" {
		t.Fatalf("unexpected billing export requests: %#v", requests)
	}
	if requests[0].Billable {
		t.Fatal("expected billing export request to report non-billable usage")
	}
	if requests[0].ExternalExportEnabled {
		t.Fatal("expected billing export request to report external export disabled")
	}
}

func TestListSupportCasesSendsPathAndFilters(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotID string
	var gotCategory string
	var gotSeverity string
	var gotStatus string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.EscapedPath()
		gotID = r.URL.Query().Get("id")
		gotCategory = r.URL.Query().Get("category")
		gotSeverity = r.URL.Query().Get("severity")
		gotStatus = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(supportCasesResponse{
			Data: []SupportCase{
				{
					ID:                       "qrsup_local_demo_import_review",
					AccountID:                "acct local/qr demo",
					AccountServiceID:         "asvc local/qr demo",
					QRWorkspaceID:            "qrw_local_demo",
					ServiceID:                "qr-codes",
					TargetType:               "qr_import_job",
					TargetID:                 "qrimp_local_demo_open_qr_links",
					Category:                 "import_review",
					Severity:                 "low",
					Status:                   "blocked",
					Subject:                  "Review Starter Open QR import before tenant writes",
					CreatedByUserID:          "usr_local_qr_owner",
					SourcePacketID:           "local-qr-starter-demo",
					CustomerVisible:          false,
					ExternalNotificationSent: false,
					ExternalTicketCreated:    false,
					Metadata: struct {
						QRCodeID                 string   `json:"qr_code_id"`
						Slug                     string   `json:"slug"`
						CustomerVisible          bool     `json:"customer_visible"`
						ExternalTicketID         string   `json:"external_ticket_id"`
						ExternalNotificationSent bool     `json:"external_notification_sent"`
						EscalationPerformed      bool     `json:"escalation_performed"`
						ApprovalRequired         string   `json:"approval_required"`
						BlockedReason            string   `json:"blocked_reason"`
						BlockedExternalActions   []string `json:"blocked_external_actions"`
					}{
						QRCodeID:               "qrc_local_demo_welcome",
						Slug:                   "welcome",
						ApprovalRequired:       "matthew",
						BlockedReason:          "support_case_external_action_not_approved",
						BlockedExternalActions: []string{"create_external_ticket", "send_customer_notification"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	supportCases, err := client.ListSupportCases(context.Background(), "acct local/qr demo", "asvc local/qr demo", map[string]string{
		"id":       "qrsup_local_demo_import_review",
		"category": "import_review",
		"severity": "low",
		"status":   "blocked",
	})
	if err != nil {
		t.Fatalf("expected support cases, got error: %v", err)
	}
	if gotAuth != "Bearer dev-token" {
		t.Fatalf("expected bearer token header, got %q", gotAuth)
	}
	if gotPath != "/v1/accounts/acct%20local%2Fqr%20demo/services/asvc%20local%2Fqr%20demo/support-cases" {
		t.Fatalf("expected escaped support cases path, got %q", gotPath)
	}
	if gotID != "qrsup_local_demo_import_review" {
		t.Fatalf("expected id filter, got %q", gotID)
	}
	if gotCategory != "import_review" {
		t.Fatalf("expected category filter, got %q", gotCategory)
	}
	if gotSeverity != "low" {
		t.Fatalf("expected severity filter, got %q", gotSeverity)
	}
	if gotStatus != "blocked" {
		t.Fatalf("expected status filter, got %q", gotStatus)
	}
	if len(supportCases) != 1 || supportCases[0].ID != "qrsup_local_demo_import_review" {
		t.Fatalf("unexpected support cases: %#v", supportCases)
	}
	if supportCases[0].CustomerVisible {
		t.Fatal("expected support case to report customer visibility disabled")
	}
	if supportCases[0].ExternalTicketCreated {
		t.Fatal("expected support case to report no external ticket")
	}
}

func TestListServicesReturnsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, APIToken: "dev-token"})
	if err != nil {
		t.Fatalf("expected client, got error: %v", err)
	}

	if _, err := client.ListServices(context.Background(), ""); err == nil {
		t.Fatal("expected HTTP error")
	}
}
