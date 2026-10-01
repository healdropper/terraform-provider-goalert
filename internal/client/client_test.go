package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const sid = "11111111-1111-4111-8111-111111111111"
const pid = "22222222-2222-4222-8222-222222222222"
const validService = `{"id":"` + sid + `","name":"api","description":"","escalationPolicy":{"id":"` + pid + `"}}`

func TestFixedDocumentProtocol(t *testing.T) {
	seen := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("invalid authentication protocol")
		}
		var body struct {
			Query     string
			Operation string `json:"operationName"`
			Variables map[string]string
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Query != "" {
			t.Error("must let GoAlert inject canonical document")
		}
		seen = append(seen, body.Operation)
		switch body.Operation {
		case "ProviderCreateService":
			if body.Variables["name"] != "api" || body.Variables["escalationPolicyID"] != pid {
				t.Error("incorrect create variables")
			}
			fmt.Fprint(w, `{"data":{"createService":`+validService+`}}`)
		case "ProviderReadService":
			if body.Variables["id"] != sid {
				t.Error("incorrect read ID")
			}
			fmt.Fprint(w, `{"data":{"service":`+validService+`}}`)
		case "ProviderUpdateService":
			if v, ok := body.Variables["description"]; !ok || v != "" {
				t.Error("must transmit empty description to clear it")
			}
			fmt.Fprint(w, `{"data":{"updateService":true}}`)
		case "ProviderDeleteService":
			if body.Variables["id"] != sid {
				t.Error("incorrect delete ID")
			}
			fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
		default:
			t.Error("unknown operation")
		}
	}))
	defer server.Close()
	c, err := New(server.URL, "test-token", false)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	created, err := c.CreateService(ctx, "api", "", pid)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != sid {
		t.Fatal("incorrect ID")
	}
	if _, err = c.ReadService(ctx, sid); err != nil {
		t.Fatal(err)
	}
	if err = c.UpdateService(ctx, sid, "api", "", pid); err != nil {
		t.Fatal(err)
	}
	if err = c.DeleteService(ctx, sid); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 4 {
		t.Fatal("unexpected calls/retries")
	}
	for _, op := range seen {
		if !strings.Contains(Document, op+"(") {
			t.Errorf("operation %s absent from key document", op)
		}
	}
}

