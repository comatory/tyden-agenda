This is a Go codebase. This is a CLI tool which ingests JSON data and produces HTML.
For verification, use `golangci-lint` to lint the code, `go build` to build it.
Ensure code is formatted with `go fmt` after producing changes.

Renderer concept: `slots` are configured lesson/activity periods from input data. Dashed vertical lines mark each slot's start and end boundary and use `data-marker-type="slot-boundary"`. They are not generic full-hour markers; full-hour guide lines are separate solid/light lines and use `data-marker-type="hour-guide"`.
