package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Token struct {
	Raw string
}

type TokenSet struct {
	AccessToken  Token
	RefreshToken Token
}

func (client *KeycloakClient) Authenticate(ctx context.Context, username string, password string) error {
	postURL := client.GetAdminTokenURL()

	body := url.Values{}
	body.Set("client_id", "admin-cli")
	body.Set("grant_type", "password")
	body.Set("username", username)
	body.Set("password", password)

	var result authenticateResponse
	_, err := client.PostForm(ctx, postURL.String(), body, &result)
	if err != nil {
		return err
	}

	client.setTokenSet(TokenSet{
		AccessToken:  Token{Raw: result.AccessToken},
		RefreshToken: Token{Raw: result.RefreshToken},
	})

	return nil
}

func (client *KeycloakClient) GetAdminTokenURL() *url.URL {
	path := fmt.Sprintf("./realms/%s/protocol/openid-connect/token", client.AdminRealm)
	return client.BaseURL.JoinPath(path)
}

type authenticateResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
}

func (client *KeycloakClient) getTokenSet() TokenSet {
	client.tokenSetMu.RLock()
	defer client.tokenSetMu.RUnlock()
	return client.tokenSet
}

func (client *KeycloakClient) setTokenSet(tokens TokenSet) {
	client.tokenSetMu.Lock()
	defer client.tokenSetMu.Unlock()
	client.tokenSet = tokens
}

func (client *KeycloakClient) setAuthHeaders(req *http.Request) {
	tokens := client.getTokenSet()
	if tokens.AccessToken.Raw != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken.Raw))
	}
}
