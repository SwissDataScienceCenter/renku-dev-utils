package oci

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/containerd/containerd/v2/core/remotes/docker/auth"
	dockerListSpec "github.com/distribution/distribution/v3/manifest/manifestlist"
	dockerSpec "github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/reference"
	ociSpec "github.com/opencontainers/image-spec/specs-go/v1"
)

type RegistryClient struct {
	client *http.Client
}

func NewRegistryClient() (rc *RegistryClient, err error) {
	rc = &RegistryClient{
		client: http.DefaultClient,
	}
	return rc, nil
}

func (rc *RegistryClient) CheckImage(ctx context.Context, named reference.Named) (res *http.Response, err error) {
	manifestURL, err := GetManifestURLForImage(named)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "HEAD", manifestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", ociSpec.MediaTypeImageIndex)
	req.Header.Add("Accept", ociSpec.MediaTypeImageManifest)
	req.Header.Add("Accept", dockerListSpec.MediaTypeManifestList)
	req.Header.Add("Accept", dockerSpec.MediaTypeManifest)
	res, err = rc.client.Do(req)
	if err != nil {
		return res, err
	}

	// Check if authentication is required
	if res.StatusCode == http.StatusUnauthorized {
		challenges := auth.ParseAuthHeader(res.Header)
		var challenge *auth.Challenge = nil
		token := ""
		for i := range challenges {
			to, err := auth.GenerateTokenOptions(ctx, manifestURL.Host, "", "", challenges[i])
			if err != nil {
				log.Printf("could not generate token options from challenge: %+v", challenges[i])
				continue
			}
			tokenRes, err := auth.FetchToken(ctx, rc.client, http.Header{}, to)
			if err != nil {
				log.Printf("could not fetch token: %s", err.Error())
				continue
			}
			challenge = &challenges[i]
			token = tokenRes.Token
			break
		}
		if challenge == nil {
			return nil, fmt.Errorf("could not authenticate with registry at %s", manifestURL.Host)
		}
		req, err := http.NewRequestWithContext(ctx, "HEAD", manifestURL.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Accept", ociSpec.MediaTypeImageIndex)
		req.Header.Add("Accept", ociSpec.MediaTypeImageManifest)
		req.Header.Add("Accept", dockerListSpec.MediaTypeManifestList)
		req.Header.Add("Accept", dockerSpec.MediaTypeManifest)
		scheme := "Bearer"
		switch challenge.Scheme {
		case auth.BasicAuth:
			scheme = "Basic"
		case auth.DigestAuth:
			scheme = "Digest"
		}
		req.Header.Add("Authorization", fmt.Sprintf("%s %s", scheme, token))
		res, err = rc.client.Do(req)
		if err != nil {
			return res, err
		}
	}

	if res.StatusCode != http.StatusOK {
		return res, fmt.Errorf("image does not exist: %s", res.Status)
	}

	contentType := strings.ToLower(res.Header.Get("Content-Type"))
	if contentType != ociSpec.MediaTypeImageIndex && contentType != ociSpec.MediaTypeImageManifest && contentType != dockerListSpec.MediaTypeManifestList && contentType != dockerSpec.MediaTypeManifest {
		return res, fmt.Errorf("unexpected response content type %s", contentType)
	}

	return res, nil
}

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
