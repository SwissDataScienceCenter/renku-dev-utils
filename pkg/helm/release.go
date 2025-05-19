package helm

import (
	"encoding/json"
	"fmt"
)

func (cli *HelmCLI) ListReleases(namespace string) (releases []string, err error) {
	out, err := cli.RunCmd("list", "--namespace", namespace, "--all", "--output", "json")
	if err != nil {
		return nil, err
	}

	var res []helmListOutput
	err = json.Unmarshal(out, &res)
	if err != nil {
		return nil, err
	}

	for _, release := range res {
		releases = append(releases, release.Name)
	}
	return releases, nil
}

type helmListOutput struct {
	Name string
}

func (cli *HelmCLI) UninstallReleases(namespace string, releases []string) error {
	args := []string{"uninstall", "--dry-run", "--namespace", namespace}
	for _, release := range releases {
		args = append(args, release)
	}

	out, err := cli.RunCmd(args...)
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

func (cli *HelmCLI) UninstallAllReleases(namespace string) error {
	releases, err := cli.ListReleases(namespace)
	if err != nil {
		return err
	}
	return cli.UninstallReleases(namespace, releases)
}
