package session

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/utils/ptr"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types,client,spec -package session -o session_gen.go api.spec.yaml

type RenkuSessionClient struct {
	baseClient     *ClientWithResponses
	httpClient     *http.Client
	requestEditors []RequestEditorFn
}

func NewRenkuSessionClient(apiURL string, options ...RenkuSessionClientOption) (c *RenkuSessionClient, err error) {
	c = &RenkuSessionClient{}
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

type RenkuSessionClientOption func(*RenkuSessionClient) error

func WithHttpClient(httpClient *http.Client) RenkuSessionClientOption {
	return func(c *RenkuSessionClient) error {
		c.httpClient = httpClient
		return nil
	}
}

func WithRequestEditors(editors ...RequestEditorFn) RenkuSessionClientOption {
	return func(c *RenkuSessionClient) error {
		c.requestEditors = append(c.requestEditors, editors...)
		return nil
	}
}

func (c *RenkuSessionClient) GetGlobalEnvironments(ctx context.Context) (environments EnvironmentList, err error) {
	res, err := c.baseClient.GetEnvironmentsWithResponse(ctx, nil)
	if err != nil {
		return environments, err
	}
	if res.JSON200 == nil {
		message := ""
		if res.JSONDefault != nil {
			message = res.JSONDefault.Error.Message
		}
		if message != "" {
			return environments, fmt.Errorf("could not get global environments: %s", message)
		}
		return environments, fmt.Errorf("could not get global environments: HTTP %d", res.StatusCode())
	}
	return *res.JSON200, nil
}

func (c *RenkuSessionClient) PostGlobalEnvironment(ctx context.Context, body EnvironmentPost) (environment Environment, err error) {
	res, err := c.baseClient.PostEnvironmentsWithResponse(ctx, body)
	if err != nil {
		return environment, err
	}
	if res.JSON201 == nil {
		message := ""
		if res.JSONDefault != nil {
			message = res.JSONDefault.Error.Message
		}
		if message != "" {
			return environment, fmt.Errorf("could not get global environments: %s", message)
		}
		return environment, fmt.Errorf("could not get global environments: HTTP %d", res.StatusCode())
	}
	return *res.JSON201, nil
}

func (c *RenkuSessionClient) PatchGlobalEnvironment(ctx context.Context, environmentId Ulid, body EnvironmentPatch) (environment Environment, err error) {
	res, err := c.baseClient.PatchEnvironmentsEnvironmentIdWithResponse(ctx, environmentId, body)
	if err != nil {
		return environment, err
	}
	if res.JSON200 == nil {
		message := ""
		if res.JSON404 != nil {
			message = res.JSON404.Error.Message
		}
		if res.JSONDefault != nil {
			message = res.JSONDefault.Error.Message
		}
		if message != "" {
			return environment, fmt.Errorf("could not get global environments: %s", message)
		}
		return environment, fmt.Errorf("could not get global environments: HTTP %d", res.StatusCode())
	}
	return *res.JSON200, nil
}

func (c *RenkuSessionClient) UpdateGlobalImages(ctx context.Context, images []string, tag string, existingEnvironments EnvironmentList, dryRun bool) error {
	if dryRun {
		fmt.Println("The following updates would be performed:")
	} else {
		fmt.Println("Performing the following updates:")
	}
	for _, image := range images {
		_, err := c.updateGlobalImage(ctx, image, tag, existingEnvironments, dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *RenkuSessionClient) updateGlobalImage(ctx context.Context, image string, tag string, existingEnvironments EnvironmentList, dryRun bool) (environment Environment, err error) {
	var existing *Environment
	for _, env := range existingEnvironments {
		existingImage, existingTag, _ := strings.Cut(env.ContainerImage, ":")
		if existingImage != image {
			continue
		}
		if existingTag == "" {
			existingTag = "latest"
		}
		if existingTag == tag {
			fmt.Printf("= untouched: %s:%s\n", existingImage, existingTag)
			return env, nil
		}
		fmt.Printf("~ update: %s:%s -> %s\n", existingImage, existingTag, tag)
		existing = &env
	}

	if existing == nil {
		fmt.Printf("+ add: %s:%s\n", image, tag)
		if dryRun {
			return environment, nil
		}
		body := getDefaultEnvironmentPost()
		body.ContainerImage = fmt.Sprintf("%s:%s", image, tag)
		body.Name = body.ContainerImage
		return c.PostGlobalEnvironment(ctx, body)
	}

	if dryRun {
		return *existing, nil
	}
	patch := EnvironmentPatch{
		ContainerImage: ptr.To(fmt.Sprintf("%s:%s", image, tag)),
	}
	return c.PatchGlobalEnvironment(ctx, existing.Id, patch)
}

func getDefaultEnvironmentPost() EnvironmentPost {
	return EnvironmentPost{
		ContainerImage:         "", // Leave blank
		DefaultUrl:             ptr.To("/"),
		Description:            ptr.To("Created by renku-dev-utils"),
		EnvironmentImageSource: Image,
		Uid:                    ptr.To(1000),
		Gid:                    ptr.To(1000),
		Name:                   "", // Leave blank
		Port:                   ptr.To(8888),
		MountDirectory:         ptr.To("/home/renku/work"),
		WorkingDirectory:       ptr.To("/home/renku/work"),
	}
}
