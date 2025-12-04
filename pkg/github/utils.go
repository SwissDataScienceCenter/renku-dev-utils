package github

import "fmt"

// Maps repositories to namespace templates
var repoToNamespaceTemplateMap map[string]string

// Contains the list of Renku global images
var globalImagesSlice []string

func DeriveK8sNamespace(repo string, pr int) (string, error) {
	tpl, found := repoToNamespaceTemplateMap[repo]
	if found {
		return fmt.Sprintf(tpl, pr), nil
	}
	return "", fmt.Errorf("could not derive namespace from repository: %s", repo)
}

func GetGlobalImages() []string {
	return globalImagesSlice[:]
}

func init() {
	initRepoToNamespaceTemplateMap()
	initGlobalImagesSlice()
}

func initRepoToNamespaceTemplateMap() {
	repoToNamespaceTemplateMap = map[string]string{
		"SwissDataScienceCenter/amalthea":            "renku-ci-am-%d",
		"SwissDataScienceCenter/renku":               "ci-renku-%d",
		"SwissDataScienceCenter/renku-data-services": "renku-ci-ds-%d",
		"SwissDataScienceCenter/renku-ui":            "renku-ci-ui-%d",
	}
}

func initGlobalImagesSlice() {
	// TODO: can we derive this from GitHub API calls?
	prefix := "ghcr.io/swissdatasciencecenter/renku"
	packageVariants := []string{"basic", "datascience"}
	frontendVariants := []string{"jupyterlab", "ttyd", "vscodium"}
	globalImagesSlice = make([]string, 0, len(packageVariants)*len(frontendVariants))
	for _, packageVariant := range packageVariants {
		for _, frontendVariant := range frontendVariants {
			globalImagesSlice = append(globalImagesSlice, fmt.Sprintf("%s/py-%s-%s", prefix, packageVariant, frontendVariant))
		}
	}
}
