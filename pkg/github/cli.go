package github

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/executils"
)

type GitHubCLI struct {
	gh string
}

func NewGitHubCLI(gh string) (*GitHubCLI, error) {
	if gh == "" {
		gh = "gh"
	}

	path, err := exec.LookPath(gh)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found GitHub CLI: %s", path)
	fmt.Println()
	return &GitHubCLI{gh: gh}, nil
}

func (cli *GitHubCLI) RunCmd(ctx context.Context, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, cli.gh, arg...)
	return executils.FormatOutput(cmd.Output())
}
