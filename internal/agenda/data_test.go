package agenda

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadData(t *testing.T) {
	dataPath := writeTestData(t, `{
  "$schema": "./schema/agenda.schema.json",
  "title": "School week",
  "week": "2026-W36",
  "timezone": "Europe/Prague",
  "days": ["mon", "tue"],
  "people": [{ "id": "p1", "label": "P1" }],
  "timeRange": { "start": "07:00", "end": "19:00" },
  "slots": [{ "label": "1", "start": "08:00", "end": "08:45" }],
  "events": [{
    "person": "p1",
    "day": "mon",
    "start": "08:00",
    "end": "08:45",
    "title": "Cj",
    "subtitle": "FrE",
    "note": "1C",
    "style": "muted",
    "stack": { "index": 1, "total": 2 }
  }]
}`)

	data, err := LoadData(dataPath)
	if err != nil {
		t.Fatalf("LoadData() error = %v", err)
	}
	if data.Schema != "./schema/agenda.schema.json" {
		t.Fatalf("Schema = %q", data.Schema)
	}
	if data.Title != "School week" {
		t.Fatalf("Title = %q", data.Title)
	}
	if data.Week != "2026-W36" {
		t.Fatalf("Week = %q", data.Week)
	}
	if data.Timezone != "Europe/Prague" {
		t.Fatalf("Timezone = %q", data.Timezone)
	}
	if got, want := data.Days, []Day{"mon", "tue"}; !equalDays(got, want) {
		t.Fatalf("Days = %#v, want %#v", got, want)
	}
	if data.TimeRange.Start != "07:00" || data.TimeRange.End != "19:00" {
		t.Fatalf("TimeRange = %#v", data.TimeRange)
	}
	if len(data.People) != 1 || data.People[0].ID != "p1" || data.People[0].Label != "P1" {
		t.Fatalf("People = %#v", data.People)
	}
	if len(data.Slots) != 1 || data.Slots[0].Label != "1" {
		t.Fatalf("Slots = %#v", data.Slots)
	}
	if len(data.Events) != 1 || data.Events[0].Title != "Cj" || data.Events[0].Style != "muted" {
		t.Fatalf("Events = %#v", data.Events)
	}
	if data.Events[0].Stack.Index != 1 || data.Events[0].Stack.Total != 2 {
		t.Fatalf("Stack = %#v", data.Events[0].Stack)
	}
}

func TestLoadDataRejectsUnknownFields(t *testing.T) {
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "people": [{ "id": "p1", "label": "P1" }],
  "events": [],
  "unknown": true
}`)

	err := func() error {
		_, err := LoadData(dataPath)
		return err
	}()
	if err == nil {
		t.Fatal("LoadData() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "validate data:") {
		t.Fatalf("error = %q, want unknown field error", got)
	}
}

func TestLoadDataValidatesSchema(t *testing.T) {
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "7:00", "end": "19:00" },
  "people": [{ "id": "p1", "label": "P1" }],
  "events": []
}`)

	err := func() error {
		_, err := LoadData(dataPath)
		return err
	}()
	if err == nil {
		t.Fatal("LoadData() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "validate data:") {
		t.Fatalf("error = %q, want schema validation error", got)
	}
}

func TestLoadDataRequiresPeople(t *testing.T) {
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "events": []
}`)

	err := func() error {
		_, err := LoadData(dataPath)
		return err
	}()
	if err == nil {
		t.Fatal("LoadData() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "validate data:") {
		t.Fatalf("error = %q, want schema validation error", got)
	}
}

func writeTestData(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "agenda.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test data: %v", err)
	}
	return path
}

func equalDays(a, b []Day) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
