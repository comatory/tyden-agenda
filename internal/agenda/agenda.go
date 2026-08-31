package agenda

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

const usage = `Usage: tyden --data <path-to-data.json> [--show <option>] [--hide <option>]

Create a printable weekly agenda HTML document.

Flags:
  --data string
        path to agenda JSON data
  --show string
        optional render layer to show. Repeat or comma-separate. Supported: hour-labels
  --hide string
        default render layer to hide. Repeat or comma-separate. Supported: slot-times
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
	show := newShowFlags()
	flags.Var(show, "show", "optional render layer to show")
	hide := newHideFlags()
	flags.Var(hide, "hide", "default render layer to hide")
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

	return Render(data, stdout, renderOptions(show, hide))
}

type showFlags map[string]bool

func newShowFlags() showFlags {
	return make(showFlags)
}

func (flags showFlags) String() string {
	return ""
}

func (flags showFlags) Set(value string) error {
	for _, option := range strings.Split(value, ",") {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}
		switch option {
		case "hour-labels":
			flags[option] = true
		default:
			return fmt.Errorf("unknown --show option %q", option)
		}
	}
	return nil
}

type hideFlags map[string]bool

func newHideFlags() hideFlags {
	return make(hideFlags)
}

func (flags hideFlags) String() string {
	return ""
}

func (flags hideFlags) Set(value string) error {
	for _, option := range strings.Split(value, ",") {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}
		switch option {
		case "slot-times":
			flags[option] = true
		default:
			return fmt.Errorf("unknown --hide option %q", option)
		}
	}
	return nil
}

func renderOptions(show showFlags, hide hideFlags) RenderOptions {
	return RenderOptions{
		ShowHourLabels: show["hour-labels"],
		ShowSlotTimes:  !hide["slot-times"],
	}
}
