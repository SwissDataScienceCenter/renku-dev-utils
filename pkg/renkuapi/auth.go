package renkuapi

import (
	"fmt"
	"net/url"

	"github.com/zalando/go-keyring"
)

type RenkuApiAuth struct {
	baseURL   *url.URL
	issuerURL *url.URL
}

func NewRenkuApiAuth(baseURL string) (auth *RenkuApiAuth, err error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.EscapedPath() == "/" {
		parsedURL.Path = ""
	}
	auth = &RenkuApiAuth{
		baseURL: parsedURL,
	}
	if auth.issuerURL == nil {
		auth.issuerURL = parsedURL.JoinPath("auth/realms/Renku")
	}
	return auth, nil
}

func (auth *RenkuApiAuth) GetAccessToken() (token string, err error) {
	token, err = auth.getAccessTokenFromKeyring()
	fmt.Println(token)
	fmt.Println(err)

	return "", fmt.Errorf("not implemented")
}

func (auth *RenkuApiAuth) getAccessTokenFromKeyring() (token string, err error) {
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "access_token")
	return keyring.Get(keyringService, kUser)
}

func (auth *RenkuApiAuth) getRefreshTokenFromKeyring() (token string, err error) {
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "refresh_token")
	return keyring.Get(keyringService, kUser)
}

func (auth *RenkuApiAuth) getKeyringUserPrefix() string {
	return fmt.Sprintf("rdu:%s", auth.baseURL.String())
}
