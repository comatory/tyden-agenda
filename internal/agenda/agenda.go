package agenda

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

func Run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("tyden", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	dataPath := flags.String("data", "", "path to agenda JSON data")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *dataPath == "" {
		return errors.New("missing required --data")
	}

	_, err := fmt.Fprintf(stdout, "<!-- TODO render agenda from %s -->\n", *dataPath)
	return err
}
