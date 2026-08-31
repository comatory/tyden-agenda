package agenda

import (
	"bytes"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	data := Data{
		Title:     "School week",
		Week:      "2026-W36",
		Days:      []Day{"mon", "tue"},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots:     []Slot{{Label: "1", Start: "08:00", End: "08:45"}},
		Events: []Event{{
			Day:      "mon",
			Start:    "08:00",
			End:      "08:45",
			Title:    "Cj",
			Subtitle: "FrE",
			Note:     "1C",
			Stack:    Stack{Index: 1, Total: 2},
		}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowHourLabels: true, ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := output.String()
	for _, want := range []string{
		"<!doctype html>",
		"School week 2026-W36",
		"07:00-19:00",
		"Mon",
		"Tue",
		"Cj",
		"FrE",
		"1C",
		"08:00-08:45",
		"--top: 56.0000%; --height: 36.0000%;",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, html)
		}
	}
}

func TestRenderHidesUnpopulatedSlotTimes(t *testing.T) {
	data := Data{
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots:     []Slot{{Label: "1", Start: "08:00", End: "08:45"}},
		Events:    []Event{},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); strings.Contains(got, "08:00-08:45") {
		t.Fatalf("rendered HTML contains unpopulated slot time:\n%s", got)
	}
}

func TestRenderHidesSlotLabelsAfterLastPopulatedSlot(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots: []Slot{
			{Label: "1", Start: "08:00", End: "08:45"},
			{Label: "2", Start: "08:55", End: "09:40"},
		},
		Events: []Event{{Day: "mon", Start: "08:00", End: "08:45", Title: "Cj"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); strings.Contains(got, ">2<br") {
		t.Fatalf("rendered HTML contains slot after last populated slot:\n%s", got)
	}
}

func TestRenderRejectsInvalidStack(t *testing.T) {
	data := Data{
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Events: []Event{{
			Day:   "mon",
			Start: "08:00",
			End:   "08:45",
			Title: "Cj",
			Stack: Stack{Index: 2, Total: 2},
		}},
	}

	err := Render(data, &bytes.Buffer{}, RenderOptions{})
	if err == nil {
		t.Fatal("Render() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "index must be between 0 and total - 1") {
		t.Fatalf("error = %q, want stack error", got)
	}
}

func TestRenderDefaultsDays(t *testing.T) {
	data := Data{
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Events:    []Event{{Day: "fri", Start: "08:00", End: "08:45", Title: "M"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{"Mon", "Tue", "Wed", "Thu", "Fri", "M"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, output.String())
		}
	}
}

func TestRenderRejectsInvalidTimes(t *testing.T) {
	data := Data{
		TimeRange: TimeRange{Start: "19:00", End: "07:00"},
		Events:    []Event{},
	}

	err := Render(data, &bytes.Buffer{}, RenderOptions{ShowSlotTimes: true})
	if err == nil {
		t.Fatal("Render() error = nil, want error")
	}
	if got := err.Error(); !strings.Contains(got, "time range end must be after start") {
		t.Fatalf("error = %q, want time range error", got)
	}
}
