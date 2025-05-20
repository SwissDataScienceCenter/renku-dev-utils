package namespace

import (
	"fmt"
	"net/url"
)

func GetDeploymentURL(namespace string) (deploymentURL *url.URL, err error) {
	// TODO: Can we derive the URL by inspecting ingresses in the k8s namespace?
	openURL, err := url.Parse(fmt.Sprintf("https://%s.dev.renku.ch", namespace))
	if err != nil {
		return nil, err
	}
	return openURL, nil
}
