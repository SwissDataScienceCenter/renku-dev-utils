package git

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/executils"
)

type GitCLI struct {
	git string
}

func NewGitCLI(git string) (*GitCLI, error) {
	if git == "" {
		git = "git"
	}

	path, err := exec.LookPath(git)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found git: %s", path)
	fmt.Println()
	return &GitCLI{git: git}, nil
}

func (cli *GitCLI) RunCmd(ctx context.Context, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, cli.git, arg...)
	return executils.FormatOutput(cmd.Output())
}
