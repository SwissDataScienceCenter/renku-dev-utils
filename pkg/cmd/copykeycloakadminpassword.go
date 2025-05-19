package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/k8s"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Flags

var namespace string
var secretName string
var secretKey string

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

		repo, err := cli.GetCurrentRepository()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Repo: %s", repo)
		fmt.Println()

		prNumber, err := cli.GetCurrentPullRequest()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Pull request: %d", prNumber)
		fmt.Println()

		namespace, err = github.DeriveK8sNamespace(repo, prNumber)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Derived namespace: %s", namespace)
		fmt.Println()
	}

	clients, err := k8s.GetClientset()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	secret, err := clients.CoreV1().Secrets(namespace).Get(ctx, secretName, v1.GetOptions{})
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
