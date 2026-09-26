package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

func TestIntegrationKeyResourceMetadata(t *testing.T) {
	r := NewIntegrationKeyResource()
	req := resource.MetadataRequest{ProviderTypeName: "goalert"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "goalert_integration_key" {
		t.Fatalf("expected type name goalert_integration_key, got %s", resp.TypeName)
	}
}

func TestIntegrationKeyResourceSchema(t *testing.T) {
	r := NewIntegrationKeyResource()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, expected := range []string{"id", "service_id", "name", "type", "href"} {
		if _, ok := attrs[expected]; !ok {
			t.Fatalf("missing required schema attribute: %s", expected)
		}
	}

	hrefAttr, ok := attrs["href"].(interface{ IsSensitive() bool })
	if !ok || !hrefAttr.IsSensitive() {
		t.Fatal("expected href attribute to be sensitive")
	}
}

func TestModelFromIntegrationKey(t *testing.T) {
	ik := &client.IntegrationKey{
		ID:        "11111111-1111-4111-8111-111111111111",
		ServiceID: "22222222-2222-4222-8222-222222222222",
		Name:      "Grafana Key",
		Type:      "grafana",
		Href:      "http://example.com/api/v2/grafana/incoming?token=11111111-1111-4111-8111-111111111111",
	}

	model := modelFromIntegrationKey(ik)
	if model.ID.ValueString() != ik.ID ||
		model.ServiceID.ValueString() != ik.ServiceID ||
		model.Name.ValueString() != ik.Name ||
		model.Type.ValueString() != ik.Type ||
		model.Href.ValueString() != ik.Href {
		t.Fatalf("model mismatch: %+v", model)
	}
}

func TestIntegrationKeyImportState_Formats(t *testing.T) {
	cases := []struct {
		name       string
		id         string
		expectedID string
		shouldFail bool
	}{
		{
			name:       "bare uuid",
			id:         "11111111-1111-4111-8111-111111111111",
			expectedID: "11111111-1111-4111-8111-111111111111",
		},
		{
			name:       "compound service/key uuid",
			id:         "22222222-2222-4222-8222-222222222222/11111111-1111-4111-8111-111111111111",
			expectedID: "11111111-1111-4111-8111-111111111111",
		},
		{
			name:       "invalid uuid",
			id:         "not-a-uuid",
			shouldFail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parseImportID(tc.id)
			if tc.shouldFail {
				if err == nil {
					t.Fatal("expected error for invalid import ID, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected import error: %v", err)
			}

			if parsed != tc.expectedID {
				t.Fatalf("expected imported ID %s, got %s", tc.expectedID, parsed)
			}
		})
	}
}
