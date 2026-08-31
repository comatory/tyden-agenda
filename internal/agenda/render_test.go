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
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots:     []Slot{{Label: "1", Start: "08:00", End: "08:45"}},
		Events: []Event{{
			Person:   "p1",
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
		"P1",
		"Tue",
		"Cj",
		"FrE",
		"1C",
		"08:00-08:45",
		"data-marker-type=\"slot-boundary\"",
		"data-marker-type=\"hour-guide\"",
		"--top: 60.0000%; --height: 28.0000%;",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, html)
		}
	}
}

func TestRenderHidesSlotTimesWhenNoEvents(t *testing.T) {
	data := Data{
		People:    []Person{{ID: "p1", Label: "P1"}},
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
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots: []Slot{
			{Label: "1", Start: "08:00", End: "08:45"},
			{Label: "2", Start: "08:55", End: "09:40"},
		},
		Events: []Event{{Person: "p1", Day: "mon", Start: "08:00", End: "08:45", Title: "Cj"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); strings.Contains(got, ">2<br") {
		t.Fatalf("rendered HTML contains slot after last populated slot:\n%s", got)
	}
}

func TestRenderShowsTimeAboveEvent(t *testing.T) {
	data := Data{
		Days:      []Day{"thu"},
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots:     []Slot{{Label: "1", Start: "08:00", End: "08:45"}},
		Events:    []Event{{Person: "p1", Day: "thu", Start: "17:30", End: "18:30", Title: "Anglictina"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); !strings.Contains(got, "17:30-18:30") {
		t.Fatalf("rendered HTML missing activity time:\n%s", got)
	}
}

func TestRenderMultiplePeople(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		People:    []Person{{ID: "p1", Label: "P1"}, {ID: "p2", Label: "P2"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Events:    []Event{{Person: "p2", Day: "mon", Start: "08:00", End: "08:45", Title: "Cj"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := output.String()
	for _, want := range []string{"P1", "P2", "Cj"} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, html)
		}
	}
	if got := strings.Count(html, "<div class=\"day-name\">Mon</div>"); got != 1 {
		t.Fatalf("Mon label count = %d, want 1:\n%s", got, html)
	}
}

func TestRenderSlotLabelsOncePerDay(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		People:    []Person{{ID: "p1", Label: "P1"}, {ID: "p2", Label: "P2"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots:     []Slot{{Label: "1", Start: "08:00", End: "08:45"}},
		Events:    []Event{{Person: "p1", Day: "mon", Start: "08:00", End: "08:45", Title: "Cj"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := strings.Count(output.String(), "class=\"slot\" data-marker-type=\"slot-boundary\""); got != 1 {
		t.Fatalf("slot count = %d, want 1:\n%s", got, output.String())
	}
	if got := output.String(); !strings.Contains(got, "08:00-08:45") {
		t.Fatalf("rendered HTML missing populated slot time:\n%s", got)
	}
}

func TestRenderHidesEventTimesWhenConfigured(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots:     []Slot{{Label: "1", Start: "08:00", End: "08:45"}},
		Events:    []Event{{Person: "p1", Day: "mon", Start: "08:00", End: "08:45", Title: "Cj"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: false}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); strings.Contains(got, "08:00-08:45") {
		t.Fatalf("rendered HTML contains event time:\n%s", got)
	}
}

func TestRenderRejectsInvalidStack(t *testing.T) {
	data := Data{
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Events: []Event{{
			Person: "p1",
			Day:    "mon",
			Start:  "08:00",
			End:    "08:45",
			Title:  "Cj",
			Stack:  Stack{Index: 2, Total: 2},
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
		People:    []Person{{ID: "p1", Label: "P1"}},
		Events:    []Event{{Person: "p1", Day: "fri", Start: "08:00", End: "08:45", Title: "M"}},
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
		People:    []Person{{ID: "p1", Label: "P1"}},
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
