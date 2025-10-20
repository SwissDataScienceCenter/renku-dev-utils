package renkuapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/executils"
	"github.com/golang-jwt/jwt/v5"
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

	accessToken  string
	refreshToken string

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

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// RequestEditor returns a request editor which injects a valid access token
// for API requests.
func (auth *RenkuApiAuth) RequestEditor() RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if req.Header.Get("Authorization") != "" {
			return nil
		}
		token, err := auth.GetAccessToken(ctx)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}
}

func (auth *RenkuApiAuth) GetAccessToken(ctx context.Context) (token string, err error) {
	// Use access token if valid
	token = auth.accessToken
	if token != "" && isTokenValid(token) {
		return token, nil
	}
	token, err = auth.getAccessTokenFromKeyring()
	if err != nil {
		token = ""
	}
	if token != "" && isTokenValid(token) {
		return token, nil
	}

	// Refresh the access token if possible
	refreshToken := auth.refreshToken
	if refreshToken == "" {
		refreshToken, err = auth.getRefreshTokenFromKeyring()
		if err != nil {
			refreshToken = ""
		}
	}
	if refreshToken == "" {
		return "", fmt.Errorf("could not get access token")
	}
	tokenResult, err := auth.postRefeshToken(ctx, refreshToken)
	if err != nil {
		return token, nil
	}
	auth.accessToken = tokenResult.AccessToken
	auth.refreshToken = tokenResult.RefreshToken
	err = auth.saveAccessTokenToKeyring()
	if err != nil {
		return auth.accessToken, err
	}
	err = auth.saveRefreshTokenToKeyring()
	if err != nil {
		return auth.accessToken, err
	}
	return auth.accessToken, nil
}

func isTokenValid(token string) (isValid bool) {
	claims := jwt.RegisteredClaims{}
	parser := jwt.NewParser()
	_, _, err := parser.ParseUnverified(token, &claims)
	if err != nil {
		return false
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return false
	}
	now := time.Now()
	leeway := time.Second * 10
	return now.Before(exp.Add(-leeway))
}

func (auth *RenkuApiAuth) getAccessTokenFromKeyring() (token string, err error) {
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "access_token")
	token, err = keyring.Get(keyringService, kUser)
	if err != nil {
		return token, err
	}
	auth.accessToken = token
	return token, nil
}

func (auth *RenkuApiAuth) saveAccessTokenToKeyring() (err error) {
	if auth.accessToken == "" {
		return fmt.Errorf("access_token is not set")
	}
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "access_token")
	return keyring.Set(keyringService, kUser, auth.accessToken)
}

func (auth *RenkuApiAuth) deleteAccessTokenFromKeyring() (err error) {
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "access_token")
	return keyring.Delete(keyringService, kUser)
}

func (auth *RenkuApiAuth) getRefreshTokenFromKeyring() (token string, err error) {
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "refresh_token")
	return keyring.Get(keyringService, kUser)
}

func (auth *RenkuApiAuth) saveRefreshTokenToKeyring() (err error) {
	if auth.refreshToken == "" {
		return fmt.Errorf("refresh_token is not set")
	}
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "refresh_token")
	return keyring.Set(keyringService, kUser, auth.refreshToken)
}

func (auth *RenkuApiAuth) deleteRefreshTokenFromKeyring() (err error) {
	kUser := fmt.Sprintf("%s:%s", auth.getKeyringUserPrefix(), "refresh_token")
	return keyring.Delete(keyringService, kUser)
}

func (auth *RenkuApiAuth) getKeyringUserPrefix() string {
	return fmt.Sprintf("rdu:%s", auth.baseURL.String())
}

