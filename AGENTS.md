This is a Go codebase. This is a CLI tool which ingests JSON data and produces HTML.
For verification, use `golangci-lint` to lint the code, `go build` to build it.
Ensure code is formatted with `go fmt` after producing changes.

Renderer concept: `slots` are configured lesson/activity periods from input data and provide labels. Dashed vertical lines mark rendered event start and end boundaries and use `data-marker-type="event-boundary"`. They are not generic full-hour markers; full-hour guide lines are separate solid/light lines and use `data-marker-type="hour-guide"`.
