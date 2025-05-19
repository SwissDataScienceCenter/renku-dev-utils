package github

import "fmt"

// Maps repositories to namespace templates
var repoToNamespaceTemplateMap map[string]string

func DeriveK8sNamespace(repo string, pr int) (string, error) {
	tpl, found := repoToNamespaceTemplateMap[repo]
	if found {
		return fmt.Sprintf(tpl, pr), nil
	}
	return "", fmt.Errorf("Could not derive namespace from repository: %s", repo)
}

func init() {
	initRepoToNamespaceTemplateMap()
}

func initRepoToNamespaceTemplateMap() {
	repoToNamespaceTemplateMap = map[string]string{
		"SwissDataScienceCenter/renku":               "ci-renku-%d",
		"SwissDataScienceCenter/renku-data-services": "renku-ci-ds-%d",
		"SwissDataScienceCenter/renku-ui":            "renku-ci-ui-%d",
	}
}
