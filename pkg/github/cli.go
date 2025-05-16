package github

import (
	"fmt"
	"os/exec"
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

func (cli *GitHubCLI) RunCmd(arg ...string) ([]byte, error) {
	cmd := exec.Command(cli.gh, arg...)
	return cmd.Output()
}
