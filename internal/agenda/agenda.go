package agenda

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

const usage = `Usage: tyden --data <path-to-data.json>

Create a printable weekly agenda HTML document.

Flags:
  --data string
        path to agenda JSON data
  --help
        show help
`

func Run(args []string, stdout, stderr io.Writer) error {
	var usageErr error
	flags := flag.NewFlagSet("tyden", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, usageErr = fmt.Fprint(stderr, usage)
	}

	dataPath := flags.String("data", "", "path to agenda JSON data")
	help := flags.Bool("help", false, "show help")
	if err := flags.Parse(args); err != nil {
		if usageErr != nil {
			return usageErr
		}
		return err
	}
	if *help {
		_, err := fmt.Fprint(stdout, usage)
		return err
	}
	if *dataPath == "" {
		return errors.New("missing required --data")
	}
	data, err := LoadData(*dataPath)
	if err != nil {
		return err
	}

	return Render(data, stdout)
}
