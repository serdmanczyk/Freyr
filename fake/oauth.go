package fake

import (
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/oauth2"
	"net/http"
)

type Oauth struct {
	Email string
}

func (f *Oauth) GetRedirectURL(csrftoken string) string {
	return "/someurl?state=" + csrftoken
}

func (f *Oauth) GetCallbackCsrfToken(r *http.Request) string {
	return r.FormValue("state")
}

func (f *Oauth) GetExchangeToken(r *http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (f *Oauth) GetUserData(tok *oauth2.Token) (models.User, error) {
	return models.User{Email: f.Email}, nil
}
