package users

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types,client,spec -package users -o users_gen.go api.spec.yaml

type RenkuUsersClient struct {
	baseClient     *ClientWithResponses
	httpClient     *http.Client
	requestEditors []RequestEditorFn
}

func NewRenkuUsersClient(apiURL string, options ...RenkuUsersClientOption) (c *RenkuUsersClient, err error) {
	c = &RenkuUsersClient{}
	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	// Ensure we use the correct API URL
	if !strings.HasSuffix(parsedURL.EscapedPath(), "/api/data") {
		parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/"
		parsedURL = parsedURL.JoinPath("api/data")
	}
	// Create httpClient, if not already present
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	// Create client
	clientOpts := []ClientOption{WithHTTPClient(c.httpClient)}
	for _, fn := range c.requestEditors {
		clientOpts = append(clientOpts, WithRequestEditorFn(fn))
	}
	client, err := NewClientWithResponses(parsedURL.String(), clientOpts...)
	if err != nil {
		return nil, err
	}
	c.baseClient = client
	return c, nil
}

type RenkuUsersClientOption func(*RenkuUsersClient) error

func WithHttpClient(httpClient *http.Client) RenkuUsersClientOption {
	return func(c *RenkuUsersClient) error {
		c.httpClient = httpClient
		return nil
	}
}

func WithRequestEditors(editors ...RequestEditorFn) RenkuUsersClientOption {
	return func(c *RenkuUsersClient) error {
		c.requestEditors = append(c.requestEditors, editors...)
		return nil
	}
}

func (c *RenkuUsersClient) GetUser(ctx context.Context) (userInfo SelfUserInfo, err error) {
	res, err := c.baseClient.GetUserWithResponse(ctx)
	if err != nil {
		return userInfo, err
	}
	if res.JSON200 == nil {
		message := ""
		if res.JSONDefault != nil {
			message = res.JSONDefault.Error.Message
		}
		if message != "" {
			return userInfo, fmt.Errorf("could not get user info: %s", message)
		}
		return userInfo, fmt.Errorf("could not get user info: HTTP %d", res.StatusCode())
	}
	return *res.JSON200, nil
}
