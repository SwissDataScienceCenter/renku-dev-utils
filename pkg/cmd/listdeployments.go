package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/k8s"
	"github.com/spf13/cobra"
)

var listDeploymentsCmd = &cobra.Command{
	Use:     "list-deployments",
	Aliases: []string{"lsd"},
	Short:   "List renku deployments",
	Run:     listDeployments,
}

func listDeployments(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	cli, err := github.NewGitHubCLI("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clients, err := k8s.GetClientset()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	namespaceList, err := k8s.ListNamespaces(ctx, clients)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := range namespaceList.Items {
		ns := namespaceList.Items[i]
		repo, pr := github.MatchDeploymentNamespace(ns.Name)
		if repo != "" {
			state, err := cli.GetPullRequestState(ctx, repo, pr)
			if err != nil {
				state = "UNKNOWN"
				fmt.Println(err)
			}
			fmt.Printf("%s\t%s\t%d\t%s\n", ns.Name, repo, pr, state)
		} else {
			fmt.Printf("%s\n", ns.Name)
		}
	}
}
