package oauth

import (
	"encoding/json"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
)

const (
	profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"
	userInfoScope  = "https://www.googleapis.com/auth/userinfo.email"
)

// GoogleOauth defines the OauthHandler interface for use against Google's
// Oauth api.
type GoogleOauth struct {
	Config *oauth2.Config
}

// NewGoogleOauth is a convenience method for generating a GoogleOauth type
// with the oauth2.Config initialized.
func NewGoogleOauth(clientID, clientSecret, domain string) *GoogleOauth {
	return &GoogleOauth{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  "https://" + domain + "/api/oauth2callback",
			Scopes: []string{
				userInfoScope,
			},
		},
	}
}

// GetRedirectURL generates the redirect URL with the provided csrf token
func (g GoogleOauth) GetRedirectURL(csrfToken string) string {
	return g.Config.AuthCodeURL(csrfToken)
}

// GetCallbackCsrfToken extracts the csrf token from the Oauth redirect url
func (g GoogleOauth) GetCallbackCsrfToken(r *http.Request) string {
	return r.FormValue("state")
}

// GetExchangeToken extracts the exchange token sent from Google authorizing the system
// to make requests of a user's basic data.
func (g GoogleOauth) GetExchangeToken(r *http.Request) (*oauth2.Token, error) {
	tok, err := g.Config.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		return nil, err
	}

	return tok, nil
}

// GetUserData makes a request of the user's data using the provided access token.
func (g GoogleOauth) GetUserData(tok *oauth2.Token) (user models.User, err error) {
	client := g.Config.Client(oauth2.NoContext, tok)
	resp, err := client.Get(profileInfoURL)
	if err != nil {
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return
	}

	return
}
