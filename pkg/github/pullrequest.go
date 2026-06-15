package github

import (
	"context"
	"encoding/json"
	"fmt"
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

func (cli *GitHubCLI) GetPullRequestState(ctx context.Context, repository string, pr int) (state string, err error) {
	out, err := cli.RunCmd(ctx, "pr", "view", "--repo", repository, "--json", "state", fmt.Sprintf("%d", pr))
	if err != nil {
		return "", err
	}

	var res gitHubPRViewStateOutput
	err = json.Unmarshal(out, &res)
	if err != nil {
		return "", err
	}

	return res.State, nil
}

type gitHubPRViewStateOutput struct {
	State string `json:"state"`
}
