package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func TestStepMatchesWithTargets(t *testing.T) {
	serverStep := client.EscalationPolicyStep{
		ID:           "step-1",
		DelayMinutes: 15,
		MultiAck:     true,
		Actions: []client.Destination{
			{Type: "builtin-user", Args: map[string]string{"user_id": "u1"}},
			{Type: "builtin-rotation", Args: map[string]string{"rotation_id": "r1"}},
			{Type: "builtin-schedule", Args: map[string]string{"schedule_id": "s1"}},
			{Type: "builtin-webhook", Args: map[string]string{"webhook_url": "https://example.com/alert"}},
		},
	}

	matchingPlan := stepModel{
		DelayMinutes: types.Int64Value(15),
		MultiAck:     types.BoolValue(true),
		UserIDs:      []types.String{types.StringValue("u1")},
		RotationIDs:  []types.String{types.StringValue("r1")},
		ScheduleIDs:  []types.String{types.StringValue("s1")},
		WebhookActions: []webhookActionModel{
			{URL: types.StringValue("https://example.com/alert")},
		},
	}

	diffDelayPlan := matchingPlan
	diffDelayPlan.DelayMinutes = types.Int64Value(20)

	diffMultiAckPlan := matchingPlan
	diffMultiAckPlan.MultiAck = types.BoolValue(false)

	diffUserPlan := matchingPlan
	diffUserPlan.UserIDs = []types.String{types.StringValue("u2")}

	diffRotationPlan := matchingPlan
	diffRotationPlan.RotationIDs = []types.String{types.StringValue("r2")}

	diffSchedulePlan := matchingPlan
	diffSchedulePlan.ScheduleIDs = []types.String{types.StringValue("s2")}

	if !stepMatches(serverStep, matchingPlan) {
		t.Error("expected matchingPlan to match step")
	}
	if stepMatches(serverStep, diffDelayPlan) {
		t.Error("expected diffDelayPlan to not match")
	}
	if stepMatches(serverStep, diffMultiAckPlan) {
		t.Error("expected diffMultiAckPlan to not match")
	}
	if stepMatches(serverStep, diffUserPlan) {
		t.Error("expected diffUserPlan to not match")
	}
	if stepMatches(serverStep, diffRotationPlan) {
		t.Error("expected diffRotationPlan to not match")
	}
	if stepMatches(serverStep, diffSchedulePlan) {
		t.Error("expected diffSchedulePlan to not match")
	}
}

func TestModelFromEscalationPolicyTargets(t *testing.T) {
	ep := &client.EscalationPolicy{
		ID:   "ep-1",
		Name: "Tier 1",
		Steps: []client.EscalationPolicyStep{
			{
				ID:           "s-1",
				StepNumber:   0,
				DelayMinutes: 15,
				MultiAck:     true,
				Actions: []client.Destination{
					{Type: "builtin-user", Args: map[string]string{"user_id": "user-uuid-1"}},
					{Type: "builtin-rotation", Args: map[string]string{"rotation_id": "rotation-uuid-1"}},
					{Type: "builtin-schedule", Args: map[string]string{"schedule_id": "schedule-uuid-1"}},
					{Type: "builtin-webhook", Args: map[string]string{"webhook_url": "https://example.com"}},
				},
			},
		},
	}

	m := modelFromEscalationPolicy(ep, nil)
	if len(m.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(m.Steps))
	}
	s := m.Steps[0]
	if !s.MultiAck.ValueBool() {
		t.Errorf("expected MultiAck true, got false")
	}
	if len(s.UserIDs) != 1 || s.UserIDs[0].ValueString() != "user-uuid-1" {
		t.Errorf("expected user-uuid-1, got: %v", s.UserIDs)
	}
	if len(s.RotationIDs) != 1 || s.RotationIDs[0].ValueString() != "rotation-uuid-1" {
		t.Errorf("expected rotation-uuid-1, got: %v", s.RotationIDs)
	}
	if len(s.ScheduleIDs) != 1 || s.ScheduleIDs[0].ValueString() != "schedule-uuid-1" {
		t.Errorf("expected schedule-uuid-1, got: %v", s.ScheduleIDs)
	}
	if len(s.WebhookActions) != 1 || s.WebhookActions[0].URL.ValueString() != "https://example.com" {
		t.Errorf("expected webhook action, got: %v", s.WebhookActions)
	}
}

