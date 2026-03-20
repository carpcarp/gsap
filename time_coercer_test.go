package sap

import (
	"testing"
	"time"
)

type TestEvent struct {
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
}

type TestEventWithPointer struct {
	Name      string     `json:"name"`
	StartTime *time.Time `json:"start_time,omitempty"`
}

func TestParseTimeRFC3339(t *testing.T) {
	input := `{"name": "Meeting", "start_time": "2025-06-15T14:30:00Z"}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if event.Name != "Meeting" {
		t.Errorf("Expected name 'Meeting', got '%s'", event.Name)
	}

	expected := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimeRFC3339WithOffset(t *testing.T) {
	input := `{"name": "Call", "start_time": "2025-06-15T14:30:00+05:00"}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	loc := time.FixedZone("", 5*3600)
	expected := time.Date(2025, 6, 15, 14, 30, 0, 0, loc)
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimeDateOnly(t *testing.T) {
	input := `{"name": "Birthday", "start_time": "2025-03-20"}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimeUnixTimestamp(t *testing.T) {
	input := `{"name": "Epoch Event", "start_time": 1718451000}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := time.Unix(1718451000, 0).UTC()
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimeUnixMilliseconds(t *testing.T) {
	input := `{"name": "JS Event", "start_time": 1718451000000}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := time.Unix(1718451000, 0).UTC()
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimePointerWithNull(t *testing.T) {
	input := `{"name": "No Date", "start_time": null}`

	event, err := Parse[TestEventWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if event.StartTime != nil {
		t.Errorf("Expected start_time to be nil, got %v", event.StartTime)
	}
}

func TestParseTimePointerWithValue(t *testing.T) {
	input := `{"name": "Dated", "start_time": "2025-12-25T00:00:00Z"}`

	event, err := Parse[TestEventWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if event.StartTime == nil {
		t.Fatalf("Expected start_time to be non-nil")
	}

	expected := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, *event.StartTime)
	}
}

func TestParseTimeDatetimeWithoutTimezone(t *testing.T) {
	input := `{"name": "Local", "start_time": "2025-06-15T14:30:00"}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimeDatetimeWithSpace(t *testing.T) {
	input := `{"name": "Spaced", "start_time": "2025-06-15 14:30:00"}`

	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)
	if !event.StartTime.Equal(expected) {
		t.Errorf("Expected start_time %v, got %v", expected, event.StartTime)
	}
}

func TestParseTimeInvalidString(t *testing.T) {
	input := `{"name": "Bad", "start_time": "not a date"}`

	// Should still parse, but start_time will be zero since coercion fails and is skipped
	event, err := Parse[TestEvent](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if event.Name != "Bad" {
		t.Errorf("Expected name 'Bad', got '%s'", event.Name)
	}

	if !event.StartTime.IsZero() {
		t.Errorf("Expected start_time to be zero, got %v", event.StartTime)
	}
}

func TestParseTimeWithScore(t *testing.T) {
	input := `{"name": "Scored", "start_time": "2025-06-15T14:30:00Z"}`

	event, score, err := ParseWithScore[TestEvent](input)
	if err != nil {
		t.Fatalf("ParseWithScore failed: %v", err)
	}

	if event.Name != "Scored" {
		t.Errorf("Expected name 'Scored', got '%s'", event.Name)
	}

	if score == nil {
		t.Fatalf("Expected score to be non-nil")
	}

	flags := score.Flags()
	if _, ok := flags[FlagStringToTime]; !ok {
		t.Errorf("Expected StringToTime flag in score, got flags: %v", flags)
	}
}

func TestParseTimeNullStringVariant(t *testing.T) {
	input := `{"name": "TBD Event", "start_time": "TBD"}`

	event, err := Parse[TestEventWithPointer](input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if event.StartTime != nil {
		t.Errorf("Expected start_time to be nil for 'TBD', got %v", event.StartTime)
	}
}
