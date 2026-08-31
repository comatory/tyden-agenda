package main

import (
	"fmt"
	"os"

	"github.com/comatory/tyden-agenda/internal/agenda"
)

func main() {
	if err := agenda.Run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "tyden: %v\n", err)
		os.Exit(1)
	}
}
