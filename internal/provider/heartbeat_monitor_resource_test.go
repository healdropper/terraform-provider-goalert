package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

func TestHeartbeatMonitorResourceMetadata(t *testing.T) {
	r := NewHeartbeatMonitorResource()
	req := resource.MetadataRequest{ProviderTypeName: "goalert"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "goalert_heartbeat_monitor" {
		t.Fatalf("expected type name goalert_heartbeat_monitor, got %s", resp.TypeName)
	}
}

func TestHeartbeatMonitorResourceSchema(t *testing.T) {
	r := NewHeartbeatMonitorResource()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, expected := range []string{"id", "service_id", "name", "timeout_minutes", "href"} {
		if _, ok := attrs[expected]; !ok {
			t.Fatalf("missing required schema attribute: %s", expected)
		}
	}

	hrefAttr, ok := attrs["href"].(interface{ IsSensitive() bool })
	if !ok || !hrefAttr.IsSensitive() {
		t.Fatal("expected href attribute to be sensitive")
	}
}

func TestModelFromHeartbeatMonitor(t *testing.T) {
	hb := &client.HeartbeatMonitor{
		ID:             "33333333-3333-4333-8333-333333333333",
		ServiceID:      "22222222-2222-4222-8222-222222222222",
		Name:           "Backup Monitor",
		TimeoutMinutes: 15,
		Href:           "http://example.com/api/v2/heartbeat/33333333-3333-4333-8333-333333333333",
	}

	model := modelFromHeartbeatMonitor(hb)
	if model.ID.ValueString() != hb.ID ||
		model.ServiceID.ValueString() != hb.ServiceID ||
		model.Name.ValueString() != hb.Name ||
		model.TimeoutMinutes.ValueInt64() != hb.TimeoutMinutes ||
		model.Href.ValueString() != hb.Href {
		t.Fatalf("model mismatch: %+v", model)
	}
}
