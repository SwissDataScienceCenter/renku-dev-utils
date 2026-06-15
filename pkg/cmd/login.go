package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/renkuapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a renku instance",
	Run:   login,
}

func login(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	url := viper.GetString("url")
	namespace := viper.GetString("namespace")

	if url == "" {
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

	rac, err := renkuapi.NewRenkuApiClient(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = rac.Auth().Login(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ruc, err := rac.Users()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	userInfo, err := ruc.GetUser(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Logged in as:")
	fmt.Printf("  username: %s\n", userInfo.Username)
	fmt.Printf("  email: %s\n", *userInfo.Email)
	fmt.Printf("  first name: %s\n", *userInfo.FirstName)
	fmt.Printf("  last name: %s\n", *userInfo.LastName)
	fmt.Printf("  is admin: %t\n", userInfo.IsAdmin)
}

func init() {
	loginCmd.Flags().String("url", "", "instance URL")
	loginCmd.Flags().StringP("namespace", "n", "", "k8s namespace")
}
