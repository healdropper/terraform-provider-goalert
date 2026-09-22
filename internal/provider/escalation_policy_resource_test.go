package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

func TestEscalationPolicyResourceMetadata(t *testing.T) {
	r := NewEscalationPolicyResource()
	req := resource.MetadataRequest{ProviderTypeName: "goalert"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "goalert_escalation_policy" {
		t.Fatalf("expected type name goalert_escalation_policy, got %s", resp.TypeName)
	}
}

func TestEscalationPolicyResourceSchema(t *testing.T) {
	r := NewEscalationPolicyResource()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, expected := range []string{"id", "name", "description", "repeat"} {
		if _, ok := attrs[expected]; !ok {
			t.Fatalf("missing required schema attribute: %s", expected)
		}
	}
	blocks := resp.Schema.Blocks
	if _, ok := blocks["step"]; !ok {
		t.Fatalf("missing required schema block: step")
	}
}

func TestModelFromEscalationPolicy_SortingAndActions(t *testing.T) {
	ep := &client.EscalationPolicy{
		ID:          "policy-123",
		Name:        "Critical Escalation",
		Description: "Production alerts",
		Repeat:      3,
		Steps: []client.EscalationPolicyStep{
			{
				ID:           "step-2",
				StepNumber:   1,
				DelayMinutes: 10,
				Actions: []client.Destination{
					{
						Type: "builtin-webhook",
						Args: map[string]string{"webhook_url": "https://alerts.example.com/step2"},
					},
				},
			},
			{
				ID:           "step-1",
				StepNumber:   0,
				DelayMinutes: 5,
				Actions: []client.Destination{
					{
						Type: "builtin-webhook",
						Args: map[string]string{"webhook_url": "https://alerts.example.com/step1"},
					},
					{
						Type: "other-type",
						Args: map[string]string{"arg": "val"},
					},
				},
			},
		},
	}

	model := modelFromEscalationPolicy(ep, nil)

	if model.ID.ValueString() != "policy-123" {
		t.Errorf("expected ID policy-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "Critical Escalation" {
		t.Errorf("expected Name Critical Escalation, got %s", model.Name.ValueString())
	}
	if model.Repeat.ValueInt64() != 3 {
		t.Errorf("expected Repeat 3, got %d", model.Repeat.ValueInt64())
	}
	if len(model.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(model.Steps))
	}

	// Must be sorted by StepNumber (0 then 1)
	if model.Steps[0].ID.ValueString() != "step-1" || model.Steps[0].StepNumber.ValueInt64() != 0 {
		t.Errorf("expected first step step-1 at 0, got %s at %d", model.Steps[0].ID.ValueString(), model.Steps[0].StepNumber.ValueInt64())
	}
	if model.Steps[1].ID.ValueString() != "step-2" || model.Steps[1].StepNumber.ValueInt64() != 1 {
		t.Errorf("expected second step step-2 at 1, got %s at %d", model.Steps[1].ID.ValueString(), model.Steps[1].StepNumber.ValueInt64())
	}

	// Step 1 only extracts builtin-webhook (ignoring other-type)
	if len(model.Steps[0].WebhookActions) != 1 {
		t.Fatalf("expected 1 webhook action in step 0, got %d", len(model.Steps[0].WebhookActions))
	}
	if model.Steps[0].WebhookActions[0].URL.ValueString() != "https://alerts.example.com/step1" {
		t.Errorf("unexpected webhook action URL: %s", model.Steps[0].WebhookActions[0].URL.ValueString())
	}
}

func TestActionsMatch(t *testing.T) {
	serverActions := []client.Destination{
		{Type: "builtin-webhook", Args: map[string]string{"webhook_url": "https://example.com/hook1"}},
		{Type: "builtin-webhook", Args: map[string]string{"webhook_url": "https://example.com/hook2"}},
	}

	matchingPlan := []webhookActionModel{
		{URL: types.StringValue("https://example.com/hook1")},
		{URL: types.StringValue("https://example.com/hook2")},
	}

	mismatchedPlan := []webhookActionModel{
		{URL: types.StringValue("https://example.com/hook1")},
		{URL: types.StringValue("https://example.com/hook-different")},
	}

	diffCountPlan := []webhookActionModel{
		{URL: types.StringValue("https://example.com/hook1")},
	}

	if !actionsMatch(serverActions, matchingPlan) {
		t.Error("expected matchingPlan to match")
	}
	if actionsMatch(serverActions, mismatchedPlan) {
		t.Error("expected mismatchedPlan to not match")
	}
	if actionsMatch(serverActions, diffCountPlan) {
		t.Error("expected diffCountPlan to not match")
	}
}
