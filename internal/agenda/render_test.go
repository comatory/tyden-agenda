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
		`<time datetime="08:00">08:00</time>-<time datetime="08:45">08:45</time>`,
		"data-marker-type=\"event-boundary\"",
		"data-marker-type=\"hour-guide\"",
		"--top: 32.0000%; --height: 56.0000%;",
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
	if got := output.String(); strings.Contains(got, `<time datetime="08:00">08:00</time>`) {
		t.Fatalf("rendered HTML contains unpopulated slot time:\n%s", got)
	}
}

func TestRenderHidesUnpopulatedSlotLabels(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Slots: []Slot{
			{Label: "1", Start: "08:00", End: "08:45"},
			{Label: "2", Start: "08:55", End: "09:40"},
			{Label: "3", Start: "10:00", End: "10:45"},
		},
		Events: []Event{{Person: "p1", Day: "mon", Start: "10:00", End: "10:45", Title: "Cj"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := output.String(); strings.Contains(got, ">1</span>") || strings.Contains(got, ">2</span>") {
		t.Fatalf("rendered HTML contains unpopulated slot label:\n%s", got)
	}
	if got := output.String(); !strings.Contains(got, ">3</span>") {
		t.Fatalf("rendered HTML missing populated slot label:\n%s", got)
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
	if got := output.String(); !strings.Contains(got, `<time datetime="17:30">17:30</time>-<time datetime="18:30">18:30</time>`) {
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
	if got := strings.Count(output.String(), "class=\"slot\""); got != 1 {
		t.Fatalf("slot count = %d, want 1:\n%s", got, output.String())
	}
	if got := output.String(); !strings.Contains(got, `<time datetime="08:00">08:00</time>-<time datetime="08:45">08:45</time>`) {
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
	if got := output.String(); strings.Contains(got, `<time datetime="08:00">08:00</time>`) {
		t.Fatalf("rendered HTML contains event time:\n%s", got)
	}
}

func TestRenderStacksOverlappingEvents(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Events: []Event{{
			Person: "p1",
			Day:    "mon",
			Start:  "07:30",
			End:    "15:00",
			Title:  "Skolka",
		}, {
			Person: "p1",
			Day:    "mon",
			Start:  "10:00",
			End:    "10:45",
			Title:  "Keramika",
		}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	got := output.String()
	for _, want := range []string{
		"--top: 32.0000%; --height: 28.0000%;",
		"--top: 60.0000%; --height: 28.0000%;",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, got)
		}
	}
}

func TestRenderShrinksLongEventText(t *testing.T) {
	data := Data{
		Days:      []Day{"mon"},
		People:    []Person{{ID: "p1", Label: "P1"}},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		Events:    []Event{{Person: "p1", Day: "mon", Start: "13:40", End: "13:55", Title: "Logopedie", Subtitle: "dlouhy text", Note: "note"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	got := output.String()
	for _, want := range []string{
		"--title-size: 5.0pt;",
		"--subtitle-size: 4.0pt;",
		"--note-size: 3.0pt;",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, got)
		}
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

func TestRenderLocalizesCzechLabelsAndTimes(t *testing.T) {
	data := Data{
		Days:      []Day{"mon", "tue", "wed", "thu", "fri", "sat", "sun"},
		TimeRange: TimeRange{Start: "07:00", End: "19:00"},
		People:    []Person{{ID: "p1", Label: "P1"}},
		Events:    []Event{{Person: "p1", Day: "mon", Start: "08:00", End: "08:45", Title: "M"}},
	}

	var output bytes.Buffer
	if err := Render(data, &output, RenderOptions{ShowSlotTimes: true, ShowHourLabels: true, Locale: "cs"}); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	got := output.String()
	for _, want := range []string{`<html lang="cs">`, "07:00-19:00", `<time datetime="08:00">08:00</time>-<time datetime="08:45">08:45</time>`, "Po", "Út", "St", "Čt", "Pá", "So", "Ne"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, got)
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
