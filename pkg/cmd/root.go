package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rdu",
	Short: "renku-dev-utils is a dev utility CLI",
	RunE:  runRoot,
}

func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func init() {
	rootCmd.AddCommand(copyKeycloakAdminPasswordCmd)
	rootCmd.AddCommand(versionCmd)
}

func Main() {
	if err := Execute(); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}
