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

type servicesResponse struct {
	Data []Service `json:"data"`
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
	endpoint := c.baseURL + "/v1/services"
	if strings.TrimSpace(id) != "" {
		query := url.Values{}
		query.Set("id", id)
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create services request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL is provider-configured and parsed before use.
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("list services returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result servicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode services response: %w", err)
	}
	if result.Data == nil {
		return []Service{}, nil
	}

	return result.Data, nil
}
