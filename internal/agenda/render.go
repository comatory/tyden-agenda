package agenda

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
)

var defaultDays = []Day{"mon", "tue", "wed", "thu", "fri"}

type RenderOptions struct {
	ShowHourLabels bool
	ShowSlotTimes  bool
}

func Render(data Data, writer io.Writer, options RenderOptions) error {
	view, err := newRenderView(data, options)
	if err != nil {
		return err
	}

	return agendaTemplate.Execute(writer, view)
}

type renderView struct {
	Title       string
	TimeStart   string
	TimeEnd     string
	Days        []dayView
	Slots       []slotView
	Events      []eventView
	MinuteLines []minuteLineView
	Options     RenderOptions
}

type dayView struct {
	Day    Day
	Label  string
	Row    int
	People []personView
	Slots  []daySlotView
}

type personView struct {
	ID    string
	Label string
}

type eventView struct {
	Person   string
	Day      Day
	Time     string
	Title    string
	Subtitle string
	Note     string
	Style    string
	Left     string
	Width    string
	Top      string
	Height   string
}

type slotView struct {
	Label string
	Start string
	End   string
	Left  string
	Width string
}

type daySlotView struct {
	slotView
	Populated bool
}

type minuteLineView struct {
	Label string
	Left  string
}

func newRenderView(data Data, options RenderOptions) (renderView, error) {
	start, err := parseMinutes(data.TimeRange.Start)
	if err != nil {
		return renderView{}, fmt.Errorf("parse time range start: %w", err)
	}
	end, err := parseMinutes(data.TimeRange.End)
	if err != nil {
		return renderView{}, fmt.Errorf("parse time range end: %w", err)
	}
	if end <= start {
		return renderView{}, fmt.Errorf("time range end must be after start")
	}

	days := data.Days
	if len(days) == 0 {
		days = defaultDays
	}

	slots := make([]slotView, 0, len(data.Slots))
	for _, slot := range data.Slots {
		slotStart, err := parseMinutes(slot.Start)
		if err != nil {
			return renderView{}, fmt.Errorf("parse slot %q start: %w", slot.Label, err)
		}
		slotEnd, err := parseMinutes(slot.End)
		if err != nil {
			return renderView{}, fmt.Errorf("parse slot %q end: %w", slot.Label, err)
		}
		if slotEnd <= slotStart {
			return renderView{}, fmt.Errorf("slot %q end must be after start", slot.Label)
		}

		slots = append(slots, slotView{
			Label: slot.Label,
			Start: slot.Start,
			End:   slot.End,
			Left:  percent(slotStart-start, end-start),
			Width: percent(slotEnd-slotStart, end-start),
		})
	}

	events := make([]eventView, 0, len(data.Events))
	for _, event := range data.Events {
		eventStart, err := parseMinutes(event.Start)
		if err != nil {
			return renderView{}, fmt.Errorf("parse event %q start: %w", event.Title, err)
		}
		eventEnd, err := parseMinutes(event.End)
		if err != nil {
			return renderView{}, fmt.Errorf("parse event %q end: %w", event.Title, err)
		}
		if eventEnd <= eventStart {
			return renderView{}, fmt.Errorf("event %q end must be after start", event.Title)
		}

		stackTop, stackHeight, err := stackPosition(event.Stack)
		if err != nil {
			return renderView{}, fmt.Errorf("event %q stack: %w", event.Title, err)
		}

		events = append(events, eventView{
			Person:   event.Person,
			Day:      event.Day,
			Time:     event.Start + "-" + event.End,
			Title:    event.Title,
			Subtitle: event.Subtitle,
			Note:     event.Note,
			Style:    eventStyle(event.Style),
			Left:     percent(eventStart-start, end-start),
			Width:    percent(eventEnd-eventStart, end-start),
			Top:      stackTop,
			Height:   stackHeight,
		})
	}

	dayViews := make([]dayView, 0, len(days))
	for index, day := range days {
		people := make([]personView, 0, len(data.People))
		for _, person := range data.People {
			people = append(people, personView(person))
		}
		dayViews = append(dayViews, dayView{Day: day, Label: dayLabel(day), Row: index + 2, People: people, Slots: daySlots(slots, events, day)})
	}

	title := data.Title
	if title == "" {
		title = "tyden agenda"
	}
	if data.Week != "" {
		title += " " + data.Week
	}

	return renderView{
		Title:       title,
		TimeStart:   data.TimeRange.Start,
		TimeEnd:     data.TimeRange.End,
		Days:        dayViews,
		Slots:       slots,
		Events:      events,
		MinuteLines: minuteLines(start, end),
		Options:     options,
	}, nil
}

