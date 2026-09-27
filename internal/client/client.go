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
var ErrInUse = errors.New("resource is currently in use")

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

type EscalationPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Repeat      int64                  `json:"repeat"`
	Steps       []EscalationPolicyStep `json:"steps"`
}

type EscalationPolicyStep struct {
	ID           string        `json:"id"`
	StepNumber   int64         `json:"stepNumber"`
	DelayMinutes int64         `json:"delayMinutes"`
	Actions      []Destination `json:"actions"`
}

type Destination struct {
	Type string            `json:"type"`
	Args map[string]string `json:"args"`
}

type CreateEscalationPolicyInput struct {
	Name        string                            `json:"name"`
	Description *string                           `json:"description,omitempty"`
	Repeat      *int64                            `json:"repeat,omitempty"`
	Steps       []CreateEscalationPolicyStepInput `json:"steps,omitempty"`
}

type CreateEscalationPolicyStepInput struct {
	EscalationPolicyID *string            `json:"escalationPolicyID,omitempty"`
	DelayMinutes       int64              `json:"delayMinutes"`
	Targets            []TargetInput      `json:"targets,omitempty"`
	Actions            []DestinationInput `json:"actions,omitempty"`
}

type UpdateEscalationPolicyInput struct {
	ID          string    `json:"id"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Repeat      *int64    `json:"repeat,omitempty"`
	StepIDs     *[]string `json:"stepIDs,omitempty"`
}

type UpdateEscalationPolicyStepInput struct {
	ID           string             `json:"id"`
	DelayMinutes *int64             `json:"delayMinutes,omitempty"`
	Targets      []TargetInput      `json:"targets,omitempty"`
	Actions      []DestinationInput `json:"actions,omitempty"`
}

type DestinationInput struct {
	Type string            `json:"type"`
	Args map[string]string `json:"args,omitempty"`
}

type IntegrationKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Href      string `json:"href"`
	ServiceID string `json:"serviceID"`
}

type CreateIntegrationKeyInput struct {
	ServiceID string `json:"serviceID"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type HeartbeatMonitor struct {
	ID             string `json:"id"`
	ServiceID      string `json:"serviceID"`
	Name           string `json:"name"`
	TimeoutMinutes int64  `json:"timeoutMinutes"`
	Href           string `json:"href"`
}

type CreateHeartbeatMonitorInput struct {
	ServiceID      string `json:"serviceID"`
	Name           string `json:"name"`
	TimeoutMinutes int64  `json:"timeoutMinutes"`
}

type UpdateHeartbeatMonitorInput struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	TimeoutMinutes int64  `json:"timeoutMinutes"`
}

type ServiceLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetLabelInput struct {
	Target TargetInput `json:"target"`
	Key    string      `json:"key"`
	Value  string      `json:"value"`
}

