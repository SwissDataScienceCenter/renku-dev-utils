package helm

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/executils"
)

type HelmCLI struct {
	helm string
}

func NewHelmCLI(helm string) (cli *HelmCLI, err error) {
	if helm == "" {
		helm = "helm"
	}

	path, err := exec.LookPath(helm)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found Helm CLI: %s", path)
	fmt.Println()
	return &HelmCLI{helm: helm}, nil
}

func (cli *HelmCLI) RunCmd(ctx context.Context, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, cli.helm, arg...)
	return executils.FormatOutput(cmd.Output())
}