func TestReadFailureNeverMeansDeleted(t *testing.T) {
	tests := []struct {
		name, body string
		status     int
		notFound   bool
	}{
		{"deleted", `{"data":{"service":null}}`, 200, true},
		{"unauthorized null", `{"data":{"service":null},"errors":[{"message":"test-secret","extensions":{"code":"unauthorized"}}]}`, 200, false},
		{"wrong document", `{"errors":[{"message":"wrong query for API key","extensions":{"code":"invalid_query"}}]}`, 200, false},
		{"missing field", `{"data":{}}`, 200, false},
		{"missing description", `{"data":{"service":{"id":"x","name":"api","escalationPolicy":{"id":"y"}}}}`, 200, false},
		{"missing policy", `{"data":{"service":{"id":"x","name":"api","description":"","escalationPolicy":null}}}`, 200, false},
		{"wrong ID", `{"data":{"service":` + strings.Replace(validService, sid, pid, 1) + `}}`, 200, false},
		{"missing data", `{}`, 200, false},
		{"null data", `{"data":null}`, 200, false},
		{"malformed", `<html>test-secret</html>`, 200, false},
		{"unauthenticated", `test-secret`, 401, false},
		{"forbidden", `test-secret`, 403, false},
		{"endpoint missing", `test-secret`, 404, false},
		{"server error", `test-secret`, 503, false},
		{"oversized", strings.Repeat("x", 1024*1024+1), 200, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			c, _ := New(server.URL, "test-secret", false)
			_, err := c.ReadService(context.Background(), sid)
			if err == nil {
				t.Fatal("expected error")
			}
			if errors.Is(err, ErrNotFound) != tt.notFound {
				t.Fatalf("incorrect absence classification: %v", err)
			}
			if strings.Contains(err.Error(), "test-secret") {
				t.Fatal("secret leaked")
			}
			if calls != 1 {
				t.Fatal("unexpected retry")
			}
		})
	}
}
func TestMutationRequiresConfirmation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"data":{}}`) }))
	defer server.Close()
	c, _ := New(server.URL, "token", false)
	if _, err := c.CreateService(context.Background(), "api", "", pid); err == nil {
		t.Fatal("accepted missing create result")
	}
	if err := c.UpdateService(context.Background(), sid, "api", "", pid); err == nil {
		t.Fatal("accepted missing update result")
	}
	if err := c.DeleteService(context.Background(), sid); err == nil {
		t.Fatal("accepted missing delete result")
	}
}
func TestRedirectIsNotFollowed(t *testing.T) {
	followed := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { followed = true }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
	defer server.Close()
	c, _ := New(server.URL, "token", false)
	if _, err := c.ReadService(context.Background(), sid); err == nil {
		t.Fatal("accepted redirect")
	}
	if followed {
		t.Fatal("forwarded credentials to redirect")
	}
}
func TestCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(200 * time.Millisecond):
		}
	}))
	defer server.Close()
	c, _ := New(server.URL, "token", false)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := c.ReadService(ctx, sid)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancellation not preserved: %v", err)
	}
}
func TestConfiguration(t *testing.T) {
	for _, endpoint := range []string{"", "://", "ftp://example.com", "https://user:secret@example.com", "https://example.com?token=secret", "https://example.com/#secret", "http://example.com/api/graphql"} {
		if _, err := New(endpoint, "token", false); err == nil {
			t.Errorf("accepted unsafe endpoint %s", endpoint)
		}
	}
	for _, endpoint := range []string{"https://example.com/prefix/api/graphql", "http://localhost:8081/api/graphql", "http://127.0.0.1/api/graphql", "http://[::1]/api/graphql"} {
		if _, err := New(endpoint, "token", false); err != nil {
			t.Errorf("valid endpoint: %v", err)
		}
	}
	if _, err := New("http://example.com/api/graphql", "token", true); err != nil {
		t.Fatal(err)
	}
	if _, err := New("https://example.com/api/graphql", " ", false); err == nil {
		t.Fatal("accepted empty token")
	}
}

func TestEscalationPolicyProtocol(t *testing.T) {
	const epID = "33333333-3333-4333-8333-333333333333"
	const stepID = "44444444-4444-4444-8444-444444444444"
	validEP := `{"id":"` + epID + `","name":"critical","description":"on-call alerts","repeat":2,"steps":[{"id":"` + stepID + `","stepNumber":0,"delayMinutes":5,"actions":[{"type":"builtin-webhook","args":{"webhook_url":"https://example.com/alerts"}}]}]}`

	seen := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("invalid authentication protocol")
		}
		var body struct {
			Query     string          `json:"query"`
			Operation string          `json:"operationName"`
			Variables json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Query != "" {
			t.Error("must let GoAlert inject canonical document")
		}
		seen = append(seen, body.Operation)
		switch body.Operation {
		case "ProviderCreateEscalationPolicy":
			fmt.Fprint(w, `{"data":{"createEscalationPolicy":`+validEP+`}}`)
		case "ProviderReadEscalationPolicy":
			fmt.Fprint(w, `{"data":{"escalationPolicy":`+validEP+`}}`)
		case "ProviderUpdateEscalationPolicy":
			fmt.Fprint(w, `{"data":{"updateEscalationPolicy":true}}`)
		case "ProviderDeleteEscalationPolicy":
			fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
		case "ProviderCreateEscalationPolicyStep":
			fmt.Fprint(w, `{"data":{"createEscalationPolicyStep":{"id":"`+stepID+`","stepNumber":1,"delayMinutes":10,"actions":[]}}}`)
		case "ProviderUpdateEscalationPolicyStep":
			fmt.Fprint(w, `{"data":{"updateEscalationPolicyStep":true}}`)
		default:
			t.Errorf("unexpected operation: %s", body.Operation)
		}
	}))
	defer server.Close()

	c, err := New(server.URL, "test-token", false)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	desc := "on-call alerts"
	rep := int64(2)
	created, err := c.CreateEscalationPolicy(ctx, CreateEscalationPolicyInput{
		Name:        "critical",
		Description: &desc,
		Repeat:      &rep,
		Steps: []CreateEscalationPolicyStepInput{
			{
				DelayMinutes: 5,
				Actions: []DestinationInput{
					{Type: "builtin-webhook", Args: map[string]string{"webhook_url": "https://example.com/alerts"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateEscalationPolicy: %v", err)
	}
	if created.ID != epID || len(created.Steps) != 1 || created.Steps[0].ID != stepID {
		t.Fatalf("unexpected created policy: %+v", created)
	}

	read, err := c.ReadEscalationPolicy(ctx, epID)
	if err != nil {
		t.Fatalf("ReadEscalationPolicy: %v", err)
	}
	if read.Name != "critical" || read.Repeat != 2 || len(read.Steps) != 1 {
		t.Fatalf("unexpected read policy: %+v", read)
	}

	if err = c.UpdateEscalationPolicy(ctx, UpdateEscalationPolicyInput{
		ID:      epID,
		Name:    &created.Name,
		StepIDs: &[]string{stepID},
	}); err != nil {
		t.Fatalf("UpdateEscalationPolicy: %v", err)
	}

	policyID := epID
	step, err := c.CreateEscalationPolicyStep(ctx, CreateEscalationPolicyStepInput{
		EscalationPolicyID: &policyID,
		DelayMinutes:       10,
	})
	if err != nil {
		t.Fatalf("CreateEscalationPolicyStep: %v", err)
	}
	if step.StepNumber != 1 {
		t.Fatalf("unexpected step number: %d", step.StepNumber)
	}

	delMinutes := int64(15)
	if err = c.UpdateEscalationPolicyStep(ctx, UpdateEscalationPolicyStepInput{
		ID:           stepID,
		DelayMinutes: &delMinutes,
	}); err != nil {
		t.Fatalf("UpdateEscalationPolicyStep: %v", err)
	}

	if err = c.DeleteEscalationPolicy(ctx, epID); err != nil {
		t.Fatalf("DeleteEscalationPolicy: %v", err)
	}

	if len(seen) != 6 {
		t.Fatalf("expected 6 operations, saw %d", len(seen))
	}
	for _, op := range seen {
		if !strings.Contains(Document, op+"(") {
			t.Errorf("operation %s absent from key document", op)
		}
	}
}

func TestEscalationPolicyNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":{"escalationPolicy":null}}`)
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", false)
	_, err := c.ReadEscalationPolicy(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestKeyDocumentMismatchAndInUseErrors(t *testing.T) {
	t.Run("missing operation error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprint(w, `{"errors":[{"message":"operation ProviderReadEscalationPolicy not found","extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}]}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", false)
		_, err := c.ReadEscalationPolicy(context.Background(), "00000000-0000-0000-0000-000000000000")
		if err == nil || !strings.Contains(err.Error(), "API key document mismatch") {
			t.Fatalf("expected API key document mismatch error, got: %v", err)
		}
	})

	t.Run("currently in use error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"errors":[{"message":"resource is currently in use"}]}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", false)
		err := c.DeleteEscalationPolicy(context.Background(), "00000000-0000-0000-0000-000000000000")
		if err == nil || !errors.Is(err, ErrInUse) {
			t.Fatalf("expected ErrInUse, got: %v", err)
		}
	})
}

