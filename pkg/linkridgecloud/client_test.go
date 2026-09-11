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
