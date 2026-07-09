// Command atlas is the entrypoint for the Atlas CLI.
package main

import (
	"fmt"
	"os"

	"github.com/Haykhay/atlas/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "atlas:", err)
		os.Exit(1)
	}
}
