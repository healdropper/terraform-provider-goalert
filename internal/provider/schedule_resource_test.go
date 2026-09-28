package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

func TestScheduleResourceMetadata(t *testing.T) {
	r := NewScheduleResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_schedule" {
		t.Fatalf("expected goalert_schedule, got: %s", resp.TypeName)
	}
}

func TestScheduleResourceSchema(t *testing.T) {
	r := NewScheduleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"id", "name", "description", "time_zone",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_schedule schema", attr)
		}
	}
}

func TestScheduleDataSourceMetadata(t *testing.T) {
	d := NewScheduleDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_schedule" {
		t.Fatalf("expected goalert_schedule, got: %s", resp.TypeName)
	}
}

func TestScheduleDataSourceSchema(t *testing.T) {
	d := NewScheduleDataSource()
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"id", "name", "description", "time_zone",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in data.goalert_schedule schema", attr)
		}
	}
}

func TestScheduleNamePattern(t *testing.T) {
	validNames := []string{
		"Primary Schedule",
		"tier-1_support",
		"SRE's Ops Schedule",
		"Schedule123",
	}
	for _, name := range validNames {
		if !scheduleNamePattern.MatchString(name) {
			t.Errorf("expected %q to be valid name", name)
		}
	}

	invalidNames := []string{
		"Invalid@Name",
		"Schedule/1",
		"Schedule:Main",
	}
	for _, name := range invalidNames {
		if scheduleNamePattern.MatchString(name) {
			t.Errorf("expected %q to be invalid name", name)
		}
	}
}

func TestModelFromSchedule(t *testing.T) {
	sched := &client.Schedule{
		ID:          "sched-1",
		Name:        "Engineering On-Call",
		Description: "Engineering coverage",
		TimeZone:    "Europe/Madrid",
	}

	m := modelFromSchedule(sched)
	if m.ID.ValueString() != "sched-1" {
		t.Errorf("expected ID sched-1, got: %s", m.ID.ValueString())
	}
	if m.Name.ValueString() != "Engineering On-Call" {
		t.Errorf("expected Name Engineering On-Call, got: %s", m.Name.ValueString())
	}
	if m.TimeZone.ValueString() != "Europe/Madrid" {
		t.Errorf("expected TimeZone Europe/Madrid, got: %s", m.TimeZone.ValueString())
	}
}
