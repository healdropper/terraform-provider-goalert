package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUserResourceMetadata(t *testing.T) {
	r := NewUserResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_user" {
		t.Fatalf("expected goalert_user, got: %s", resp.TypeName)
	}
}

func TestUserResourceSchema(t *testing.T) {
	r := NewUserResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{"id", "name", "email", "role", "username", "password"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_user schema", attr)
		}
	}
}

func TestParseUserImportID(t *testing.T) {
	const validUUID = "11111111-1111-4111-8111-111111111111"

	t.Run("compound format with username", func(t *testing.T) {
		u, username, err := parseUserImportID(validUUID + "/alice")
		if err != nil || u != validUUID || username != "alice" {
			t.Fatalf("unexpected result: u=%s, username=%s, err=%v", u, username, err)
		}
	})

	t.Run("standalone UUID", func(t *testing.T) {
		u, username, err := parseUserImportID(validUUID)
		if err != nil || u != validUUID || username != "" {
			t.Fatalf("unexpected result: u=%s, username=%s, err=%v", u, username, err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		_, _, err := parseUserImportID("invalid")
		if err == nil {
			t.Fatal("expected error on invalid format")
		}
	})
}
