package github

import (
	"fmt"
	"regexp"
	"strconv"
)

// Maps repositories to namespace templates
var repoToNamespaceTemplateMap map[string]string

// Deployment namespace regexes
var namespaceRegexes []namespaceRegex

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

func MatchDeploymentNamespace(namespace string) (repository string, pr int) {
	for i := range namespaceRegexes {
		res := namespaceRegexes[i].regex.FindStringSubmatch(namespace)
		if res != nil {
			pr, err := strconv.Atoi(res[1])
			if err == nil && pr > 0 {
				return namespaceRegexes[i].repository, pr
			}
		}
	}
	return "", 0
}

type namespaceRegex struct {
	regex      *regexp.Regexp
	repository string
}

func init() {
	initRepoToNamespaceTemplateMap()
	initGlobalImagesSlice()
	initNamespaceRegexes()
}

func initRepoToNamespaceTemplateMap() {
	repoToNamespaceTemplateMap = map[string]string{
		"SwissDataScienceCenter/amalthea":            "renku-ci-am-%d",
		"SwissDataScienceCenter/renku":               "ci-renku-%d",
		"SwissDataScienceCenter/renku-data-services": "renku-ci-ds-%d",
		"SwissDataScienceCenter/renku-ui":            "renku-ci-ui-%d",
		"SwissDataScienceCenter/renku-gateway":       "renku-ci-gw-%d",
	}
}

func initNamespaceRegexes() {
	namespaceRegexes = []namespaceRegex{
		{
			regex:      regexp.MustCompile(`^renku-ci-am-(\d+)$`),
			repository: "SwissDataScienceCenter/amalthea",
		},
		{
			regex:      regexp.MustCompile(`^ci-renku-(\d+)$`),
			repository: "SwissDataScienceCenter/renku",
		},
		{
			regex:      regexp.MustCompile(`^renku-ci-ds-(\d+)$`),
			repository: "SwissDataScienceCenter/renku-data-services",
		},
		{
			regex:      regexp.MustCompile(`^renku-ci-ui-(\d+)$`),
			repository: "SwissDataScienceCenter/renku-ui",
		},
		{
			regex:      regexp.MustCompile(`^renku-ci-gw-(\d+)$`),
			repository: "SwissDataScienceCenter/renku-gateway",
		},
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
