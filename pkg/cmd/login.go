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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a renku instance",
	Run:   login,
}

func login(cmd *cobra.Command, args []string) {
	ctx := context.Background()

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

	fmt.Printf("URL '%s'\n", url)

	auth, err := renkuapi.NewRenkuApiAuth(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = auth.Login(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	loginCmd.Flags().String("url", "", "instance URL")
	loginCmd.Flags().StringP("namespace", "n", "", "k8s namespace")
}
