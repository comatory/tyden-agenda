package agenda

import (
	"fmt"
	"html/template"
	"io"
	"sort"
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
	Day        Day
	Label      string
	Row        int
	People     []personView
	Slots      []daySlotView
	Boundaries []eventBoundaryView
}

type personView struct {
	ID    string
	Label string
}

type eventView struct {
	Person   string
	Day      Day
	Start    int
	End      int
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

type eventBoundaryView struct {
	Day  Day
	Left string
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
	boundaries := make([]eventBoundaryView, 0, len(data.Events)*2)
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

		eventLeft := percent(eventStart-start, end-start)
		eventEndLeft := percent(eventEnd-start, end-start)

		events = append(events, eventView{
			Person:   event.Person,
			Day:      event.Day,
			Start:    eventStart,
			End:      eventEnd,
			Time:     event.Start + "-" + event.End,
			Title:    event.Title,
			Subtitle: event.Subtitle,
			Note:     event.Note,
			Style:    eventStyle(event.Style),
			Left:     eventLeft,
			Width:    percent(eventEnd-eventStart, end-start),
		})
		boundaries = append(boundaries,
			eventBoundaryView{Day: event.Day, Left: eventLeft},
			eventBoundaryView{Day: event.Day, Left: eventEndLeft},
		)
	}
	stackEvents(events)

	dayViews := make([]dayView, 0, len(days))
	for index, day := range days {
		people := make([]personView, 0, len(data.People))
		for _, person := range data.People {
			people = append(people, personView(person))
		}
		dayViews = append(dayViews, dayView{Day: day, Label: dayLabel(day), Row: index + 2, People: people, Slots: daySlots(slots, events, day), Boundaries: dayBoundaries(boundaries, day)})
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

func stackEvents(events []eventView) {
	groups := map[string][]int{}
	for index, event := range events {
		key := string(event.Day) + "\x00" + event.Person
		groups[key] = append(groups[key], index)
	}

	for _, indexes := range groups {
		sort.SliceStable(indexes, func(i, j int) bool {
			left := events[indexes[i]]
			right := events[indexes[j]]
			if left.Start == right.Start {
				return left.End < right.End
			}
			return left.Start < right.Start
		})

		clusters := overlappingClusters(events, indexes)
		for _, cluster := range clusters {
			assignClusterLanes(events, cluster)
		}
	}
}

func overlappingClusters(events []eventView, indexes []int) [][]int {
	var clusters [][]int
	for _, index := range indexes {
		if len(clusters) == 0 || events[index].Start >= clusterEnd(events, clusters[len(clusters)-1]) {
			clusters = append(clusters, []int{index})
			continue
		}
		clusters[len(clusters)-1] = append(clusters[len(clusters)-1], index)
	}
	return clusters
}

func clusterEnd(events []eventView, indexes []int) int {
	end := 0
	for _, index := range indexes {
		if events[index].End > end {
			end = events[index].End
		}
	}
	return end
}

func assignClusterLanes(events []eventView, indexes []int) {
	laneEnds := []int{}
	lanes := map[int]int{}
	for _, index := range indexes {
		lane := 0
		for lane < len(laneEnds) && events[index].Start < laneEnds[lane] {
			lane++
		}
		if lane == len(laneEnds) {
			laneEnds = append(laneEnds, events[index].End)
		} else {
			laneEnds[lane] = events[index].End
		}
		lanes[index] = lane
	}

	total := len(laneEnds)
	for _, index := range indexes {
		events[index].Top, events[index].Height = stackPosition(lanes[index], total)
	}
}

func stackPosition(index, total int) (string, string) {
	height := 100.0 / float64(total)
	top := 32.0 + float64(index)*height*0.56
	return fmt.Sprintf("%.4f%%", top), fmt.Sprintf("%.4f%%", height*0.56)
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
	for _, slot := range slots {
		view := daySlotView{slotView: slot}
		for _, event := range events {
			if event.Day == day && event.Left == slot.Left && event.Width == slot.Width {
				view.Populated = true
				break
			}
		}
		if !view.Populated {
			continue
		}
		daySlots = append(daySlots, view)
	}
	return daySlots
}

func dayBoundaries(boundaries []eventBoundaryView, day Day) []eventBoundaryView {
	var dayBoundaries []eventBoundaryView
	seen := map[string]bool{}
	for _, boundary := range boundaries {
		if boundary.Day != day || seen[boundary.Left] {
			continue
		}
		seen[boundary.Left] = true
		dayBoundaries = append(dayBoundaries, boundary)
	}
	return dayBoundaries
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
.slot { position: absolute; left: var(--left); width: var(--width); top: 0; height: 100%; text-align: center; font-size: 7pt; font-weight: 700; pointer-events: none; z-index: 1; }
.slot-label { position: relative; z-index: 4; display: inline-block; background: #fff; padding: 0 0.5mm; }
.slot small { font-size: 6pt; }
.event-boundary { position: absolute; left: var(--left); top: 0; height: 100%; border-left: 1px dashed #777; pointer-events: none; z-index: 2; }
.event { container-type: inline-size; position: absolute; left: var(--left); width: var(--width); top: var(--top); height: var(--height); border: 1px solid #555; background: #ddd; padding: 2mm 1mm 1mm; text-align: center; overflow: hidden; z-index: 3; }
.event.muted { background: #eee; color: #444; }
.event.outline { background: #fff; }
.event-title { font-size: clamp(7pt, calc(5.5pt + 8cqw), 11pt); font-weight: 700; line-height: 1.1; white-space: nowrap; }
.event-subtitle { font-size: clamp(5pt, calc(4.25pt + 4cqw), 7pt); margin-top: 0.5mm; white-space: nowrap; }
.event-note { position: absolute; right: 1mm; top: 1mm; max-width: 35%; font-size: clamp(4pt, calc(3.25pt + 3cqw), 6pt); white-space: nowrap; }
.event-time { position: absolute; left: var(--left); width: var(--width); top: 13%; text-align: center; font-size: 7pt; font-weight: 700; z-index: 3; }
.event-time span { display: inline-block; background: #fff; padding: 0 0.5mm; }
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
      {{ range .Boundaries }}<div class="event-boundary" data-marker-type="event-boundary" style="--left: {{ .Left }};"></div>{{ end }}
      {{ range .People }}{{ $person := .ID }}
        <section class="person-grid">
          {{ range $.MinuteLines }}<div class="line" data-marker-type="hour-guide" style="--left: {{ .Left }};">{{ if $.Options.ShowHourLabels }}<span>{{ .Label }}</span>{{ end }}</div>{{ end }}
          {{ range $.Events }}{{ if samePerson . $day $person }}{{ if $.Options.ShowSlotTimes }}<div class="event-time" style="--left: {{ .Left }}; --width: {{ .Width }};"><span>{{ .Time }}</span></div>{{ end }}{{ end }}{{ end }}
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
