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

	t.Run("user data source", func(t *testing.T) {
		d := NewUserDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_user" {
			t.Fatalf("expected goalert_user, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "email", "role"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})
	t.Run("rotation data source", func(t *testing.T) {
		d := NewRotationDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_rotation" {
			t.Fatalf("expected goalert_rotation, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "description", "type", "start_time", "time_zone", "shift_length", "user_ids", "active_user_index"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})

	t.Run("schedule data source", func(t *testing.T) {
		d := NewScheduleDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_schedule" {
			t.Fatalf("expected goalert_schedule, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "description", "time_zone"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})

	t.Run("slack channel data source", func(t *testing.T) {
		d := NewSlackChannelDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_slack_channel" {
			t.Fatalf("expected goalert_slack_channel, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "team_id"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})

	t.Run("slack user group data source", func(t *testing.T) {
		d := NewSlackUserGroupDataSource()
		mReq := datasource.MetadataRequest{ProviderTypeName: "goalert"}
		mResp := datasource.MetadataResponse{}
		d.Metadata(ctx, mReq, &mResp)
		if mResp.TypeName != "goalert_slack_user_group" {
			t.Fatalf("expected goalert_slack_user_group, got %s", mResp.TypeName)
		}

		sReq := datasource.SchemaRequest{}
		sResp := datasource.SchemaResponse{}
		d.Schema(ctx, sReq, &sResp)
		if sResp.Diagnostics.HasError() {
			t.Fatalf("unexpected schema diagnostics: %v", sResp.Diagnostics)
		}
		for _, attr := range []string{"id", "name", "handle"} {
			if _, ok := sResp.Schema.Attributes[attr]; !ok {
				t.Fatalf("missing attribute: %s", attr)
			}
		}
	})
}