func TestIntegrationKeyOperations(t *testing.T) {
	const ikID = "33333333-3333-4333-8333-333333333333"
	const ikServiceID = "44444444-4444-4444-8444-444444444444"
	const ikHref = "http://goalert.example.com/api/v2/grafana/incoming?token=" + ikID

	t.Run("create read and delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Operation string `json:"operationName"`
				Variables json.RawMessage
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch body.Operation {
			case "ProviderCreateIntegrationKey":
				var vars struct {
					Input CreateIntegrationKeyInput `json:"input"`
				}
				_ = json.Unmarshal(body.Variables, &vars)
				if vars.Input.ServiceID != ikServiceID || vars.Input.Name != "Grafana Key" || vars.Input.Type != "grafana" {
					t.Errorf("unexpected create input: %+v", vars.Input)
				}
				fmt.Fprintf(w, `{"data":{"createIntegrationKey":{"id":"%s","name":"Grafana Key","type":"grafana","href":"%s","serviceID":"%s"}}}`,
					ikID, ikHref, ikServiceID)
			case "ProviderReadIntegrationKey":
				var vars struct {
					ID string `json:"id"`
				}
				_ = json.Unmarshal(body.Variables, &vars)
				if vars.ID != ikID {
					t.Errorf("unexpected read ID: %s", vars.ID)
				}
				fmt.Fprintf(w, `{"data":{"integrationKey":{"id":"%s","name":"Grafana Key","type":"grafana","href":"%s","serviceID":"%s"}}}`,
					ikID, ikHref, ikServiceID)
			case "ProviderDeleteIntegrationKey":
				var vars struct {
					ID string `json:"id"`
				}
				_ = json.Unmarshal(body.Variables, &vars)
				if vars.ID != ikID {
					t.Errorf("unexpected delete ID: %s", vars.ID)
				}
				fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
			default:
				t.Errorf("unexpected operation: %s", body.Operation)
			}
		}))
		defer server.Close()

		c, err := New(server.URL, "token", true)
		if err != nil {
			t.Fatal(err)
		}

		created, err := c.CreateIntegrationKey(context.Background(), CreateIntegrationKeyInput{
			ServiceID: ikServiceID,
			Name:      "Grafana Key",
			Type:      "grafana",
		})
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}
		if created.ID != ikID || created.Href != ikHref || created.ServiceID != ikServiceID {
			t.Fatalf("unexpected created key: %+v", created)
		}

		read, err := c.ReadIntegrationKey(context.Background(), ikID)
		if err != nil {
			t.Fatalf("unexpected read error: %v", err)
		}
		if read.ID != ikID || read.Name != "Grafana Key" || read.Type != "grafana" {
			t.Fatalf("unexpected read key: %+v", read)
		}

		if err := c.DeleteIntegrationKey(context.Background(), ikID); err != nil {
			t.Fatalf("unexpected delete error: %v", err)
		}
	})

	t.Run("read not found returns ErrNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"integrationKey":null}}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		_, err := c.ReadIntegrationKey(context.Background(), ikID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("delete unconfirmed failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"deleteAll":false}}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		err := c.DeleteIntegrationKey(context.Background(), ikID)
		if err == nil || !strings.Contains(err.Error(), "did not confirm success") {
			t.Fatalf("expected unconfirmed success error, got: %v", err)
		}
	})
}

