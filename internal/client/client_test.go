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
