package git

import (
	"context"
	"fmt"
	"strings"
)

func (cli *GitCLI) GetUserEmail(ctx context.Context) (email string, err error) {
	out, err := cli.RunCmd(ctx, "config", "--get", "user.email")
	if err != nil {
		return "", err
	}

	email = strings.TrimSpace(string(out))
	fmt.Printf("Detected user email: %s\n", email)
	return email, nil
}
