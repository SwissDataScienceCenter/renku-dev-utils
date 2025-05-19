package cmd

import (
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/cmd"
)

func Main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
