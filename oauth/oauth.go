package oauth

import (
	"errors"
	"github.com/serdmanczyk/gardenspark/models"
	"github.com/serdmanczyk/gardenspark/token"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

const (
	CookieName = "_gspk_"
)

var (
	oauthClaim    = map[string]interface{}{"oauthlogin": true}
	InvalidClaims = errors.New("Claims are invalid")
)

type OauthHandler interface {
	GetRedirectURL(csrftoken string) string
	GetCallbackCsrfToken(r *http.Request) string
	GetExchangeToken(r *http.Request) (*oauth2.Token, error)
	GetUserData(tok *oauth2.Token) (*models.User, error)
}

func checkOauthClaim(claims map[string]interface{}) bool {
	if len(claims) != 1 {
		return false
	}

	loginClaim, ok := claims["oauthlogin"]
	if !ok {
		return false
	}

	oauthLogin, ok := loginClaim.(bool)
	if !ok || !oauthLogin {
		return false
	}

	return true
}

func setUserCookie(w http.ResponseWriter, t token.TokenSource, u *models.User) error {
	expiry := time.Now().Add(time.Hour * 744) // ~1 month
	claims := map[string]interface{}{
		"email": u.Email,
	}
	token, err := t.GenerateToken(expiry, claims)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:   CookieName,
		Value:  token,
		MaxAge: int(expiry.Unix() - time.Now().Unix()),
		// Secure:   true,
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

		err = userStore.AddUser(*user)
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
