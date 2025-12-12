package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:               "rdu",
	Short:             "renku-dev-utils is a dev utility CLI",
	PersistentPreRunE: preRunRoot,
	RunE:              runRoot,
}

func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func preRunRoot(cmd *cobra.Command, args []string) error {
	return viper.BindPFlags(cmd.Flags())
}

func init() {
	rootCmd.AddCommand(cleanupDeploymentCmd)
	rootCmd.AddCommand(copyKeycloakAdminPasswordCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(makeMeAdminCmd)
	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(openDeploymentCmd)
	rootCmd.AddCommand(updateGlobalImagesCmd)
	rootCmd.AddCommand(versionCmd)
}

func Main() {
	if err := Execute(); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}