type TargetInput struct {
	ID   string `json:"id"`
	Type string `json:"type"`
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnprocessableEntity {
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
			if e.Extensions.Code == "invalid_query" ||
				strings.Contains(e.Message, "wrong query for API key") ||
				(strings.Contains(e.Message, "operation") && strings.Contains(e.Message, "not found")) {
				return fmt.Errorf("%s: API key document mismatch; create a key with the canonical document shipped with this provider: %s", operation, e.Message)
			}
			if strings.Contains(e.Message, "currently in use") {
				return fmt.Errorf("%s: %w: %s", operation, ErrInUse, e.Message)
			}
		}
		return fmt.Errorf("%s: GraphQL rejected the operation; check the canonical document, admin role, key expiry and input values (server details suppressed)", operation)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP status %d; check endpoint and API key validity", operation, resp.StatusCode)
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

func escalationPolicyResult(raw json.RawMessage, allowMissing bool) (*EscalationPolicy, error) {
	if bytes.Equal(raw, []byte("null")) && allowMissing {
		return nil, ErrNotFound
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, errors.New("missing escalation policy result")
	}
	var ep EscalationPolicy
	if err := json.Unmarshal(raw, &ep); err != nil {
		return nil, errors.New("malformed escalation policy result")
	}
	if ep.ID == "" || ep.Name == "" {
		return nil, errors.New("incomplete escalation policy result")
	}
	if ep.Steps == nil {
		ep.Steps = []EscalationPolicyStep{}
	}
	for i := range ep.Steps {
		if ep.Steps[i].Actions == nil {
			ep.Steps[i].Actions = []Destination{}
		}
	}
	return &ep, nil
}

func (c *Client) ReadEscalationPolicy(ctx context.Context, id string) (*EscalationPolicy, error) {
	var result struct {
		EscalationPolicy json.RawMessage `json:"escalationPolicy"`
	}
	if err := c.execute(ctx, "ProviderReadEscalationPolicy", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	ep, err := escalationPolicyResult(result.EscalationPolicy, true)
	if err != nil {
		return nil, fmt.Errorf("read escalation policy: %w", err)
	}
	if ep.ID != id {
		return nil, errors.New("read escalation policy: response ID differs from requested ID")
	}
	return ep, nil
}

func (c *Client) CreateEscalationPolicy(ctx context.Context, input CreateEscalationPolicyInput) (*EscalationPolicy, error) {
	var result struct {
		EscalationPolicy json.RawMessage `json:"createEscalationPolicy"`
	}
	if err := c.execute(ctx, "ProviderCreateEscalationPolicy", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	ep, err := escalationPolicyResult(result.EscalationPolicy, false)
	if err != nil {
		return nil, fmt.Errorf("create escalation policy: %w", err)
	}
	return ep, nil
}

func (c *Client) UpdateEscalationPolicy(ctx context.Context, input UpdateEscalationPolicyInput) error {
	var result struct {
		Success bool `json:"updateEscalationPolicy"`
	}
	if err := c.execute(ctx, "ProviderUpdateEscalationPolicy", map[string]any{"input": input}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("update escalation policy: API did not confirm success")
	}
	return nil
}

func (c *Client) DeleteEscalationPolicy(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteEscalationPolicy", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete escalation policy: API did not confirm success")
	}
	return nil
}

func (c *Client) CreateEscalationPolicyStep(ctx context.Context, input CreateEscalationPolicyStepInput) (*EscalationPolicyStep, error) {
	var result struct {
		Step *EscalationPolicyStep `json:"createEscalationPolicyStep"`
	}
	if err := c.execute(ctx, "ProviderCreateEscalationPolicyStep", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	if result.Step == nil || result.Step.ID == "" {
		return nil, errors.New("create escalation policy step: missing step result")
	}
	if result.Step.Actions == nil {
		result.Step.Actions = []Destination{}
	}
	return result.Step, nil
}

func (c *Client) UpdateEscalationPolicyStep(ctx context.Context, input UpdateEscalationPolicyStepInput) error {
	var result struct {
		Success bool `json:"updateEscalationPolicyStep"`
	}
	if err := c.execute(ctx, "ProviderUpdateEscalationPolicyStep", map[string]any{"input": input}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("update escalation policy step: API did not confirm success")
	}
	return nil
}

func integrationKeyResult(raw json.RawMessage, allowMissing bool) (*IntegrationKey, error) {
	if bytes.Equal(raw, []byte("null")) && allowMissing {
		return nil, ErrNotFound
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, errors.New("missing integration key result")
	}
	var ik IntegrationKey
	if err := json.Unmarshal(raw, &ik); err != nil {
		return nil, errors.New("malformed integration key result")
	}
	if ik.ID == "" || ik.Name == "" || ik.Type == "" || ik.Href == "" || ik.ServiceID == "" {
		return nil, errors.New("incomplete integration key result")
	}
	return &ik, nil
}

func (c *Client) ReadIntegrationKey(ctx context.Context, id string) (*IntegrationKey, error) {
	var result struct {
		IntegrationKey json.RawMessage `json:"integrationKey"`
	}
	if err := c.execute(ctx, "ProviderReadIntegrationKey", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	ik, err := integrationKeyResult(result.IntegrationKey, true)
	if err != nil {
		return nil, fmt.Errorf("read integration key: %w", err)
	}
	if ik.ID != id {
		return nil, errors.New("read integration key: response ID differs from requested ID")
	}
	return ik, nil
}

func (c *Client) CreateIntegrationKey(ctx context.Context, input CreateIntegrationKeyInput) (*IntegrationKey, error) {
	var result struct {
		IntegrationKey json.RawMessage `json:"createIntegrationKey"`
	}
	if err := c.execute(ctx, "ProviderCreateIntegrationKey", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	ik, err := integrationKeyResult(result.IntegrationKey, false)
	if err != nil {
		return nil, fmt.Errorf("create integration key: %w", err)
	}
	return ik, nil
}

func (c *Client) DeleteIntegrationKey(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteIntegrationKey", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete integration key: API did not confirm success")
	}
	return nil
}

func heartbeatMonitorResult(raw json.RawMessage, allowMissing bool) (*HeartbeatMonitor, error) {
	if bytes.Equal(raw, []byte("null")) && allowMissing {
		return nil, ErrNotFound
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, errors.New("missing heartbeat monitor result")
	}
	var hb HeartbeatMonitor
	if err := json.Unmarshal(raw, &hb); err != nil {
		return nil, errors.New("malformed heartbeat monitor result")
	}
	if hb.ID == "" || hb.Name == "" || hb.Href == "" || hb.ServiceID == "" || hb.TimeoutMinutes <= 0 {
		return nil, errors.New("incomplete heartbeat monitor result")
	}
	return &hb, nil
}

func (c *Client) ReadHeartbeatMonitor(ctx context.Context, id string) (*HeartbeatMonitor, error) {
	var result struct {
		HeartbeatMonitor json.RawMessage `json:"heartbeatMonitor"`
	}
	if err := c.execute(ctx, "ProviderReadHeartbeatMonitor", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	hb, err := heartbeatMonitorResult(result.HeartbeatMonitor, true)
	if err != nil {
		return nil, fmt.Errorf("read heartbeat monitor: %w", err)
	}
	if hb != nil && hb.ID != id {
		return nil, errors.New("read heartbeat monitor: response ID differs from requested ID")
	}
	return hb, nil
}

func (c *Client) CreateHeartbeatMonitor(ctx context.Context, input CreateHeartbeatMonitorInput) (*HeartbeatMonitor, error) {
	var result struct {
		HeartbeatMonitor json.RawMessage `json:"createHeartbeatMonitor"`
	}
	if err := c.execute(ctx, "ProviderCreateHeartbeatMonitor", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	hb, err := heartbeatMonitorResult(result.HeartbeatMonitor, false)
	if err != nil {
		return nil, fmt.Errorf("create heartbeat monitor: %w", err)
	}
	return hb, nil
}

func (c *Client) UpdateHeartbeatMonitor(ctx context.Context, input UpdateHeartbeatMonitorInput) error {
	var result struct {
		Success bool `json:"updateHeartbeatMonitor"`
	}
	if err := c.execute(ctx, "ProviderUpdateHeartbeatMonitor", map[string]any{"input": input}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("update heartbeat monitor: API did not confirm success")
	}
	return nil
}

func (c *Client) DeleteHeartbeatMonitor(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteHeartbeatMonitor", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete heartbeat monitor: API did not confirm success")
	}
	return nil
}

func (c *Client) SetServiceLabel(ctx context.Context, serviceID, key, value string) error {
	input := SetLabelInput{
		Target: TargetInput{ID: serviceID, Type: "service"},
		Key:    key,
		Value:  value,
	}
	var result struct {
		Success bool `json:"setLabel"`
	}
	if err := c.execute(ctx, "ProviderSetServiceLabel", map[string]any{"input": input}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("set service label: API did not confirm success")
	}
	return nil
}

func (c *Client) ReadServiceLabels(ctx context.Context, serviceID string) (map[string]string, error) {
	var result struct {
		Service *struct {
			ID     string         `json:"id"`
			Labels []ServiceLabel `json:"labels"`
		} `json:"service"`
	}
	if err := c.execute(ctx, "ProviderReadServiceLabels", map[string]string{"id": serviceID}, &result); err != nil {
		return nil, err
	}
	if result.Service == nil {
		return nil, ErrNotFound
	}
	labels := make(map[string]string, len(result.Service.Labels))
	for _, l := range result.Service.Labels {
		labels[l.Key] = l.Value
	}
	return labels, nil
}

func (c *Client) SearchServices(ctx context.Context, search string) ([]Service, error) {
	var result struct {
		Services struct {
			Nodes []Service `json:"nodes"`
		} `json:"services"`
	}
	if err := c.execute(ctx, "ProviderSearchServices", map[string]string{"search": search}, &result); err != nil {
		return nil, err
	}
	return result.Services.Nodes, nil
}

func (c *Client) SearchEscalationPolicies(ctx context.Context, search string) ([]EscalationPolicy, error) {
	var result struct {
		EscalationPolicies struct {
			Nodes []EscalationPolicy `json:"nodes"`
		} `json:"escalationPolicies"`
	}
	if err := c.execute(ctx, "ProviderSearchEscalationPolicies", map[string]string{"search": search}, &result); err != nil {
		return nil, err
	}
	return result.EscalationPolicies.Nodes, nil
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateUserInput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UserContactMethod struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Disabled    bool        `json:"disabled"`
	Destination Destination `json:"dest"`
}

func (cm *UserContactMethod) Value() string {
	if cm.Destination.Args == nil {
		return ""
	}
	if v, ok := cm.Destination.Args["webhook_url"]; ok && v != "" {
		return v
	}
	if v, ok := cm.Destination.Args["phone_number"]; ok && v != "" {
		return v
	}
	if v, ok := cm.Destination.Args["email_address"]; ok && v != "" {
		return v
	}
	for _, v := range cm.Destination.Args {
		if v != "" {
			return v
		}
	}
	return ""
}

func (cm *UserContactMethod) Type() string {
	switch cm.Destination.Type {
	case "builtin-webhook":
		return "WEBHOOK"
	case "builtin-smtp-email":
		return "EMAIL"
	case "builtin-twilio-sms":
		return "SMS"
	case "builtin-twilio-voice":
		return "VOICE"
	case "builtin-slack-dm":
		return "SLACK_DM"
	default:
		return cm.Destination.Type
	}
}

type CreateUserContactMethodInput struct {
	UserID string `json:"userID"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type UpdateUserContactMethodInput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserNotificationRule struct {
	ID              string `json:"id"`
	DelayMinutes    int64  `json:"delayMinutes"`
	ContactMethodID string `json:"contactMethodID"`
}

type CreateUserNotificationRuleInput struct {
	UserID          string `json:"userID"`
	ContactMethodID string `json:"contactMethodID"`
	DelayMinutes    int64  `json:"delayMinutes"`
}

func (c *Client) ReadUser(ctx context.Context, id string) (*User, error) {
	var result struct {
		User *User `json:"user"`
	}
	if err := c.execute(ctx, "ProviderReadUser", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	if result.User == nil {
		return nil, ErrNotFound
	}
	return result.User, nil
}

func (c *Client) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	var result struct {
		User *User `json:"createUser"`
	}
	if err := c.execute(ctx, "ProviderCreateUser", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	if result.User == nil {
		return nil, errors.New("incomplete create user result")
	}
	return result.User, nil
}

func (c *Client) UpdateUser(ctx context.Context, input UpdateUserInput) error {
	var result struct {
		Success bool `json:"updateUser"`
	}
	return c.execute(ctx, "ProviderUpdateUser", map[string]any{"input": input}, &result)
}

func (c *Client) DeleteUser(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteUser", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete user: unconfirmed deletion")
	}
	return nil
}

func (c *Client) SearchUsers(ctx context.Context, search string) ([]User, error) {
	var result struct {
		Users struct {
			Nodes []User `json:"nodes"`
		} `json:"users"`
	}
	if err := c.execute(ctx, "ProviderSearchUsers", map[string]string{"search": search}, &result); err != nil {
		return nil, err
	}
	return result.Users.Nodes, nil
}

func (c *Client) ReadUserContactMethod(ctx context.Context, id string) (*UserContactMethod, error) {
	var result struct {
		CM *UserContactMethod `json:"userContactMethod"`
	}
	if err := c.execute(ctx, "ProviderReadUserContactMethod", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	if result.CM == nil {
		return nil, ErrNotFound
	}
	return result.CM, nil
}

func (c *Client) CreateUserContactMethod(ctx context.Context, input CreateUserContactMethodInput) (*UserContactMethod, error) {
	var result struct {
		CM *UserContactMethod `json:"createUserContactMethod"`
	}
	if err := c.execute(ctx, "ProviderCreateUserContactMethod", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	if result.CM == nil {
		return nil, errors.New("incomplete create contact method result")
	}
	return result.CM, nil
}

func (c *Client) UpdateUserContactMethod(ctx context.Context, input UpdateUserContactMethodInput) error {
	var result struct {
		Success bool `json:"updateUserContactMethod"`
	}
	return c.execute(ctx, "ProviderUpdateUserContactMethod", map[string]any{"input": input}, &result)
}

func (c *Client) DeleteUserContactMethod(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteUserContactMethod", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete contact method: unconfirmed deletion")
	}
	return nil
}

func (c *Client) ReadUserNotificationRules(ctx context.Context, userID string) ([]UserNotificationRule, error) {
	var result struct {
		User *struct {
			NotificationRules []struct {
				ID            string `json:"id"`
				DelayMinutes  int64  `json:"delayMinutes"`
				ContactMethod struct {
					ID string `json:"id"`
				} `json:"contactMethod"`
			} `json:"notificationRules"`
		} `json:"user"`
	}
	if err := c.execute(ctx, "ProviderReadUserNotificationRules", map[string]string{"userID": userID}, &result); err != nil {
		return nil, err
	}
	if result.User == nil {
		return nil, ErrNotFound
	}
	rules := make([]UserNotificationRule, 0, len(result.User.NotificationRules))
	for _, r := range result.User.NotificationRules {
		rules = append(rules, UserNotificationRule{
			ID:              r.ID,
			DelayMinutes:    r.DelayMinutes,
			ContactMethodID: r.ContactMethod.ID,
		})
	}
	return rules, nil
}

func (c *Client) ReadUserNotificationRule(ctx context.Context, userID, ruleID string) (*UserNotificationRule, error) {
	rules, err := c.ReadUserNotificationRules(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		if r.ID == ruleID {
			return &r, nil
		}
	}
	return nil, ErrNotFound
}

func (c *Client) CreateUserNotificationRule(ctx context.Context, input CreateUserNotificationRuleInput) (*UserNotificationRule, error) {
	var result struct {
		Rule *struct {
			ID            string `json:"id"`
			DelayMinutes  int64  `json:"delayMinutes"`
			ContactMethod struct {
				ID string `json:"id"`
			} `json:"contactMethod"`
		} `json:"createUserNotificationRule"`
	}
	if err := c.execute(ctx, "ProviderCreateUserNotificationRule", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	if result.Rule == nil {
		return nil, errors.New("incomplete create notification rule result")
	}
	return &UserNotificationRule{
		ID:              result.Rule.ID,
		DelayMinutes:    result.Rule.DelayMinutes,
		ContactMethodID: result.Rule.ContactMethod.ID,
	}, nil
}

func (c *Client) DeleteUserNotificationRule(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteUserNotificationRule", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete notification rule: unconfirmed deletion")
	}
	return nil
}

type Rotation struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Type            string   `json:"type"`
	Start           string   `json:"start"`
	TimeZone        string   `json:"timeZone"`
	ShiftLength     int64    `json:"shiftLength"`
	UserIDs         []string `json:"userIDs"`
	ActiveUserIndex int64    `json:"activeUserIndex"`
}

type CreateRotationInput struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Type        string   `json:"type"`
	Start       string   `json:"start"`
	TimeZone    string   `json:"timeZone"`
	ShiftLength *int64   `json:"shiftLength,omitempty"`
	UserIDs     []string `json:"userIDs"`
}

type UpdateRotationInput struct {
	ID          string    `json:"id"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Type        *string   `json:"type,omitempty"`
	Start       *string   `json:"start,omitempty"`
	TimeZone    *string   `json:"timeZone,omitempty"`
	ShiftLength *int64    `json:"shiftLength,omitempty"`
	UserIDs     *[]string `json:"userIDs,omitempty"`
}

func rotationResult(raw json.RawMessage, allowMissing bool) (*Rotation, error) {
	if bytes.Equal(raw, []byte("null")) && allowMissing {
		return nil, ErrNotFound
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, errors.New("missing rotation result")
	}
	var rot Rotation
	if err := json.Unmarshal(raw, &rot); err != nil {
		return nil, errors.New("malformed rotation result")
	}
	if rot.ID == "" || rot.Name == "" {
		return nil, errors.New("incomplete rotation result")
	}
	if rot.UserIDs == nil {
		rot.UserIDs = []string{}
	}
	return &rot, nil
}

func (c *Client) ReadRotation(ctx context.Context, id string) (*Rotation, error) {
	var result struct {
		Rotation json.RawMessage `json:"rotation"`
	}
	if err := c.execute(ctx, "ProviderReadRotation", map[string]string{"id": id}, &result); err != nil {
		return nil, err
	}
	rot, err := rotationResult(result.Rotation, true)
	if err != nil {
		return nil, fmt.Errorf("read rotation: %w", err)
	}
	if rot.ID != id {
		return nil, errors.New("read rotation: response ID differs from requested ID")
	}
	return rot, nil
}

func (c *Client) SearchRotations(ctx context.Context, search string) ([]Rotation, error) {
	var result struct {
		Rotations struct {
			Nodes []Rotation `json:"nodes"`
		} `json:"rotations"`
	}
	if err := c.execute(ctx, "ProviderSearchRotations", map[string]string{"search": search}, &result); err != nil {
		return nil, err
	}
	for i := range result.Rotations.Nodes {
		if result.Rotations.Nodes[i].UserIDs == nil {
			result.Rotations.Nodes[i].UserIDs = []string{}
		}
	}
	return result.Rotations.Nodes, nil
}

func (c *Client) CreateRotation(ctx context.Context, input CreateRotationInput) (*Rotation, error) {
	var result struct {
		Rotation json.RawMessage `json:"createRotation"`
	}
	if err := c.execute(ctx, "ProviderCreateRotation", map[string]any{"input": input}, &result); err != nil {
		return nil, err
	}
	rot, err := rotationResult(result.Rotation, false)
	if err != nil {
		return nil, fmt.Errorf("create rotation: %w", err)
	}
	return rot, nil
}

func (c *Client) UpdateRotation(ctx context.Context, input UpdateRotationInput) error {
	var result struct {
		Success bool `json:"updateRotation"`
	}
	if err := c.execute(ctx, "ProviderUpdateRotation", map[string]any{"input": input}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("update rotation: API did not confirm success")
	}
	return nil
}

func (c *Client) DeleteRotation(ctx context.Context, id string) error {
	var result struct {
		Success bool `json:"deleteAll"`
	}
	if err := c.execute(ctx, "ProviderDeleteRotation", map[string]string{"id": id}, &result); err != nil {
		return err
	}
	if !result.Success {
		return errors.New("delete rotation: API did not confirm success")
	}
	return nil
}
