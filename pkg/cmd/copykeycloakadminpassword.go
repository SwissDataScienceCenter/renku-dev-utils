package cmd

import (
	"context"
	"fmt"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/k8s"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Flags

var namespace string
var secretName string
var secretKey string

var copyKeycloakAdminPasswordCmd = &cobra.Command{
	Use:     "copy-keycloak-admin-password",
	Aliases: []string{"ckap"},
	Short:   "Copy the Keycloak admin password to the clipboard",
	RunE:    runCopyKeycloakAdminPassword,
}

func runCopyKeycloakAdminPassword(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	clients, err := k8s.GetClientset()
	if err != nil {
		return err
	}

	secret, err := clients.CoreV1().Secrets(namespace).Get(ctx, secretName, v1.GetOptions{})
	if err != nil {
		return err
	}

	secretValue, found := secret.Data[secretKey]
	if !found {
		return fmt.Errorf("The secret did not contain '%s'", secretKey)
	}

	if err := clipboard.Init(); err != nil {
		return err
	}

	clipboard.Write(clipboard.FmtText, secretValue)
	fmt.Printf("Copied Keycloak admin password into the clipboard")
	fmt.Println()

	return nil
}

func init() {
	copyKeycloakAdminPasswordCmd.Flags().StringVarP(&namespace, "namespace", "n", "renku", "k8s namespace")
	copyKeycloakAdminPasswordCmd.Flags().StringVar(&secretName, "secret-name", "keycloak-password-secret", "secret name")
	copyKeycloakAdminPasswordCmd.Flags().StringVar(&secretKey, "secret-key", "KEYCLOAK_ADMIN_PASSWORD", "secret key")
}
