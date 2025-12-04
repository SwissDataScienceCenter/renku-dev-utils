package keycloak

import (
	"context"
	"fmt"
	"net/url"
)

const renkuAdminRole string = "renku-admin"

func (client *KeycloakClient) FindUser(ctx context.Context, realm string, email string) (userID string, err error) {
	getURL := client.GetAdminUsersURL(realm)

	query := url.Values{}
	query.Set("email", email)
	query.Set("exact", "true")
	getURL.RawQuery = query.Encode()

	var result []getUsersResponse
	_, err = client.GetJSON(ctx, getURL.String(), &result)
	if err != nil {
		return "", err
	}

	for _, user := range result {
		if user.Email == email {
			userID = user.ID
			fmt.Printf("Found user ID: %s\n", userID)
			return userID, nil
		}
	}

	return "", fmt.Errorf("could not find user '%s' in Keycloak", email)
}

func (client *KeycloakClient) GetAdminUsersURL(realm string) *url.URL {
	path := fmt.Sprintf("./admin/realms/%s/users", realm)
	return client.BaseURL.JoinPath(path)
}

type getUsersResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (client *KeycloakClient) IsRenkuAdmin(ctx context.Context, realm string, userID string) (isAdmin bool, err error) {
	getURL := client.GetAdminRolesURL(realm, userID)

	var result []roleMapping
	_, err = client.GetJSON(ctx, getURL.String(), &result)
	if err != nil {
		return false, err
	}

	for _, role := range result {
		if role.Name == renkuAdminRole {
			return true, nil
		}
	}
	return false, nil
}

func (client *KeycloakClient) findRenkuAdminRole(ctx context.Context, realm string, userID string) (role roleMapping, err error) {
	getURL := client.GetAdminAvailavleRolesURL(realm, userID)

	var result []roleMapping
	_, err = client.GetJSON(ctx, getURL.String(), &result)
	if err != nil {
		return roleMapping{}, err
	}

	for _, roleObj := range result {
		if roleObj.Name == renkuAdminRole {
			role = roleObj
			return role, err
		}
	}
	return roleMapping{}, fmt.Errorf("could not find role '%s' in Keycloak", renkuAdminRole)
}

func (client *KeycloakClient) AddRenkuAdminRoleToUser(ctx context.Context, realm string, userID string) error {
	role, err := client.findRenkuAdminRole(ctx, realm, userID)
	if err != nil {
		return err
	}

	postURL := client.GetAdminRolesURL(realm, userID)

	body := []roleMapping{
		role,
	}

	_, err = client.PostJSON(ctx, postURL.String(), body, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *KeycloakClient) GetAdminRolesURL(realm string, userID string) *url.URL {
	path := fmt.Sprintf("./admin/realms/%s/users/%s/role-mappings/realm", realm, userID)
	return client.BaseURL.JoinPath(path)
}

func (client *KeycloakClient) GetAdminAvailavleRolesURL(realm string, userID string) *url.URL {
	path := fmt.Sprintf("./admin/realms/%s/users/%s/role-mappings/realm/available", realm, userID)
	return client.BaseURL.JoinPath(path)
}

type roleMapping struct {
	ID          string `json:"id,omitempty"`
	ContainerID string `json:"containerId,omitempty"`
	Name        string `json:"name,omitempty"`
}
