// Package oauth defines methods for handling Three-legged Oauth with various
// third-parties (e.g. Google) as well as HTTP handlers.
package oauth

import (
	"errors"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
	"golang.org/x/oauth2"
	"net/http"
	"reflect"
	"time"
)

const (
	// CookieName defines the name of the cookie that web access tokens
	// will be stored as in web requests.
	CookieName = "_freyr_"
)

var (
	oauthClaim = map[string]interface{}{"oauthlogin": true}
	// ErrorInvalidClaims is returned when an oauth requests claims do
	// not match that which the system sets.
	ErrorInvalidClaims = errors.New("Claims are invalid")
)

// Handler is an interface representing types that provide the interface
// to perform Oauth three-legged authorization against a system such as
// Google or Twitter.
type Handler interface {
	GetRedirectURL(csrftoken string) string
	GetCallbackCsrfToken(r *http.Request) string
	GetExchangeToken(r *http.Request) (*oauth2.Token, error)
	GetUserData(tok *oauth2.Token) (models.User, error)
}

func checkOauthClaim(claims map[string]interface{}) bool {
	if !reflect.DeepEqual(claims, oauthClaim) {
		return false
	}

	return true
}

func setUserCookie(w http.ResponseWriter, t token.Source, u models.User) error {
	expiry := time.Now().Add(time.Hour * 744) // ~1 month
	token, err := token.GenerateWebToken(t, expiry, u.Email)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    token,
		MaxAge:   int(expiry.Unix() - time.Now().Unix()),
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	return nil
}

// HandleAuthorize accepts HTTP requests to be authorized and redirects the
// user to the Oauth providers authorization URL.
func HandleAuthorize(o Handler, t token.Source) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiry := time.Now().Add(time.Minute * 5)
		token, err := t.GenerateToken(expiry, oauthClaim)
		if err != nil {
			http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		url := o.GetRedirectURL(token)
		http.Redirect(w, r, url, http.StatusFound)
	})
}

// HandleOAuth2Callback handles verification of Oauth redirects from Oauth
// provider's and ensuring the redirected request is valid and was initiated
// by the system.
func HandleOAuth2Callback(o Handler, t token.Source, userStore models.UserStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfToken := o.GetCallbackCsrfToken(r)
		claims, err := t.ValidateToken(csrfToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if !checkOauthClaim(claims) {
			http.Error(w, ErrorInvalidClaims.Error(), http.StatusForbidden)
			return
		}

		oauthToken, err := o.GetExchangeToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		user, err := o.GetUserData(oauthToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		err = userStore.StoreUser(user)
		if err != nil && err != models.ErrorUserAlreadyExists {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = setUserCookie(w, t, user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

// SetDemoUser accepts an HTTP request and provides a signed a web access JWT
// to allow the current site user to view example data in the demo user's
// account.
func SetDemoUser(t token.Source, userStore models.UserStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user, err := userStore.GetUser("noone@nothing.com")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = setUserCookie(w, t, user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

// LogOut resets the user's JWT web access cookie so their session is no
// longer valid.
func LogOut() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{
			Name:     CookieName,
			Value:    "",
			MaxAge:   int(time.Unix(0, 0).Unix()),
			Secure:   true,
			HttpOnly: true,
		}

		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)
	})
}
