// Copyright (c) LinkRidge
// SPDX-License-Identifier: MPL-2.0

// Package linkridgecloud provides a small API client for LinkRidge Cloud.
package linkridgecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a LinkRidge Cloud HTTP API client.
type Client struct {
	baseURL    string
	apiToken   string //nolint:gosec // API token is user-supplied provider configuration, not a hardcoded credential.
	httpClient *http.Client
}

// ClientConfig configures a Client.
type ClientConfig struct {
	BaseURL    string
	APIToken   string //nolint:gosec // API token is user-supplied provider configuration, not a hardcoded credential.
	HTTPClient *http.Client
}

// Service is the subset of the /v1/services representation exposed by the provider.
type Service struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	Plans       []string `json:"plans"`
}

// Account is the subset of the /v1/accounts representation exposed by the provider.
type Account struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Status                 string `json:"status"`
	PrimaryOwnerUserID     string `json:"primary_owner_user_id"`
	BillingCustomerID      string `json:"billing_customer_id"`
	SourcePacketID         string `json:"source_packet_id"`
	ExternalEffectsEnabled bool   `json:"external_effects_enabled"`
}

// AccountService is the subset of the /v1/accounts/{account_id}/services representation exposed by the provider.
type AccountService struct {
	ID                     string `json:"id"`
	AccountID              string `json:"account_id"`
	ServiceID              string `json:"service_id"`
	PlanID                 string `json:"plan_id"`
	PlanKey                string `json:"plan_key"`
	Status                 string `json:"status"`
	ApprovalRequired       bool   `json:"approval_required"`
	SourcePacketID         string `json:"source_packet_id"`
	ExternalEffectsEnabled bool   `json:"external_effects_enabled"`
}

// QRWorkspace is the subset of the /v1/qr/workspaces representation exposed by the provider.
type QRWorkspace struct {
	ID                       string `json:"id"`
	AccountID                string `json:"account_id"`
	AccountServiceID         string `json:"account_service_id"`
	ServiceID                string `json:"service_id"`
	PlanKey                  string `json:"plan_key"`
	Status                   string `json:"status"`
	SourcePacketID           string `json:"source_packet_id"`
	ExternalRedirectsEnabled bool   `json:"external_redirects_enabled"`
}

type servicesResponse struct {
	Data []Service `json:"data"`
}

type accountsResponse struct {
	Data []Account `json:"data"`
}

type accountServicesResponse struct {
	Data []AccountService `json:"data"`
}

type qrWorkspacesResponse struct {
	Data []QRWorkspace `json:"data"`
}

// NewClient creates a LinkRidge Cloud client.
func NewClient(cfg ClientConfig) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if strings.TrimSpace(cfg.APIToken) == "" {
		return nil, fmt.Errorf("API token is required")
	}

	baseURL, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("base URL must include scheme and host")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		baseURL:    baseURL.String(),
		apiToken:   cfg.APIToken,
		httpClient: httpClient,
	}, nil
}

// BaseURL returns the configured API base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// ListServices returns LinkRidge Cloud services. id is optional.
func (c *Client) ListServices(ctx context.Context, id string) ([]Service, error) {
	endpoint := c.listEndpoint("/v1/services", map[string]string{"id": id})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create services request: %w", err)
	}
	resp, err := c.doJSON(req, "list services")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result servicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode services response: %w", err)
	}
	if result.Data == nil {
		return []Service{}, nil
	}

	return result.Data, nil
}

// ListAccounts returns LinkRidge Cloud account planning records. id and status are optional.
func (c *Client) ListAccounts(ctx context.Context, id string, status string) ([]Account, error) {
	endpoint := c.listEndpoint("/v1/accounts", map[string]string{
		"id":     id,
		"status": status,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create accounts request: %w", err)
	}
	resp, err := c.doJSON(req, "list accounts")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result accountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode accounts response: %w", err)
	}
	if result.Data == nil {
		return []Account{}, nil
	}

	return result.Data, nil
}

// ListAccountServices returns service planning records for an account. accountID is required.
func (c *Client) ListAccountServices(ctx context.Context, accountID string, filters map[string]string) ([]AccountService, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	endpoint := c.listEndpoint("/v1/accounts/"+url.PathEscape(accountID)+"/services", filters)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create account services request: %w", err)
	}
	resp, err := c.doJSON(req, "list account services")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result accountServicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode account services response: %w", err)
	}
	if result.Data == nil {
		return []AccountService{}, nil
	}

	return result.Data, nil
}

// ListQRWorkspaces returns LinkRidge Cloud QR workspaces. filters are optional.
func (c *Client) ListQRWorkspaces(ctx context.Context, filters map[string]string) ([]QRWorkspace, error) {
	endpoint := c.listEndpoint("/v1/qr/workspaces", filters)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create QR workspaces request: %w", err)
	}
	resp, err := c.doJSON(req, "list QR workspaces")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result qrWorkspacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode QR workspaces response: %w", err)
	}
	if result.Data == nil {
		return []QRWorkspace{}, nil
	}

	return result.Data, nil
}

func (c *Client) listEndpoint(path string, filters map[string]string) string {
	endpoint := c.baseURL + path
	query := url.Values{}
	for key, value := range filters {
		if strings.TrimSpace(value) != "" {
			query.Set(key, value)
		}
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	return endpoint
}

func (c *Client) doJSON(req *http.Request, action string) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL is provider-configured and parsed before use.
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%s returned HTTP %d: %s", action, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return resp, nil
}
