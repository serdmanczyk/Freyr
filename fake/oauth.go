package fake

import (
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/oauth2"
	"net/http"
)

// Oauth specifies a fake oauth handler for use in unit tests of higher
// level libraries.
type Oauth struct {
	Email string
}

// GetRedirectURL is a fake implementation of oauth GetRedirectURL
func (f *Oauth) GetRedirectURL(csrftoken string) string {
	return "/someurl?state=" + csrftoken
}

// GetCallbackCsrfToken is a fake implementation of oauth GetCallbackCsrfToken
func (f *Oauth) GetCallbackCsrfToken(r *http.Request) string {
	return r.FormValue("state")
}

// GetExchangeToken is a fake implementation of oauth GetExchangeToken
func (f *Oauth) GetExchangeToken(r *http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

// GetUserData is a fake implementation of oauth GetUserData
func (f *Oauth) GetUserData(tok *oauth2.Token) (models.User, error) {
	return models.User{Email: f.Email}, nil
}
