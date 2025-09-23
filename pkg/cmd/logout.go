package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/renkuapi"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of a renku instance",
	Run:   logout,
}

func logout(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	logoutAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if logoutAll {
		err := renkuapi.LogoutAll(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	url, err := cmd.Flags().GetString("url")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if url == "" {
		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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

		deploymentURL, err := ns.GetDeploymentURL(namespace)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		url = deploymentURL.String()
	}

	fmt.Printf("Renku URL: %s\n", url)

	auth, err := renkuapi.NewRenkuApiAuth(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = auth.Logout(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	logoutCmd.Flags().String("url", "", "instance URL")
	logoutCmd.Flags().StringP("namespace", "n", "", "k8s namespace")
	logoutCmd.Flags().Bool("all", false, "remove all saved logins")
}