func parseMinutes(value string) (int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time %q", value)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid time %q", value)
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid time %q", value)
	}
	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("invalid time %q", value)
	}

	return hours*60 + minutes, nil
}

func percent(value, total int) string {
	return fmt.Sprintf("%.4f%%", float64(value)/float64(total)*100)
}

func dayLabel(day Day) string {
	switch day {
	case "mon":
		return "Mon"
	case "tue":
		return "Tue"
	case "wed":
		return "Wed"
	case "thu":
		return "Thu"
	case "fri":
		return "Fri"
	case "sat":
		return "Sat"
	case "sun":
		return "Sun"
	default:
		return string(day)
	}
}

func eventStyle(style string) string {
	switch style {
	case "muted", "outline":
		return style
	default:
		return "normal"
	}
}

func stackPosition(stack Stack) (string, string, error) {
	if stack.Total == 0 && stack.Index == 0 {
		return "32.0000%", "56.0000%", nil
	}
	if stack.Total <= 0 {
		return "", "", fmt.Errorf("total must be positive")
	}
	if stack.Index < 0 || stack.Index >= stack.Total {
		return "", "", fmt.Errorf("index must be between 0 and total - 1")
	}

	height := 100.0 / float64(stack.Total)
	top := 32.0 + float64(stack.Index)*height*0.56
	return fmt.Sprintf("%.4f%%", top), fmt.Sprintf("%.4f%%", height*0.56), nil
}

func minuteLines(start, end int) []minuteLineView {
	firstHour := start
	if firstHour%60 != 0 {
		firstHour += 60 - firstHour%60
	}

	var lines []minuteLineView
	for minute := firstHour; minute <= end; minute += 60 {
		lines = append(lines, minuteLineView{
			Label: fmt.Sprintf("%02d:%02d", minute/60, minute%60),
			Left:  percent(minute-start, end-start),
		})
	}
	return lines
}

func daySlots(slots []slotView, events []eventView, day Day) []daySlotView {
	daySlots := make([]daySlotView, 0, len(slots))
	lastPopulated := -1
	for _, slot := range slots {
		view := daySlotView{slotView: slot}
		for _, event := range events {
			if event.Day == day && event.Left == slot.Left && event.Width == slot.Width {
				view.Populated = true
				lastPopulated = len(daySlots)
				break
			}
		}
		daySlots = append(daySlots, view)
	}
	if lastPopulated == -1 {
		return nil
	}

	daySlots = daySlots[:lastPopulated+1]
	return daySlots
}

