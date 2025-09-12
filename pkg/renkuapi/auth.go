package renkuapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

const jsonContentType string = "application/json"

type RenkuApiAuth struct {
	baseURL           *url.URL
	issuerURL         *url.URL
	authenticationURI *url.URL
	tokenURI          *url.URL

	clientID string
	scope    string

	httpClient *http.Client
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
	if auth.clientID == "" {
		auth.clientID = "renku-cli"
	}
	if auth.scope == "" {
		auth.scope = "offline_access"
	}
	if auth.httpClient == nil {
		auth.httpClient = http.DefaultClient
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

func (auth *RenkuApiAuth) Login(ctx context.Context) error {
	err := auth.performLogin(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (auth *RenkuApiAuth) performLogin(ctx context.Context) error {
	deviceAuthorization, err := auth.startLogin(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("deviceAuthorization: %+v\n", deviceAuthorization)
	return fmt.Errorf("not implemented")
}

func (auth *RenkuApiAuth) startLogin(ctx context.Context) (result deviceAuthorization, err error) {
	authenticationURI, err := auth.getAuthenticationURI(ctx)
	if err != nil {
		return result, err
	}

	body := url.Values{}
	body.Set("client_id", auth.clientID)
	body.Set("scope", auth.scope)

	var res deviceAuthorizationResponse
	_, err = auth.postForm(ctx, authenticationURI.String(), body, &res)
	if err != nil {
		return result, err
	}

	result = deviceAuthorization{
		DeviceCode:              res.DeviceCode,
		VerificationURIComplete: res.VerificationURIComplete,
		ExpiresAt:               time.Now().Add(time.Second * time.Duration(res.ExpiresIn)),
		Interval:                time.Second * time.Duration(res.Interval),
	}
	if result.Interval == time.Duration(0) {
		result.Interval = time.Second * 5
	}
	return result, nil
}

type deviceAuthorization struct {
	DeviceCode              string
	VerificationURIComplete string
	ExpiresAt               time.Time
	Interval                time.Duration
}

type deviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int32  `json:"expires_in"`
	Interval                int32  `json:"interval"`
}

func (auth *RenkuApiAuth) getAuthenticationURI(ctx context.Context) (authenticationURI *url.URL, err error) {
	if auth.authenticationURI != nil {
		return auth.authenticationURI, nil
	}
	err = auth.getOpenIDConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	return auth.authenticationURI, nil
}

func (auth *RenkuApiAuth) getTokenURI(ctx context.Context) (tokenURI *url.URL, err error) {
	if auth.tokenURI != nil {
		return auth.tokenURI, nil
	}
	err = auth.getOpenIDConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	return auth.tokenURI, nil
}

func (auth *RenkuApiAuth) getOpenIDConfiguration(ctx context.Context) error {
	configurationURL := auth.issuerURL.JoinPath("./.well-known/openid-configuration")
	fmt.Printf("configurationURL: %s\n", configurationURL.String())
	var result openIDConfigurationResponse
	_, err := auth.get(ctx, configurationURL.String(), &result)
	if err != nil {
		return err
	}

	fmt.Printf("result: %+v\n", result)

	parsed, err := url.Parse(result.DeviceAuthorizationEndpoint)
	if err != nil {
		return err
	}
	auth.authenticationURI = parsed

	parsed, err = url.Parse(result.TokenEndpoint)
	if err != nil {
		return err
	}
	auth.tokenURI = parsed
	return nil
}

type openIDConfigurationResponse struct {
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
}

// TODO: refactor this method to avoid duplication with package keycloak

func (auth *RenkuApiAuth) get(ctx context.Context, url string, result any) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonContentType)

	resp, err = auth.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	var parseErr error
	if resp.Header.Get("Content-Type") == jsonContentType {
		parseErr = tryParseResponse(resp, result)
	} else {
		return resp, fmt.Errorf("Expected '%s' but got response with content type '%s'", jsonContentType, resp.Header.Get("Content-Type"))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && parseErr != nil {
		return resp, parseErr
	}

	return resp, nil
}

func (auth *RenkuApiAuth) postForm(ctx context.Context, url string, data url.Values, result any) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonContentType)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = auth.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	var parseErr error
	if resp.Header.Get("Content-Type") == jsonContentType {
		parseErr = tryParseResponse(resp, result)
	} else {
		return resp, fmt.Errorf("Expected '%s' but got response with content type '%s'", jsonContentType, resp.Header.Get("Content-Type"))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && parseErr != nil {
		return resp, parseErr
	}

	return resp, nil
}

func tryParseResponse(resp *http.Response, result any) error {
	defer resp.Body.Close()

	outBuf := new(bytes.Buffer)
	_, err := outBuf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(outBuf.Bytes(), result)
}
