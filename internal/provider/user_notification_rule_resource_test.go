package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUserNotificationRuleResourceMetadata(t *testing.T) {
	r := NewUserNotificationRuleResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_user_notification_rule" {
		t.Fatalf("expected goalert_user_notification_rule, got: %s", resp.TypeName)
	}
}

func TestUserNotificationRuleResourceSchema(t *testing.T) {
	r := NewUserNotificationRuleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{"id", "user_id", "contact_method_id", "delay_minutes"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_user_notification_rule schema", attr)
		}
	}
}

func TestParseUserNotificationRuleImportID(t *testing.T) {
	const validUser = "11111111-1111-4111-8111-111111111111"
	const validRule = "33333333-3333-4333-8333-333333333333"

	t.Run("compound format", func(t *testing.T) {
		u, r, err := parseUserNotificationRuleImportID(validUser + "/" + validRule)
		if err != nil || u != validUser || r != validRule {
			t.Fatalf("unexpected result: u=%s, r=%s, err=%v", u, r, err)
		}
	})

	t.Run("standalone UUID rejected", func(t *testing.T) {
		_, _, err := parseUserNotificationRuleImportID(validRule)
		if err == nil {
			t.Fatal("expected standalone UUID to be rejected for notification rule import")
		}
	})
}
