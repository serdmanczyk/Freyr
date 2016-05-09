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
	CookieName = "_freyr_"
)

var (
	oauthClaim    = map[string]interface{}{"oauthlogin": true}
	InvalidClaims = errors.New("Claims are invalid")
)

type OauthHandler interface {
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

func setUserCookie(w http.ResponseWriter, t token.TokenSource, u models.User) error {
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

func HandleAuthorize(o OauthHandler, t token.TokenSource) http.Handler {
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

func HandleOAuth2Callback(o OauthHandler, t token.TokenSource, userStore models.UserStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfToken := o.GetCallbackCsrfToken(r)
		claims, err := t.ValidateToken(csrfToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if !checkOauthClaim(claims) {
			http.Error(w, InvalidClaims.Error(), http.StatusForbidden)
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
		if err != nil && err != models.UserAlreadyExists {
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

func SetDemoUser(t token.TokenSource, userStore models.UserStore) http.Handler {
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
