# tyden agenda

Create week agendas. Use case is to set up recurring events in external data file and then use this tool to print it out on a paper.

## Install

Install with Go:

```shell
go install github.com/comatory/tyden-agenda/cmd/tyden@latest
```

This installs the `tyden` command into `GOBIN`, or `GOPATH/bin` when `GOBIN` is not set.

## Development

Build:

```shell
go build ./cmd/tyden
```

Format:

```shell
go fmt ./...
```

Test:

```shell
go test ./...
```

Lint:

```shell
golangci-lint run
```

## How to use

First you need to create the data, see [data source](#data-source). Then run the tool to produce printable HTML

```shell
tyden --data <path-to-data.json> > agenda.html
```

You can then open the `agenda.html` in any browser. The HTML is already styled and ready to be printed.

Use `--locale` to render localized day names. Locale values use standard BCP 47 language tags. Default is `en`.

```shell
tyden --data <path-to-data.json> --locale cs > agenda.html
```

Supported locales:

| Value | Description |
| --- | --- |
| `en` | English day labels. Default. |
| `cs` | Czech day labels. |

### Adding translations

Translations live in `internal/agenda/locale.go`.

To add a locale:

1. Add a new case in `newLocale`.
2. Add a translation function, for example `germanTranslations`.
3. Return only that locale's map from the new case.
4. Add the locale to this README and CLI usage text.
5. Add or update render tests.

Translation keys use dot notation. Day labels use `day.mon`, `day.tue`, `day.wed`, `day.thu`, `day.fri`, `day.sat`, `day.sun`.

Do not create one global map with all languages. Keep translations behind per-locale functions so only the selected locale map is instantiated at runtime.

Optional render layers are enabled with `--show`. Default render layers are disabled with `--hide`. Repeat the flag or comma-separate values:

```shell
tyden --data <path-to-data.json> --show hour-labels > agenda.html
```

Supported `--show` values:

| Value | Description |
| --- | --- |
| `hour-labels` | Show hour labels on each day grid. Hidden by default. |

Supported `--hide` values:

| Value | Description |
| --- | --- |
| `slot-times` | Hide lesson slot time ranges. Visible by default. |

## Data source

The data source is one JSON file. It is intended to be edited by hand, kept in git, and reused week after week.

The first supported source type is a weekly template: define the week metadata, visible time range, optional lesson slots, and events placed on weekdays.

Machine-readable schema: [`schema/agenda.schema.json`](schema/agenda.schema.json).

Minimal example:

```json
{
  "$schema": "./schema/agenda.schema.json",
  "timeRange": {
    "start": "07:00",
    "end": "19:00"
  },
  "events": [
    {
      "day": "mon",
      "start": "08:00",
      "end": "08:45",
      "title": "Cj"
    }
  ]
}
```

Example with optional metadata and lesson slots:

```json
{
  "$schema": "./schema/agenda.schema.json",
  "title": "School week",
  "week": "2026-W36",
  "timezone": "Europe/Prague",
  "days": ["mon", "tue", "wed", "thu", "fri"],
  "timeRange": {
    "start": "07:00",
    "end": "19:00"
  },
  "slots": [
    { "label": "0", "start": "07:00", "end": "07:45" },
    { "label": "1", "start": "08:00", "end": "08:45" },
    { "label": "2", "start": "08:55", "end": "09:40" }
  ],
  "events": [
    {
      "day": "mon",
      "start": "08:00",
      "end": "08:45",
      "title": "Cj",
      "subtitle": "FrE",
      "note": "1C"
    },
    {
      "day": "mon",
      "start": "08:00",
      "end": "08:45",
      "title": "Aj",
      "subtitle": "Blz"
    },
    {
      "day": "thu",
      "start": "17:30",
      "end": "18:30",
      "title": "Anglictina"
    }
  ]
}
```

Fields:

| Field | Required | Description |
| --- | --- | --- |
| `$schema` | no | Path or URL to the JSON Schema. Useful for editor validation. |
| `title` | no | Optional human-readable agenda title. Omitted by default. |
| `week` | no | Optional ISO week label, for example `2026-W36`. Used as printed heading when present. |
| `timezone` | no | Optional IANA timezone name. Defaults to local runtime timezone if omitted. |
| `days` | no | Weekdays to render, in order. Defaults to Monday-Friday. |
| `timeRange` | yes | Visible agenda range. Events should fit inside this range. |
| `slots` | no | Named time blocks shown in the header, useful for school lessons. |
| `events` | yes | Agenda items rendered on the grid. |

Time values use 24-hour `HH:MM` format.

Day values are `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, `sun`.

Person fields:

| Field | Required | Description |
| --- | --- | --- |
| `id` | yes | Stable identifier used by events. |
| `label` | yes | Short label printed in the day/person row header. |

Event fields:

| Field | Required | Description |
| --- | --- | --- |
| `day` | yes | Weekday where the event appears. |
| `start` | yes | Start time in `HH:MM`. |
| `end` | yes | End time in `HH:MM`. |
| `title` | yes | Main text shown inside the event. |
| `subtitle` | no | Secondary text, for example teacher or room. |
| `note` | no | Small note, for example class/group marker. |
| `style` | no | Optional rendering hint. Supported initial values: `normal`, `muted`, `outline`. |

Overlapping events for the same person and day are stacked automatically. Put only real event data in JSON; the renderer computes vertical lanes from event time ranges.

Open decisions for later:

| Topic | Current direction |
| --- | --- |
| Recurrence | Not in v1. Repeat events by listing them explicitly. |
| Dates | Use weekday + time first. Add exact dates later only if needed. |
| Multiple agendas | One JSON file represents one printable weekly agenda. |
| Validation | The CLI should validate required fields and time ordering before rendering. |
