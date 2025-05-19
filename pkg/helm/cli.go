package helm

import (
	"fmt"
	"os/exec"
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

func (cli *HelmCLI) RunCmd(arg ...string) ([]byte, error) {
	cmd := exec.Command(cli.helm, arg...)
	return cmd.Output()
}
