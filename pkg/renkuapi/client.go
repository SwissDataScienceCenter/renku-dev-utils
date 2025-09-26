package renkuapi

import (
	"context"
	"net/http"
	"net/url"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/renkuapi/session"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/renkuapi/users"
)

type RenkuApiClient struct {
	baseURL *url.URL

	auth *RenkuApiAuth

	rsc *session.RenkuSessionClient
	ruc *users.RenkuUsersClient

	httpClient *http.Client
}

func NewRenkuApiClient(baseURL string) (rac *RenkuApiClient, err error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.EscapedPath() == "/" {
		parsedURL.Path = ""
	}
	rac = &RenkuApiClient{
		baseURL: parsedURL,
	}
	if rac.httpClient == nil {
		rac.httpClient = http.DefaultClient
	}

	// initialize auth
	auth, err := NewRenkuApiAuth(baseURL)
	if err != nil {
		return nil, err
	}
	rac.auth = auth

	return rac, nil
}

func (rac *RenkuApiClient) Auth() *RenkuApiAuth {
	return rac.auth
}

func (rac *RenkuApiClient) IsLoggedIn(ctx context.Context) bool {
	token, _ := rac.auth.GetAccessToken(ctx)
	return token != ""
}

func (rac *RenkuApiClient) IsAdmin(ctx context.Context) bool {
	ruc, err := rac.Users()
	if err != nil {
		return false
	}
	userInfo, err := ruc.GetUser(ctx)
	if err != nil {
		return false
	}
	return userInfo.IsAdmin
}

func (rac *RenkuApiClient) Session() (rsc *session.RenkuSessionClient, err error) {
	if rac.rsc == nil {
		rsc, err = session.NewRenkuSessionClient(rac.baseURL.String(), session.WithRequestEditors(session.RequestEditorFn(rac.auth.RequestEditor())))
		if err != nil {
			return nil, err
		}
		rac.rsc = rsc
	}
	return rac.rsc, nil
}

func (rac *RenkuApiClient) Users() (ruc *users.RenkuUsersClient, err error) {
	if rac.ruc == nil {
		ruc, err = users.NewRenkuUsersClient(rac.baseURL.String(), users.WithRequestEditors(users.RequestEditorFn(rac.auth.RequestEditor())))
		if err != nil {
			return nil, err
		}
		rac.ruc = ruc
	}
	return rac.ruc, nil
}
