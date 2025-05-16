package main

import (
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
