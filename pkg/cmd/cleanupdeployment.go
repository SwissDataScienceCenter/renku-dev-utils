package cmd

import (
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/helm"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/spf13/cobra"
)

var cleanupDeploymentCmd = &cobra.Command{
	Use:   "cleanup-deployment",
	Short: "Cleanup a renku deployment",
	Run:   cleanupDeployment,
}

func cleanupDeployment(cmd *cobra.Command, args []string) {
	// ctx := context.Background()

	if namespace == "" {
		cli, err := github.NewGitHubCLI("")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		namespace, err = ns.FindCurrentNamespace(cli)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	helm, err := helm.NewHelmCLI("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = helm.UninstallAllReleases(namespace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// client, err := k8s.GetDynamicClient()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// err = k8s.ListJS(ctx, client, nil)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
}

func init() {
	cleanupDeploymentCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "k8s namespace")
	cleanupDeploymentCmd.Flags().BoolVar(&deleteNamespace, "delete-namespace", false, "if set, the namespace will be deleted")
}
