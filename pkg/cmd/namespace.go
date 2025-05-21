package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/spf13/cobra"
)

var namespaceCmd = &cobra.Command{
	Use:     "namespace",
	Aliases: []string{"ns"},
	Short:   "Print the kubernetes namespace of the current deployment",
	Run:     namespaceFn,
}

func namespaceFn(cmd *cobra.Command, args []string) {
	ctx := context.Background()

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

	fmt.Printf("%s\n", namespace)
}
