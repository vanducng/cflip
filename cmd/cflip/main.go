package main

import (
	"fmt"
	"os"

	"github.com/vanducng/cflip/internal/cli"
)

// Build information
var (
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	if err := cli.Execute("", commit, buildTime); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
