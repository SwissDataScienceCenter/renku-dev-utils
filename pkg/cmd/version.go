package cmd

import (
	"fmt"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of renku-dev-utils",
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	// fmt.Printf("renku-dev-utils %s\n", version.Version)
	fmt.Printf("renku-dev-utils %s\n", version.BB())
}