func TestHeartbeatMonitorLifecycle(t *testing.T) {
	const hbID = "33333333-3333-4333-8333-333333333333"
	const hbServiceID = "11111111-1111-4111-8111-111111111111"
	const hbHref = "http://127.0.0.1:18081/api/v2/heartbeat/" + hbID

	t.Run("full CRUD lifecycle", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Operation string `json:"operationName"`
				Variables map[string]any
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch body.Operation {
			case "ProviderCreateHeartbeatMonitor":
				input := body.Variables["input"].(map[string]any)
				if input["name"] != "Worker Monitor" || input["serviceID"] != hbServiceID || int64(input["timeoutMinutes"].(float64)) != 15 {
					t.Errorf("unexpected create input: %+v", input)
				}
				fmt.Fprint(w, `{"data":{"createHeartbeatMonitor":{"id":"`+hbID+`","serviceID":"`+hbServiceID+`","name":"Worker Monitor","timeoutMinutes":15,"href":"`+hbHref+`"}}}`)
			case "ProviderReadHeartbeatMonitor":
				if body.Variables["id"] != hbID {
					t.Errorf("unexpected read ID: %v", body.Variables["id"])
				}
				fmt.Fprint(w, `{"data":{"heartbeatMonitor":{"id":"`+hbID+`","serviceID":"`+hbServiceID+`","name":"Worker Monitor","timeoutMinutes":15,"href":"`+hbHref+`"}}}`)
			case "ProviderUpdateHeartbeatMonitor":
				input := body.Variables["input"].(map[string]any)
				if input["id"] != hbID || input["name"] != "Renamed Monitor" || int64(input["timeoutMinutes"].(float64)) != 30 {
					t.Errorf("unexpected update input: %+v", input)
				}
				fmt.Fprint(w, `{"data":{"updateHeartbeatMonitor":true}}`)
			case "ProviderDeleteHeartbeatMonitor":
				if body.Variables["id"] != hbID {
					t.Errorf("unexpected delete ID: %v", body.Variables["id"])
				}
				fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
			default:
				t.Errorf("unexpected operation: %s", body.Operation)
			}
		}))
		defer server.Close()

		c, err := New(server.URL, "token", true)
		if err != nil {
			t.Fatal(err)
		}

		created, err := c.CreateHeartbeatMonitor(context.Background(), CreateHeartbeatMonitorInput{
			ServiceID:      hbServiceID,
			Name:           "Worker Monitor",
			TimeoutMinutes: 15,
		})
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}
		if created.ID != hbID || created.Name != "Worker Monitor" || created.TimeoutMinutes != 15 || created.Href != hbHref {
			t.Fatalf("unexpected created monitor: %+v", created)
		}

		read, err := c.ReadHeartbeatMonitor(context.Background(), hbID)
		if err != nil {
			t.Fatalf("unexpected read error: %v", err)
		}
		if read.ID != hbID || read.Name != "Worker Monitor" {
			t.Fatalf("unexpected read monitor: %+v", read)
		}

		err = c.UpdateHeartbeatMonitor(context.Background(), UpdateHeartbeatMonitorInput{
			ID:             hbID,
			Name:           "Renamed Monitor",
			TimeoutMinutes: 30,
		})
		if err != nil {
			t.Fatalf("unexpected update error: %v", err)
		}

		err = c.DeleteHeartbeatMonitor(context.Background(), hbID)
		if err != nil {
			t.Fatalf("unexpected delete error: %v", err)
		}
	})

	t.Run("read not found returns ErrNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"heartbeatMonitor":null}}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		_, err := c.ReadHeartbeatMonitor(context.Background(), hbID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestServiceLabels(t *testing.T) {
	const testServiceID = "11111111-1111-4111-8111-111111111111"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Operation string `json:"operationName"`
			Variables map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		switch body.Operation {
		case "ProviderSetServiceLabel":
			fmt.Fprint(w, `{"data":{"setLabel":true}}`)
		case "ProviderReadServiceLabels":
			fmt.Fprint(w, `{"data":{"service":{"id":"`+testServiceID+`","labels":[{"key":"example.com/env","value":"prod"}]}}}`)
		default:
			t.Errorf("unexpected operation: %s", body.Operation)
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", true)
	if err := c.SetServiceLabel(context.Background(), testServiceID, "example.com/env", "prod"); err != nil {
		t.Fatalf("unexpected set label error: %v", err)
	}

	labels, err := c.ReadServiceLabels(context.Background(), testServiceID)
	if err != nil {
		t.Fatalf("unexpected read labels error: %v", err)
	}
	if labels["example.com/env"] != "prod" {
		t.Fatalf("expected example.com/env=prod, got: %v", labels)
	}
}

func TestSearchQueries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Operation string `json:"operationName"`
			Variables map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		switch body.Operation {
		case "ProviderSearchServices":
			fmt.Fprint(w, `{"data":{"services":{"nodes":[{"id":"`+sid+`","name":"api","description":"","escalationPolicy":{"id":"`+pid+`"}}]}}}`)
		case "ProviderSearchEscalationPolicies":
			fmt.Fprint(w, `{"data":{"escalationPolicies":{"nodes":[{"id":"`+pid+`","name":"policy","description":"","repeat":3}]}}}`)
		default:
			t.Errorf("unexpected operation: %s", body.Operation)
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", true)
	services, err := c.SearchServices(context.Background(), "api")
	if err != nil || len(services) != 1 || services[0].ID != sid {
		t.Fatalf("unexpected services search result: %v, err: %v", services, err)
	}

	policies, err := c.SearchEscalationPolicies(context.Background(), "policy")
	if err != nil || len(policies) != 1 || policies[0].ID != pid {
		t.Fatalf("unexpected policies search result: %v, err: %v", policies, err)
	}
}

func TestUserOperations(t *testing.T) {
	const userID = "44444444-4444-4444-8444-444444444444"

	t.Run("full user lifecycle", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Operation string `json:"operationName"`
				Variables map[string]any
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch body.Operation {
			case "ProviderCreateUser":
				fmt.Fprint(w, `{"data":{"createUser":{"id":"`+userID+`","name":"Alice","email":"alice@example.com","role":"user"}}}`)
			case "ProviderReadUser":
				fmt.Fprint(w, `{"data":{"user":{"id":"`+userID+`","name":"Alice Senior","email":"alice@example.com","role":"admin"}}}`)
			case "ProviderUpdateUser":
				fmt.Fprint(w, `{"data":{"updateUser":true}}`)
			case "ProviderDeleteUser":
				fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
			case "ProviderSearchUsers":
				fmt.Fprint(w, `{"data":{"users":{"nodes":[{"id":"`+userID+`","name":"Alice","email":"alice@example.com","role":"user"}]}}}`)
			default:
				t.Errorf("unexpected operation: %s", body.Operation)
			}
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		u, err := c.CreateUser(context.Background(), CreateUserInput{
			Name:     "Alice",
			Email:    "alice@example.com",
			Role:     "user",
			Username: "alice",
			Password: "Password123!",
		})
		if err != nil || u.ID != userID {
			t.Fatalf("unexpected create user: %v, err: %v", u, err)
		}

		if err := c.UpdateUser(context.Background(), UpdateUserInput{
			ID:    userID,
			Name:  "Alice Senior",
			Email: "alice@example.com",
			Role:  "admin",
		}); err != nil {
			t.Fatalf("unexpected update user error: %v", err)
		}

		readU, err := c.ReadUser(context.Background(), userID)
		if err != nil || readU.Role != "admin" {
			t.Fatalf("unexpected read user: %v, err: %v", readU, err)
		}

		users, err := c.SearchUsers(context.Background(), "Alice")
		if err != nil || len(users) != 1 {
			t.Fatalf("unexpected search users: %v, err: %v", users, err)
		}

		if err := c.DeleteUser(context.Background(), userID); err != nil {
			t.Fatalf("unexpected delete user error: %v", err)
		}
	})

	t.Run("read not found returns ErrNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"user":null}}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		_, err := c.ReadUser(context.Background(), userID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestUserContactMethodOperations(t *testing.T) {
	const cmID = "55555555-5555-4555-8555-555555555555"
	const uID = "44444444-4444-4444-8444-444444444444"

	t.Run("full contact method lifecycle", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Operation string `json:"operationName"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch body.Operation {
			case "ProviderCreateUserContactMethod":
				fmt.Fprint(w, `{"data":{"createUserContactMethod":{"id":"`+cmID+`","name":"Ops Webhook","disabled":false,"dest":{"type":"builtin-webhook","args":{"webhook_url":"https://example.com/hook"}}}}}`)
			case "ProviderReadUserContactMethod":
				fmt.Fprint(w, `{"data":{"userContactMethod":{"id":"`+cmID+`","name":"Ops Webhook","disabled":false,"dest":{"type":"builtin-webhook","args":{"webhook_url":"https://example.com/hook"}}}}}`)
			case "ProviderUpdateUserContactMethod":
				fmt.Fprint(w, `{"data":{"updateUserContactMethod":true}}`)
			case "ProviderDeleteUserContactMethod":
				fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
			default:
				t.Errorf("unexpected operation: %s", body.Operation)
			}
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		cm, err := c.CreateUserContactMethod(context.Background(), CreateUserContactMethodInput{
			UserID: uID,
			Name:   "Ops Webhook",
			Type:   "WEBHOOK",
			Value:  "https://example.com/hook",
		})
		if err != nil || cm.ID != cmID || cm.Value() != "https://example.com/hook" || cm.Type() != "WEBHOOK" {
			t.Fatalf("unexpected create contact method: %v, err: %v", cm, err)
		}

		if err := c.UpdateUserContactMethod(context.Background(), UpdateUserContactMethodInput{
			ID:   cmID,
			Name: "Ops Webhook Updated",
		}); err != nil {
			t.Fatalf("unexpected update error: %v", err)
		}

		readCM, err := c.ReadUserContactMethod(context.Background(), cmID)
		if err != nil || readCM.ID != cmID {
			t.Fatalf("unexpected read contact method: %v, err: %v", readCM, err)
		}

		if err := c.DeleteUserContactMethod(context.Background(), cmID); err != nil {
			t.Fatalf("unexpected delete error: %v", err)
		}
	})

	t.Run("read not found returns ErrNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"userContactMethod":null}}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		_, err := c.ReadUserContactMethod(context.Background(), cmID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestUserNotificationRuleOperations(t *testing.T) {
	const nrID = "66666666-6666-4666-8666-666666666666"
	const uID = "44444444-4444-4444-8444-444444444444"
	const cmID = "55555555-5555-4555-8555-555555555555"

	t.Run("notification rule lifecycle", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Operation string `json:"operationName"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch body.Operation {
			case "ProviderCreateUserNotificationRule":
				fmt.Fprint(w, `{"data":{"createUserNotificationRule":{"id":"`+nrID+`","delayMinutes":10,"contactMethod":{"id":"`+cmID+`"}}}}`)
			case "ProviderReadUserNotificationRules":
				fmt.Fprint(w, `{"data":{"user":{"id":"`+uID+`","notificationRules":[{"id":"`+nrID+`","delayMinutes":10,"contactMethod":{"id":"`+cmID+`"}}]}}}`)
			case "ProviderDeleteUserNotificationRule":
				fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
			default:
				t.Errorf("unexpected operation: %s", body.Operation)
			}
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		nr, err := c.CreateUserNotificationRule(context.Background(), CreateUserNotificationRuleInput{
			UserID:          uID,
			ContactMethodID: cmID,
			DelayMinutes:    10,
		})
		if err != nil || nr.ID != nrID || nr.DelayMinutes != 10 || nr.ContactMethodID != cmID {
			t.Fatalf("unexpected create notification rule: %v, err: %v", nr, err)
		}

		rules, err := c.ReadUserNotificationRules(context.Background(), uID)
		if err != nil || len(rules) != 1 {
			t.Fatalf("unexpected read rules: %v, err: %v", rules, err)
		}

		rule, err := c.ReadUserNotificationRule(context.Background(), uID, nrID)
		if err != nil || rule.ID != nrID {
			t.Fatalf("unexpected read single rule: %v, err: %v", rule, err)
		}

		if err := c.DeleteUserNotificationRule(context.Background(), nrID); err != nil {
			t.Fatalf("unexpected delete rule error: %v", err)
		}
	})
}

