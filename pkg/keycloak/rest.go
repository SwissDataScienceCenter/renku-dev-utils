package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const jsonContentType string = "application/json"

func (client *KeycloakClient) GetJSON(ctx context.Context, url string, result any) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonContentType)
	client.setAuthHeaders(req)

	resp, err = client.httpClient.Do(req)
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

func (client *KeycloakClient) PostJSON(ctx context.Context, url string, body any, result any) (resp *http.Response, err error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewReader(bodyBytes)

	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonContentType)
	req.Header.Set("Content-Type", jsonContentType)
	client.setAuthHeaders(req)

	resp, err = client.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	var parseErr error
	if resp.Header.Get("Content-Type") == jsonContentType {
		parseErr = tryParseResponse(resp, result)
	} else if resp.StatusCode == 204 {
		// No content
		return resp, nil
	} else {
		return resp, fmt.Errorf("Expected '%s' but got response with content type '%s'", jsonContentType, resp.Header.Get("Content-Type"))
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && parseErr != nil {
		return resp, parseErr
	}

	return resp, nil
}

func (client *KeycloakClient) PostForm(ctx context.Context, url string, data url.Values, result any) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", jsonContentType)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client.setAuthHeaders(req)

	resp, err = client.httpClient.Do(req)
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

func tryParseResponse(resp *http.Response, result any) error {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Warning, could not close HTTP response: %s", err.Error())
		}
	}()

	outBuf := new(bytes.Buffer)
	_, err := outBuf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(outBuf.Bytes(), result)
}
