package helm

import (
	"context"
	"encoding/json"
	"fmt"
)

func (cli *HelmCLI) ListReleases(ctx context.Context, namespace string) (releases []string, err error) {
	out, err := cli.RunCmd(ctx, "list", "--namespace", namespace, "--all", "--output", "json")
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

func (cli *HelmCLI) UninstallReleases(ctx context.Context, namespace string, releases []string) error {
	args := []string{"uninstall", "--wait", "--namespace", namespace}
	for _, release := range releases {
		args = append(args, release)
	}

	out, err := cli.RunCmd(ctx, args...)
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

func (cli *HelmCLI) UninstallAllReleases(ctx context.Context, namespace string) error {
	releases, err := cli.ListReleases(ctx, namespace)
	if err != nil {
		return err
	}
	if len(releases) == 0 {
		return nil
	}
	return cli.UninstallReleases(ctx, namespace, releases)
}
