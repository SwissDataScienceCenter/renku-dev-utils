package namespace

import (
	"context"
	"fmt"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
)

func FindCurrentNamespace(ctx context.Context, cli *github.GitHubCLI) (namespace string, err error) {
	repo, err := cli.GetCurrentRepository(ctx)
	if err != nil {
		return "", err
	}
	fmt.Printf("Repository: %s", repo)
	fmt.Println()

	prNumber, err := cli.GetCurrentPullRequest(ctx)
	if err != nil {
		return "", err
	}
	fmt.Printf("Pull request: %d", prNumber)
	fmt.Println()

	namespace, err = github.DeriveK8sNamespace(repo, prNumber)
	if err != nil {
		return "", err
	}
	fmt.Printf("Derived namespace: %s", namespace)
	fmt.Println()
	return namespace, nil
}
