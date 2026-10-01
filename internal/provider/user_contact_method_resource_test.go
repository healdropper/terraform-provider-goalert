package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUserContactMethodResourceMetadata(t *testing.T) {
	r := NewUserContactMethodResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_user_contact_method" {
		t.Fatalf("expected goalert_user_contact_method, got: %s", resp.TypeName)
	}
}

func TestUserContactMethodResourceSchema(t *testing.T) {
	r := NewUserContactMethodResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{"id", "user_id", "name", "type", "value", "enable_status_updates", "private", "status_updates", "disabled"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_user_contact_method schema", attr)
		}
	}
}

func TestParseUserContactMethodImportID(t *testing.T) {
	const validUser = "11111111-1111-4111-8111-111111111111"
	const validCM = "22222222-2222-4222-8222-222222222222"

	t.Run("compound format", func(t *testing.T) {
		u, cm, err := parseUserContactMethodImportID(validUser + "/" + validCM)
		if err != nil || u != validUser || cm != validCM {
			t.Fatalf("unexpected result: u=%s, cm=%s, err=%v", u, cm, err)
		}
	})

	t.Run("standalone UUID", func(t *testing.T) {
		u, cm, err := parseUserContactMethodImportID(validCM)
		if err != nil || u != "" || cm != validCM {
			t.Fatalf("unexpected result: u=%s, cm=%s, err=%v", u, cm, err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, _, err := parseUserContactMethodImportID("invalid")
		if err == nil {
			t.Fatal("expected error on invalid format")
		}
	})
}
