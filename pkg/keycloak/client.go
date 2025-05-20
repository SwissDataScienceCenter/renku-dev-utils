package keycloak

import (
	"net/http"
	"net/url"
	"sync"
)

type KeycloakClient struct {
	BaseURL    *url.URL
	AdminRealm string

	httpClient *http.Client

	tokenSet   TokenSet
	tokenSetMu *sync.RWMutex
}

func NewKeycloakClient(baseURL string) (client *KeycloakClient, err error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	client = &KeycloakClient{
		BaseURL:    parsed,
		AdminRealm: "master",

		httpClient: &http.Client{},

		tokenSet:   TokenSet{},
		tokenSetMu: &sync.RWMutex{},
	}
	return client, nil
}
