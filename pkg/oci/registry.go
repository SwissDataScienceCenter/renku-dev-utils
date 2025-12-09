package oci

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/distribution/reference"
)

type RegistryClient struct {
	tokens map[string]string
	client *http.Client
}

func NewRegistryClient() (rc *RegistryClient, err error) {
	rc = &RegistryClient{
		tokens: map[string]string{},
		client: http.DefaultClient,
	}
	return rc, nil
}

func (rc *RegistryClient) CheckImage(ctx context.Context, named reference.Named) error {
	manifestURL, err := GetManifestURLForImage(named)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "HEAD", manifestURL.String(), nil)
	if err != nil {
		return err
	}
	res, err := rc.client.Do(req)
	if err != nil {
		return err
	}

	// Check if authentication is required
	res.Header.Get(http.CanonicalHeaderKey())

	return fmt.Errorf("not implemented, %s", res.Status)
}

// func (rc *RegistryClient) getToken(ctx context.Context, )

func GetManifestURLForImage(named reference.Named) (url *url.URL, err error) {
	domain := reference.Domain(named)
	if domain == "docker.io" || strings.HasSuffix(domain, ".docker.io") {
		domain = "registry-1.docker.io"
	}
	path := reference.Path(named)
	ref := ""
	if digested, ok := named.(reference.Digested); ok {
		ref = digested.Digest().String()
	} else if tagged, ok := named.(reference.Tagged); ok {
		ref = tagged.Tag()
	}
	if ref == "" {
		return nil, fmt.Errorf("could not parse reference %s", named.String())
	}
	manifestURLStr := fmt.Sprintf("https://%s/v2/%s/manifests/%s", domain, path, ref)
	return url.Parse(manifestURLStr)
}