var agendaTemplate = template.Must(template.New("agenda").Funcs(template.FuncMap{
	"samePerson": func(event eventView, day Day, person string) bool { return event.Day == day && event.Person == person },
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{ .Title }}</title>
<style>
@page { size: 297mm 210mm; margin: 0; }
* { box-sizing: border-box; }
html, body { margin: 0; width: 297mm; min-height: 210mm; color: #111; font-family: Arial, Helvetica, sans-serif; }
body { padding: 6mm; overflow: hidden; }
.agenda { display: grid; grid-template-columns: 28mm 1fr; grid-template-rows: 14mm repeat({{ len .Days }}, 1fr); width: 285mm; height: 198mm; border: 1px solid #777; overflow: hidden; }
.title { grid-column: 1 / -1; display: flex; justify-content: space-between; align-items: end; padding: 0 3mm 2mm; border-bottom: 1px solid #777; font-size: 15pt; font-weight: 700; }
.range { font-size: 8pt; font-weight: 400; }
.day-label { display: grid; grid-template-columns: 12mm 1fr; grid-template-rows: repeat({{ len (index .Days 0).People }}, 1fr); border-right: 1px solid #777; border-bottom: 1px solid #aaa; }
.day-name { grid-row: 1 / -1; padding: 2mm; font-weight: 700; border-right: 1px solid #ddd; }
.day-person { display: flex; align-items: center; padding: 1mm; border-bottom: 1px solid #ddd; font-size: 8pt; }
.day-person:last-child { border-bottom: 0; }
.day-grid { position: relative; display: grid; grid-template-rows: repeat({{ len (index .Days 0).People }}, 1fr); border-bottom: 1px solid #aaa; }
.person-grid { position: relative; border-bottom: 1px solid #ddd; background: #fff; }
.person-grid:last-child { border-bottom: 0; }
.line { position: absolute; left: var(--left); height: 100%; border-left: 1px solid #ccc; font-size: 6pt; color: #555; }
.line span { position: absolute; top: 1mm; transform: translateX(-50%); background: #fff; padding: 0 0.5mm; white-space: nowrap; }
.slot { position: absolute; left: var(--left); width: var(--width); top: 0; height: 100%; border-right: 1px dashed #777; border-left: 1px dashed #bbb; text-align: center; font-size: 7pt; font-weight: 700; pointer-events: none; z-index: 2; }
.slot-label { display: inline-block; background: #fff; padding: 0 0.5mm; }
.slot small { font-size: 6pt; }
.event { position: absolute; left: var(--left); width: var(--width); top: var(--top); height: var(--height); min-width: 18mm; border: 1px solid #555; background: #ddd; padding: 2mm 1mm 1mm; text-align: center; overflow: hidden; }
.event.muted { background: #eee; color: #444; }
.event.outline { background: #fff; }
.event-title { font-size: 11pt; font-weight: 700; line-height: 1.1; }
.event-subtitle { font-size: 7pt; margin-top: 0.5mm; }
.event-note { position: absolute; right: 1mm; top: 1mm; font-size: 6pt; }
.event-time { position: absolute; left: var(--left); width: var(--width); top: 13%; text-align: center; font-size: 7pt; font-weight: 700; z-index: 3; }
@media print {
  html, body { width: 297mm; height: 210mm; min-height: 0; }
  body { padding: 6mm; overflow: hidden; }
  .agenda { break-inside: avoid; page-break-inside: avoid; }
  * { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
}
</style>
</head>
<body>
<main class="agenda">
  <header class="title"><span>{{ .Title }}</span><span class="range">{{ .TimeStart }}-{{ .TimeEnd }}</span></header>
  {{ range .Days }}{{ $day := .Day }}
    <div class="day-label" style="grid-row: {{ .Row }};">
      <div class="day-name">{{ .Label }}</div>
      {{ range .People }}<div class="day-person">{{ .Label }}</div>{{ end }}
    </div>
    <section class="day-grid" style="grid-row: {{ .Row }};">
      {{ range .Slots }}<div class="slot" style="--left: {{ .Left }}; --width: {{ .Width }};"><span class="slot-label">{{ .Label }}</span></div>{{ end }}
      {{ range .People }}{{ $person := .ID }}
        <section class="person-grid">
          {{ range $.MinuteLines }}<div class="line" style="--left: {{ .Left }};">{{ if $.Options.ShowHourLabels }}<span>{{ .Label }}</span>{{ end }}</div>{{ end }}
          {{ range $.Events }}{{ if samePerson . $day $person }}{{ if $.Options.ShowSlotTimes }}<div class="event-time" style="--left: {{ .Left }}; --width: {{ .Width }};">{{ .Time }}</div>{{ end }}{{ end }}{{ end }}
          {{ range $.Events }}{{ if samePerson . $day $person }}<article class="event {{ .Style }}" style="--left: {{ .Left }}; --width: {{ .Width }}; --top: {{ .Top }}; --height: {{ .Height }};">
            {{ if .Note }}<div class="event-note">{{ .Note }}</div>{{ end }}
            <div class="event-title">{{ .Title }}</div>
            {{ if .Subtitle }}<div class="event-subtitle">{{ .Subtitle }}</div>{{ end }}
          </article>{{ end }}{{ end }}
        </section>
      {{ end }}
    </section>
  {{ end }}
</main>
</body>
</html>
`))
