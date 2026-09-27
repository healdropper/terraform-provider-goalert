package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestDataSourcesMetadataAndSchema(t *testing.T) {
	ctx := context.Background()

	t.Run("service data source", func(t *testing.T) {
		d := NewServiceDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_service" {
			t.Fatalf("expected goalert_service, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "description", "escalation_policy_id"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})

	t.Run("escalation policy data source", func(t *testing.T) {
		d := NewEscalationPolicyDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_escalation_policy" {
			t.Fatalf("expected goalert_escalation_policy, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "description", "repeat"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})

	t.Run("integration key data source", func(t *testing.T) {
		d := NewIntegrationKeyDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_integration_key" {
			t.Fatalf("expected goalert_integration_key, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "service_id", "name", "type", "href"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})

	t.Run("heartbeat monitor data source", func(t *testing.T) {
		d := NewHeartbeatMonitorDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_heartbeat_monitor" {
			t.Fatalf("expected goalert_heartbeat_monitor, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "service_id", "name", "timeout_minutes", "href"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})
}
