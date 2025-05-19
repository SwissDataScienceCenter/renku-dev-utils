package github

import (
	"context"
	"encoding/json"
)

func (cli *GitHubCLI) GetCurrentRepository(ctx context.Context) (string, error) {
	out, err := cli.RunCmd(ctx, "repo", "view", "--json", "nameWithOwner")
	if err != nil {
		return "", err
	}

	var res gitHubRepoViewOutput
	err = json.Unmarshal(out, &res)
	if err != nil {
		return "", err
	}

	return res.NameWithOwner, nil
}

type gitHubRepoViewOutput struct {
	NameWithOwner string `json:"nameWithOwner"`
}