func TestRotationOperations(t *testing.T) {
	const rotID = "77777777-7777-4777-8777-777777777777"
	const uID1 = "11111111-1111-4111-8111-111111111111"
	const uID2 = "22222222-2222-4222-8222-222222222222"

	t.Run("rotation lifecycle", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Operation string `json:"operationName"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			switch body.Operation {
			case "ProviderCreateRotation":
				fmt.Fprint(w, `{"data":{"createRotation":{"id":"`+rotID+`","name":"Primary On-Call","description":"Daily rotation","type":"daily","start":"2026-10-01T08:00:00Z","timeZone":"Europe/Madrid","shiftLength":1,"userIDs":["`+uID1+`","`+uID2+`"],"activeUserIndex":0}}}`)
			case "ProviderReadRotation":
				fmt.Fprint(w, `{"data":{"rotation":{"id":"`+rotID+`","name":"Primary On-Call","description":"Daily rotation","type":"daily","start":"2026-10-01T08:00:00Z","timeZone":"Europe/Madrid","shiftLength":1,"userIDs":["`+uID1+`","`+uID2+`"],"activeUserIndex":0}}}`)
			case "ProviderSearchRotations":
				fmt.Fprint(w, `{"data":{"rotations":{"nodes":[{"id":"`+rotID+`","name":"Primary On-Call","description":"Daily rotation","type":"daily","start":"2026-10-01T08:00:00Z","timeZone":"Europe/Madrid","shiftLength":1,"userIDs":["`+uID1+`","`+uID2+`"],"activeUserIndex":0}]}}}`)
			case "ProviderUpdateRotation":
				fmt.Fprint(w, `{"data":{"updateRotation":true}}`)
			case "ProviderDeleteRotation":
				fmt.Fprint(w, `{"data":{"deleteAll":true}}`)
			default:
				t.Errorf("unexpected operation: %s", body.Operation)
			}
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		shiftLength := int64(1)
		desc := "Daily rotation"
		rot, err := c.CreateRotation(context.Background(), CreateRotationInput{
			Name:        "Primary On-Call",
			Description: &desc,
			Type:        "daily",
			Start:       "2026-10-01T08:00:00Z",
			TimeZone:    "Europe/Madrid",
			ShiftLength: &shiftLength,
			UserIDs:     []string{uID1, uID2},
		})
		if err != nil || rot.ID != rotID || rot.Name != "Primary On-Call" || len(rot.UserIDs) != 2 {
			t.Fatalf("unexpected create rotation result: %v, err: %v", rot, err)
		}

		readRot, err := c.ReadRotation(context.Background(), rotID)
		if err != nil || readRot.ID != rotID || readRot.ActiveUserIndex != 0 {
			t.Fatalf("unexpected read rotation: %v, err: %v", readRot, err)
		}

		searchResults, err := c.SearchRotations(context.Background(), "Primary")
		if err != nil || len(searchResults) != 1 || searchResults[0].ID != rotID {
			t.Fatalf("unexpected search rotations: %v, err: %v", searchResults, err)
		}

		nameUpdate := "Updated On-Call"
		if err := c.UpdateRotation(context.Background(), UpdateRotationInput{
			ID:   rotID,
			Name: &nameUpdate,
		}); err != nil {
			t.Fatalf("unexpected update rotation error: %v", err)
		}

		if err := c.DeleteRotation(context.Background(), rotID); err != nil {
			t.Fatalf("unexpected delete rotation error: %v", err)
		}
	})

	t.Run("rotation not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"rotation":null}}`)
		}))
		defer server.Close()

		c, _ := New(server.URL, "token", true)
		_, err := c.ReadRotation(context.Background(), rotID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestEscalationPolicyStepTargets(t *testing.T) {
	const epID = "88888888-8888-4888-8888-888888888888"
	const stepID = "99999999-9999-4999-8999-999999999999"
	const uID = "11111111-1111-4111-8111-111111111111"
	const rotID = "77777777-7777-4777-8777-777777777777"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Operation string `json:"operationName"`
			Variables struct {
				Input struct {
					Targets []TargetInput `json:"targets"`
				} `json:"input"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		switch body.Operation {
		case "ProviderCreateEscalationPolicyStep":
			if len(body.Variables.Input.Targets) != 2 {
				t.Errorf("expected 2 targets, got %d", len(body.Variables.Input.Targets))
			}
			fmt.Fprint(w, `{"data":{"createEscalationPolicyStep":{"id":"`+stepID+`","stepNumber":0,"delayMinutes":15,"actions":[{"type":"builtin-user","args":{"user_id":"`+uID+`"}},{"type":"builtin-rotation","args":{"rotation_id":"`+rotID+`"}}]}}}`)
		case "ProviderUpdateEscalationPolicyStep":
			if len(body.Variables.Input.Targets) != 2 {
				t.Errorf("expected 2 targets, got %d", len(body.Variables.Input.Targets))
			}
			fmt.Fprint(w, `{"data":{"updateEscalationPolicyStep":true}}`)
		default:
			t.Errorf("unexpected operation: %s", body.Operation)
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", true)
	epIDVal := epID
	step, err := c.CreateEscalationPolicyStep(context.Background(), CreateEscalationPolicyStepInput{
		EscalationPolicyID: &epIDVal,
		DelayMinutes:       15,
		Targets: []TargetInput{
			{ID: uID, Type: "user"},
			{ID: rotID, Type: "rotation"},
		},
	})
	if err != nil || step.ID != stepID || len(step.Actions) != 2 {
		t.Fatalf("unexpected create step result: %v, err: %v", step, err)
	}

	delay := int64(20)
	err = c.UpdateEscalationPolicyStep(context.Background(), UpdateEscalationPolicyStepInput{
		ID:           stepID,
		DelayMinutes: &delay,
		Targets: []TargetInput{
			{ID: uID, Type: "user"},
			{ID: rotID, Type: "rotation"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected update step err: %v", err)
	}
}

func TestPolymorphicLabels(t *testing.T) {
	const targetID = "11111111-1111-4111-8111-111111111111"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Operation string          `json:"operationName"`
			Variables json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		switch body.Operation {
		case "ProviderSetServiceLabel":
			fmt.Fprint(w, `{"data":{"setLabel":true}}`)
		case "ProviderReadEscalationPolicyLabels":
			fmt.Fprint(w, `{"data":{"escalationPolicy":{"id":"`+targetID+`","labels":[{"key":"example.com/team","value":"sre"}]}}}`)
		case "ProviderReadScheduleLabels":
			fmt.Fprint(w, `{"data":{"schedule":{"id":"`+targetID+`","labels":[{"key":"example.com/tier","value":"tier-1"}]}}}`)
		case "ProviderReadRotationLabels":
			fmt.Fprint(w, `{"data":{"rotation":{"id":"`+targetID+`","labels":[{"key":"example.com/region","value":"eu-west"}]}}}`)
		default:
			t.Errorf("unexpected operation: %s", body.Operation)
		}
	}))
	defer server.Close()

	c, _ := New(server.URL, "token", true)
	ctx := context.Background()

	for _, tt := range []struct {
		targetType string
		key        string
		val        string
	}{
		{"escalation_policy", "example.com/team", "sre"},
		{"schedule", "example.com/tier", "tier-1"},
		{"rotation", "example.com/region", "eu-west"},
	} {
		if err := c.SetLabel(ctx, tt.targetType, targetID, tt.key, tt.val); err != nil {
			t.Fatalf("SetLabel(%s) failed: %v", tt.targetType, err)
		}
		labels, err := c.ReadTargetLabels(ctx, tt.targetType, targetID)
		if err != nil {
			t.Fatalf("ReadTargetLabels(%s) failed: %v", tt.targetType, err)
		}
		if labels[tt.key] != tt.val {
			t.Fatalf("expected %s=%s, got %v", tt.key, tt.val, labels)
		}
	}
}
