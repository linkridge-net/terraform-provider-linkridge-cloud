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
