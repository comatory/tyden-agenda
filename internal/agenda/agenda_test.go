package agenda

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); got != usage {
		t.Fatalf("stdout = %q, want %q", got, usage)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunRequiresData(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run(nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if got, want := err.Error(), "missing required --data"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"--unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); !strings.Contains(got, "flag provided but not defined: -unknown") {
		t.Fatalf("stderr = %q, want unknown flag message", got)
	}
}

func TestRunRendersData(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "events": [{ "day": "mon", "start": "08:00", "end": "08:45", "title": "Cj" }]
}`)

	err := Run([]string{"--data", dataPath}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "<!doctype html>") {
		t.Fatalf("stdout = %q, want HTML", got)
	}
	if got := stdout.String(); !strings.Contains(got, "Cj") {
		t.Fatalf("stdout = %q, want event title", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunShowsSlotTimesByDefault(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "slots": [{ "label": "1", "start": "08:00", "end": "08:45" }],
  "events": []
}`)

	err := Run([]string{"--data", dataPath}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "08:00-08:45") {
		t.Fatalf("stdout = %q, want slot time", got)
	}
}

func TestRunHidesSlotTimes(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "slots": [{ "label": "1", "start": "08:00", "end": "08:45" }],
  "events": []
}`)

	err := Run([]string{"--data", dataPath, "--hide", "slot-times"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); strings.Contains(got, "08:00-08:45") {
		t.Fatalf("stdout contains slot time: %q", got)
	}
}

func TestRunHidesCommaSeparatedOptions(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "slots": [{ "label": "1", "start": "08:00", "end": "08:45" }],
  "events": []
}`)

	err := Run([]string{"--data", dataPath, "--hide", "slot-times,"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); strings.Contains(got, "08:00-08:45") {
		t.Fatalf("stdout contains slot time: %q", got)
	}
}

func TestRunShowsHourLabels(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dataPath := writeTestData(t, `{
  "timeRange": { "start": "07:00", "end": "19:00" },
  "events": []
}`)

	err := Run([]string{"--data", dataPath, "--show", "hour-labels"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "<span>07:00</span>") {
		t.Fatalf("stdout = %q, want hour labels", got)
	}
}

func TestRunRejectsUnknownShowOption(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"--show", "unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "unknown --show option") {
		t.Fatalf("error = %q, want unknown show option", got)
	}
}

func TestRunRejectsUnknownHideOption(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"--hide", "unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "unknown --hide option") {
		t.Fatalf("error = %q, want unknown hide option", got)
	}
}

func TestRunReturnsDataFileError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{"--data", filepath.Join(t.TempDir(), "missing.json")}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "open data:") {
		t.Fatalf("error = %q, want open data error", got)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunReturnsJSONDecodeError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dataPath := writeTestData(t, `{`)

	err := Run([]string{"--data", dataPath}, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "decode data:") {
		t.Fatalf("error = %q, want decode data error", got)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}
