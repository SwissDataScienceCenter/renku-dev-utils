package github

import (
	"context"
	"encoding/json"
)

func (cli *GitHubCLI) GetCurrentPullRequest(ctx context.Context) (int, error) {
	out, err := cli.RunCmd(ctx, "pr", "view", "--json", "number")
	if err != nil {
		return 0, err
	}

	var res gitHubPRViewOutput
	err = json.Unmarshal(out, &res)
	if err != nil {
		return 0, err
	}

	return res.Number, nil
}

type gitHubPRViewOutput struct {
	Number int `json:"number"`
}
