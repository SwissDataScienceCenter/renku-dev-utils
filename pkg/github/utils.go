package github

import "fmt"

const RENKU_REPO string = "SwissDataScienceCenter/renku"
const RENKU_REPO_NAMESPCE_TEMPLATE string = "ci-renku-%d"

func DeriveK8sNamespace(repo string, pr int) (string, error) {
	if repo == RENKU_REPO {
		return fmt.Sprintf(RENKU_REPO_NAMESPCE_TEMPLATE, pr), nil
	}
	return "", fmt.Errorf("Could not derive namespace from repository: %s", repo)
}
