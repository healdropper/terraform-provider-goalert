package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestScheduleRuleResourceMetadata(t *testing.T) {
	r := NewScheduleRuleResource()
	req := resource.MetadataRequest{
		ProviderTypeName: "goalert",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "goalert_schedule_rule" {
		t.Fatalf("expected goalert_schedule_rule, got: %s", resp.TypeName)
	}
}

func TestScheduleRuleResourceSchema(t *testing.T) {
	r := NewScheduleRuleResource()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"id", "schedule_id", "target_type", "target_id", "start_time", "end_time", "weekday_filter",
	}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing expected attribute %q in goalert_schedule_rule schema", attr)
		}
	}
}

func TestClockTimePattern(t *testing.T) {
	validTimes := []string{"00:00", "09:30", "17:00", "23:59"}
	for _, vt := range validTimes {
		if !clockTimePattern.MatchString(vt) {
			t.Errorf("expected %q to be valid clock time", vt)
		}
	}

	invalidTimes := []string{"9:00", "24:00", "12:60", "ab:cd"}
	for _, it := range invalidTimes {
		if clockTimePattern.MatchString(it) {
			t.Errorf("expected %q to be invalid clock time", it)
		}
	}
}

func TestWeekdayBoolsConversions(t *testing.T) {
	bools := []bool{true, true, true, true, true, false, false}
	typesBools := sliceToWeekdayBools(bools)
	if len(typesBools) != 7 {
		t.Fatalf("expected 7 elements, got: %d", len(typesBools))
	}
	for i, tb := range typesBools {
		if tb.ValueBool() != bools[i] {
			t.Errorf("mismatch at index %d: got %v, want %v", i, tb.ValueBool(), bools[i])
		}
	}

	roundtrip := weekdayBoolsToSlice(typesBools)
	for i, b := range roundtrip {
		if b != bools[i] {
			t.Errorf("roundtrip mismatch at index %d: got %v, want %v", i, b, bools[i])
		}
	}
}

func TestDefaultWeekdayFilter(t *testing.T) {
	df := defaultWeekdayFilter()
	if len(df) != 7 {
		t.Fatalf("expected 7 days, got: %d", len(df))
	}
	for i, d := range df {
		if !d.ValueBool() {
			t.Errorf("expected day %d to default to true", i)
		}
	}
}
