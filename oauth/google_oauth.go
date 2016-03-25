package oauth

import (
	"github.com/serdmanczyk/gardenspark/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
)

const (
	profileInfoURL  = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"
	user_info_scope = "https://www.googleapis.com/auth/userinfo.email"
)

type GoogleOauth struct {
	Config *oauth2.Config
}

func NewGoogleOauth(clientId, clientSecret, domain string) *GoogleOauth {
	return &GoogleOauth{
		Config: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  "https://" + domain + "/api/oauth2callback",
			Scopes: []string{
				user_info_scope,
			},
		},
	}
}

func (g GoogleOauth) GetRedirectURL(csrfToken string) string {
	return g.Config.AuthCodeURL(csrfToken)
}

func (g GoogleOauth) GetCallbackCsrfToken(r *http.Request) string {
	return r.FormValue("state")
}

func (g GoogleOauth) GetExchangeToken(r *http.Request) (*oauth2.Token, error) {
	tok, err := g.Config.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func (g GoogleOauth) GetUserData(tok *oauth2.Token) (*models.User, error) {
	client := g.Config.Client(oauth2.NoContext, tok)
	resp, err := client.Get(profileInfoURL)
	if err != nil {
		return nil, err
	}

	user, err := models.UserFromJson(resp.Body)
	if err != nil {
		return nil, err
	}

	return user, nil
}