func TestEscalationPolicyStepSchemaSlackChannelIDs(t *testing.T) {
	r := NewEscalationPolicyResource()
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	step, ok := resp.Schema.Blocks["step"].(schema.ListNestedBlock)
	if !ok {
		t.Fatalf("step block has unexpected type %T", resp.Schema.Blocks["step"])
	}
	if _, ok := step.NestedObject.Attributes["slack_channel_ids"]; !ok {
		t.Fatal("missing step attribute: slack_channel_ids")
	}
}

func TestModelFromEscalationPolicySlackChannels(t *testing.T) {
	ep := &client.EscalationPolicy{
		ID:   "ep-1",
		Name: "Slack first",
		Steps: []client.EscalationPolicyStep{
			{
				ID:           "s-1",
				StepNumber:   0,
				DelayMinutes: 5,
				Actions: []client.Destination{
					{Type: "builtin-slack-channel", Args: map[string]string{"slack_channel_id": "C0123ABCD"}},
					{Type: "builtin-user", Args: map[string]string{"user_id": "user-uuid-1"}},
				},
			},
		},
	}

	s := modelFromEscalationPolicy(ep, nil).Steps[0]
	if len(s.SlackChannelIDs) != 1 || s.SlackChannelIDs[0].ValueString() != "C0123ABCD" {
		t.Errorf("expected slack channel C0123ABCD, got: %v", s.SlackChannelIDs)
	}
	if len(s.UserIDs) != 1 {
		t.Errorf("expected the user target to stay, got: %v", s.UserIDs)
	}
}

func TestModelFromEscalationPolicyEmptySlackChannelsKeepPriorEmptyList(t *testing.T) {
	ep := &client.EscalationPolicy{
		ID: "ep-1",
		Steps: []client.EscalationPolicyStep{
			{ID: "s-1", DelayMinutes: 5, Actions: []client.Destination{
				{Type: "builtin-user", Args: map[string]string{"user_id": "u1"}},
			}},
		},
	}
	prior := &escalationPolicyModel{Steps: []stepModel{{SlackChannelIDs: []types.String{}}}}

	s := modelFromEscalationPolicy(ep, prior).Steps[0]
	if s.SlackChannelIDs == nil || len(s.SlackChannelIDs) != 0 {
		t.Errorf("expected empty non-nil list, got: %#v", s.SlackChannelIDs)
	}
}

func TestStepMatchesSlackChannels(t *testing.T) {
	serverStep := client.EscalationPolicyStep{
		DelayMinutes: 5,
		Actions: []client.Destination{
			{Type: "builtin-slack-channel", Args: map[string]string{"slack_channel_id": "C1"}},
		},
	}
	plan := stepModel{
		DelayMinutes:    types.Int64Value(5),
		MultiAck:        types.BoolValue(false),
		SlackChannelIDs: []types.String{types.StringValue("C1")},
	}
	if !stepMatches(serverStep, plan) {
		t.Error("expected matching slack channel to match")
	}

	plan.SlackChannelIDs = []types.String{types.StringValue("C2")}
	if stepMatches(serverStep, plan) {
		t.Error("expected different slack channel to not match")
	}

	plan.SlackChannelIDs = nil
	if stepMatches(serverStep, plan) {
		t.Error("expected removed slack channel to not match")
	}
}

func TestStepActionsIncludesSlackChannels(t *testing.T) {
	actions := stepActions(stepModel{
		UserIDs:         []types.String{types.StringValue("u1")},
		SlackChannelIDs: []types.String{types.StringValue("C1"), types.StringValue("C2")},
	})

	var got []string
	for _, a := range actions {
		if a.Type == "builtin-slack-channel" {
			got = append(got, a.Args["slack_channel_id"])
		}
	}
	if len(got) != 2 || got[0] != "C1" || got[1] != "C2" {
		t.Errorf("expected slack actions C1, C2 in order, got: %v", got)
	}
	if len(actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(actions))
	}
}

func TestStepHasTargetsSlackOnly(t *testing.T) {
	if !stepHasTargets(stepModel{SlackChannelIDs: []types.String{types.StringValue("C1")}}) {
		t.Error("expected a step with only a slack channel to be valid")
	}
	if stepHasTargets(stepModel{}) {
		t.Error("expected a step without targets to be invalid")
	}
}
