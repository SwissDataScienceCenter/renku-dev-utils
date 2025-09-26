package github

import (
	"context"
	"encoding/json"
)

const renkuRepository = "SwissDataScienceCenter/renku"

func (cli *GitHubCLI) GetLatestRenkuRelease(ctx context.Context) (string, error) {
	out, err := cli.RunCmd(ctx, "release", "view", "--repo", renkuRepository, "--json", "tagName")
	if err != nil {
		return "", err
	}

	var res gitHubReleaseViewOutput
	err = json.Unmarshal(out, &res)
	if err != nil {
		return "", err
	}

	return res.TagName, nil
}

type gitHubReleaseViewOutput struct {
	TagName string `json:"tagName"`
}
