package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestLabelResourceMetadata(t *testing.T) {
	r := NewLabelResource()
	req := resource.MetadataRequest{ProviderTypeName: "goalert"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "goalert_label" {
		t.Fatalf("expected type name goalert_label, got %s", resp.TypeName)
	}
}

func TestLabelResourceSchema(t *testing.T) {
	r := NewLabelResource()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, expected := range []string{"id", "target_type", "target_id", "key", "value"} {
		if _, ok := attrs[expected]; !ok {
			t.Fatalf("missing required schema attribute: %s", expected)
		}
	}
}

func TestParseLabelImportID(t *testing.T) {
	const validUUID = "11111111-1111-4111-8111-111111111111"

	for _, targetType := range []string{"service", "escalation_policy", "schedule", "rotation"} {
		raw := targetType + ":" + validUUID + "/example.com/team"
		tt, tid, key, err := parseLabelImportID(raw)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", raw, err)
		}
		if tt != targetType || tid != validUUID || key != "example.com/team" {
			t.Fatalf("unexpected parsed components: tt=%s tid=%s key=%s", tt, tid, key)
		}
	}

	for _, invalid := range []string{
		"invalid",
		"unknown:" + validUUID + "/example.com/team",
		"schedule:not-a-uuid/example.com/team",
		"schedule:" + validUUID,
	} {
		if _, _, _, err := parseLabelImportID(invalid); err == nil {
			t.Errorf("expected error for invalid import ID %q", invalid)
		}
	}
}
