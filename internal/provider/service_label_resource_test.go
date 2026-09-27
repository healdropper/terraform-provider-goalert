package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestServiceLabelResourceMetadata(t *testing.T) {
	r := NewServiceLabelResource()
	req := resource.MetadataRequest{ProviderTypeName: "goalert"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "goalert_service_label" {
		t.Fatalf("expected type name goalert_service_label, got %s", resp.TypeName)
	}
}

func TestServiceLabelResourceSchema(t *testing.T) {
	r := NewServiceLabelResource()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, expected := range []string{"id", "service_id", "key", "value"} {
		if _, ok := attrs[expected]; !ok {
			t.Fatalf("missing required schema attribute: %s", expected)
		}
	}
}
