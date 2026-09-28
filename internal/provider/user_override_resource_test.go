package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUserOverrideResourceMetadata(t *testing.T) {
	r := NewUserOverrideResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_user_override" {
		t.Fatalf("expected goalert_user_override, got: %s", resp.TypeName)
	}
}

func TestUserOverrideResourceSchema(t *testing.T) {
	r := NewUserOverrideResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"id", "schedule_id", "start_time", "end_time", "add_user_id", "remove_user_id",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_user_override schema", attr)
		}
	}
}