func (auth *RenkuApiAuth) Login(ctx context.Context) error {
	token, _ := auth.GetAccessToken(ctx)
	if token != "" {
		return nil
	}
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
	err = openBrowser(ctx, deviceAuthorization.VerificationURIComplete)
	if err != nil {
		return err
	}
	tokenResult, err := auth.pollTokenEndpoint(ctx, deviceAuthorization)
	if err != nil {
		return err
	}
	auth.accessToken = tokenResult.AccessToken
	auth.refreshToken = tokenResult.RefreshToken
	err = auth.saveAccessTokenToKeyring()
	if err != nil {
		return err
	}
	return auth.saveRefreshTokenToKeyring()
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
	var result openIDConfigurationResponse
	_, err := auth.get(ctx, configurationURL.String(), &result)
	if err != nil {
		return err
	}

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
		return resp, fmt.Errorf("expected '%s' but got response with content type '%s'", jsonContentType, resp.Header.Get("Content-Type"))
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
		return resp, fmt.Errorf("expected '%s' but got response with content type '%s'", jsonContentType, resp.Header.Get("Content-Type"))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && parseErr != nil {
		return resp, parseErr
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// TODO: try to get the error from the response
		return resp, fmt.Errorf("got non successful response '%s'", resp.Status)
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

// TODO: refactor this to avoid duplication with opendeployment.go
func openBrowser(ctx context.Context, openURL string) error {
	if runtime.GOOS == "darwin" {
		fmt.Printf("Opening: %s\n", openURL)
		cmd := exec.CommandContext(ctx, "open", openURL)
		_, err := executils.FormatOutput(cmd.Output())
		if err != nil {
			return err
		}
		return nil
	}

	if runtime.GOOS == "linux" {
		fmt.Printf("Opening: %s\n", openURL)
		cmd := exec.CommandContext(ctx, "xdg-open", openURL)
		_, err := executils.FormatOutput(cmd.Output())
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Printf("Open this link in your browser: %s\n", openURL)
	return nil
}

func (auth *RenkuApiAuth) pollTokenEndpoint(ctx context.Context, deviceAuthorization deviceAuthorization) (result tokenResult, err error) {
	deadline, cancel := context.WithDeadline(ctx, deviceAuthorization.ExpiresAt.Add(deviceAuthorization.Interval))
	defer cancel()

	ticker := time.NewTicker(deviceAuthorization.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.Done():
			return result, deadline.Err()
		case <-ticker.C:
			result, err := auth.postToken(deadline, deviceAuthorization.DeviceCode)
			if err == nil {
				return result, nil
			}
		}
	}
}

func (auth *RenkuApiAuth) postToken(ctx context.Context, deviceCode string) (result tokenResult, err error) {
	tokenURI, err := auth.getTokenURI(ctx)
	if err != nil {
		return result, err
	}

	body := url.Values{}
	body.Set("client_id", auth.clientID)
	body.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	body.Set("device_code", deviceCode)

	var res tokenResponse
	_, err = auth.postForm(ctx, tokenURI.String(), body, &res)
	if err != nil {
		return result, err
	}

	result = tokenResult{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
	return result, nil
}

type tokenResult struct {
	AccessToken  string
	RefreshToken string
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int32  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int32  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int32  `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func (auth *RenkuApiAuth) postRefeshToken(ctx context.Context, refreshToken string) (result tokenResult, err error) {
	tokenURI, err := auth.getTokenURI(ctx)
	if err != nil {
		return result, err
	}

	body := url.Values{}
	body.Set("client_id", auth.clientID)
	body.Set("grant_type", "refresh_token")
	body.Set("refresh_token", refreshToken)

	var res tokenResponse
	_, err = auth.postForm(ctx, tokenURI.String(), body, &res)
	if err != nil {
		return result, err
	}

	result = tokenResult{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}
	return result, nil
}

func (auth *RenkuApiAuth) Logout(ctx context.Context) error {
	err1 := auth.deleteAccessTokenFromKeyring()
	err2 := auth.deleteRefreshTokenFromKeyring()
	if err1 != nil && err2 != nil {
		return fmt.Errorf("got errors: %w and %w", err1, err2)
	}
	if err1 != nil {
		return err1
	}
	return err2
}

func LogoutAll(ctx context.Context) error {
	return keyring.DeleteAll(keyringService)

}
