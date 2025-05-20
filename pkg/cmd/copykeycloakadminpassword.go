package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/k8s"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

var copyKeycloakAdminPasswordCmd = &cobra.Command{
	Use:     "copy-keycloak-admin-password",
	Aliases: []string{"ckap"},
	Short:   "Copy the Keycloak admin password to the clipboard",
	Run:     runCopyKeycloakAdminPassword,
}

func runCopyKeycloakAdminPassword(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	if namespace == "" {
		cli, err := github.NewGitHubCLI("")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		namespace, err = ns.FindCurrentNamespace(ctx, cli)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	clients, err := k8s.GetClientset()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	secret, err := k8s.GetSecret(ctx, clients, namespace, secretName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	secretValue, found := secret.Data[secretKey]
	if !found {
		fmt.Printf("The secret did not contain '%s'\n", secretKey)
		os.Exit(1)
	}

	if err := clipboard.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clipboard.Write(clipboard.FmtText, secretValue)
	fmt.Printf("Copied Keycloak admin password into the clipboard")
	fmt.Println()
}

func init() {
	copyKeycloakAdminPasswordCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "k8s namespace")
	copyKeycloakAdminPasswordCmd.Flags().StringVar(&secretName, "secret-name", "keycloak-password-secret", "secret name")
	copyKeycloakAdminPasswordCmd.Flags().StringVar(&secretKey, "secret-key", "KEYCLOAK_ADMIN_PASSWORD", "secret key")
}
