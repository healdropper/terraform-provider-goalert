package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

func TestRotationResourceMetadata(t *testing.T) {
	r := NewRotationResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_rotation" {
		t.Fatalf("expected goalert_rotation, got: %s", resp.TypeName)
	}
}

func TestRotationResourceSchema(t *testing.T) {
	r := NewRotationResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"id", "name", "description", "type", "start_time",
		"time_zone", "shift_length", "user_ids", "active_user_index",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_rotation schema", attr)
		}
	}
}

func TestRotationDataSourceSchema(t *testing.T) {
	d := NewRotationDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"id", "name", "description", "type", "start_time",
		"time_zone", "shift_length", "user_ids", "active_user_index",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in data.goalert_rotation schema", attr)
		}
	}
}

func TestRotationNamePattern(t *testing.T) {
	validNames := []string{
		"Primary On-Call",
		"tier-1_support",
		"Alice's Rotation",
		"SRE-Rotation-24x7",
	}
	for _, n := range validNames {
		if !rotationNamePattern.MatchString(n) {
			t.Errorf("expected valid name %q to match pattern", n)
		}
	}

	invalidNames := []string{
		"Rotation (EU)",
		"Primary@OnCall",
		"Tier#1",
		"Rotation/Daily",
	}
	for _, n := range invalidNames {
		if rotationNamePattern.MatchString(n) {
			t.Errorf("expected invalid name %q to not match pattern", n)
		}
	}
}

func TestModelFromRotation(t *testing.T) {
	rot := &client.Rotation{
		ID:              "rot-123",
		Name:            "SRE Rotation",
		Description:     "Daily handoffs",
		Type:            "daily",
		Start:           "2026-10-01T08:00:00Z",
		TimeZone:        "Europe/Madrid",
		ShiftLength:     2,
		UserIDs:         []string{"user-1", "user-2"},
		ActiveUserIndex: 1,
	}

	m := modelFromRotation(rot, nil)
	if m.ID.ValueString() != "rot-123" {
		t.Errorf("expected ID rot-123, got: %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "SRE Rotation" {
		t.Errorf("expected Name SRE Rotation, got: %s", m.Name.ValueString())
	}
	if m.ShiftLength.ValueInt64() != 2 {
		t.Errorf("expected ShiftLength 2, got: %d", m.ShiftLength.ValueInt64())
	}
	if len(m.UserIDs) != 2 || m.UserIDs[0].ValueString() != "user-1" {
		t.Errorf("unexpected UserIDs: %v", m.UserIDs)
	}
	if m.ActiveUserIndex.ValueInt64() != 1 {
		t.Errorf("expected ActiveUserIndex 1, got: %d", m.ActiveUserIndex.ValueInt64())
	}
}
