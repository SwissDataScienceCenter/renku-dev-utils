package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/git"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/k8s"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/keycloak"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/spf13/cobra"
)

var makeMeAdminCmd = &cobra.Command{
	Use:     "make-me-admin",
	Aliases: []string{"mma"},
	Short:   "Makes you admin of the current deployment",
	Run:     makeMeAdmin,
}

func makeMeAdmin(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	if userEmail == "" {
		gitCli, err := git.NewGitCLI("")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		userEmail, err = gitCli.GetUserEmail(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

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

	username, found := secret.Data[secretKeyUsername]
	if !found {
		fmt.Printf("The secret did not contain '%s'\n", secretKeyUsername)
		os.Exit(1)
	}

	password, found := secret.Data[secretKey]
	if !found {
		fmt.Printf("The secret did not contain '%s'\n", secretKey)
		os.Exit(1)
	}

	deploymentURL, err := ns.GetDeploymentURL(namespace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	kcURL := deploymentURL.JoinPath("./auth")
	kcClient, err := keycloak.NewKeycloakClient(kcURL.String())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = kcClient.Authenticate(ctx, string(username), string(password))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	userID, err := kcClient.FindUser(ctx, renkuRealm, userEmail)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	isAdmin, err := kcClient.IsRenkuAdmin(ctx, renkuRealm, userID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if isAdmin {
		fmt.Printf("User '%s' is already a renku admin\n", userEmail)
		os.Exit(0)
	}

	err = kcClient.AddRenkuAdminRoleToUser(ctx, renkuRealm, userID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Done, you are now a Renku admin!")
}

func init() {
	makeMeAdminCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "k8s namespace")
	makeMeAdminCmd.Flags().StringVar(&secretName, "secret-name", "keycloak-password-secret", "secret name")
	makeMeAdminCmd.Flags().StringVar(&secretKey, "secret-key", "KEYCLOAK_ADMIN_PASSWORD", "secret key")
	makeMeAdminCmd.Flags().StringVar(&secretKeyUsername, "secret-key-username", "KEYCLOAK_ADMIN", "secret key for the admin username")
	makeMeAdminCmd.Flags().StringVar(&renkuRealm, "renku-realm", "Renku", "the Keycloak realm used by renku")
	makeMeAdminCmd.Flags().StringVarP(&userEmail, "user-email", "u", "", "your email")
}
