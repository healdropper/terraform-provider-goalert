// Package client implements the versioned, fixed-document GoAlert API contract.
package client

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Document is the complete policy to register when creating the GoAlert API key.
// Requests intentionally omit its text; GoAlert injects the key's stored document.
//
//go:embed operations.graphql
var Document string

var ErrNotFound = errors.New("service does not exist")

type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

type Service struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	EscalationPolicy *struct {
		ID string `json:"id"`
	} `json:"escalationPolicy"`
}

func New(endpoint, token string, allowHTTP bool) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u == nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("endpoint must be an absolute HTTP(S) GraphQL URL without credentials, query or fragment")
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if u.Scheme == "http" && !allowHTTP && host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, errors.New("use HTTPS, or explicitly enable allow_insecure_http for a trusted development connection")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("an API key is required")
	}
	return &Client{endpoint: endpoint, token: token, http: &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

func (c *Client) execute(ctx context.Context, operation string, variables any, target any) error {
	payload, err := json.Marshal(struct {
		Query     string `json:"query"`
		Operation string `json:"operationName"`
		Variables any    `json:"variables"`
	}{Operation: operation, Variables: variables})
	if err != nil {
		return fmt.Errorf("%s: encode request: %w", operation, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("%s: build request", operation)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "terraform-provider-goalert")
	// Mutations are never retried: a lost response may already have committed.
	resp, err := c.http.Do(req)
	if err != nil {
		// Do not expose URL, bearer token, response body or potentially echoed variables.
		if ctx.Err() != nil {
			return fmt.Errorf("%s: request cancelled: %w", operation, ctx.Err())
		}
		return fmt.Errorf("%s: HTTP request failed; check connectivity, TLS and the 30-second timeout", operation)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP status %d; check endpoint and API key validity", operation, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024+1))
	if err != nil || len(body) > 1024*1024 {
		return fmt.Errorf("%s: invalid or oversized response", operation)
	}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("%s: malformed GraphQL response", operation)
	}
	if len(envelope.Errors) > 0 {
		for _, e := range envelope.Errors {
			if e.Extensions.Code == "invalid_query" {
				return fmt.Errorf("%s: API key document mismatch; create a key with the canonical document shipped with this provider", operation)
			}
		}
		return fmt.Errorf("%s: GraphQL rejected the operation; check the canonical document, admin role, key expiry and input values (server details suppressed)", operation)
	}
	if len(envelope.Data) == 0 || bytes.Equal(envelope.Data, []byte("null")) {
		return fmt.Errorf("%s: missing GraphQL data", operation)
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		return fmt.Errorf("%s: malformed operation data", operation)
	}
	return nil
}

func serviceResult(raw json.RawMessage, allowMissing bool) (*Service, error) {
	if bytes.Equal(raw, []byte("null")) && allowMissing {
		return nil, ErrNotFound
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, errors.New("missing service result")
	}
	// Pointer fields distinguish a valid empty description from a missing field.
	var v struct {
		ID               *string `json:"id"`
		Name             *string `json:"name"`
		Description      *string `json:"description"`
		EscalationPolicy *struct {
			ID string `json:"id"`
		} `json:"escalationPolicy"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, errors.New("malformed service result")
	}
	if v.ID == nil || *v.ID == "" || v.Name == nil || *v.Name == "" || v.Description == nil ||
		v.EscalationPolicy == nil || v.EscalationPolicy.ID == "" {
		return nil, errors.New("incomplete service result")
	}
	return &Service{ID: *v.ID, Name: *v.Name, Description: *v.Description, EscalationPolicy: v.EscalationPolicy}, nil
}

func (c *Client) ReadService(ctx context.Context, id string) (*Service, error) {
	var result struct {
		Service json.RawMessage `json:"service"`
	}
	if err := c.execute(ctx, "ProviderReadService", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	svc, err := serviceResult(result.Service, true)
	if err != nil {
		return nil, fmt.Errorf("read service: %w", err)
	}
	if svc.ID != id {
		return nil, errors.New("read service: response ID differs from requested ID")
	}
	return svc, nil
}
func (c *Client) CreateService(ctx context.Context, name, description, policyID string) (*Service, error) {
	var result struct {
		Service json.RawMessage `json:"createService"`
	}
	if err := c.execute(ctx, "ProviderCreateService", map[string]string{
		"name": name, "description": description, "escalationPolicyID": policyID,
	}, &result); err != nil {
		return nil, err
	}
	svc, err := serviceResult(result.Service, false)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}
	return svc, nil
}
func (c *Client) UpdateService(ctx context.Context, id, name, description, policyID string) error {
	var result struct {
		Success bool `json:"updateService"`
	}
	if err := c.execute(ctx, "ProviderUpdateService", map[string]string{
		"id": id, "name": name, "description": description, "escalationPolicyID": policyID,
	}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("update service: API did not confirm success")
	}
	return nil
}
func (c *Client) DeleteService(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteService", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete service: API did not confirm success")
	}
	return nil
}
