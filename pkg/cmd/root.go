package cmd

import (
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
}
