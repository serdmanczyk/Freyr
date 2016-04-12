package middleware

import (
	"errors"
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/gardenspark/oauth"
	"github.com/serdmanczyk/gardenspark/token"
	"golang.org/x/net/context"
	"net/http"
)

var (
	CookieNotFound = errors.New("Token cookie not found")
)

func checkCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(oauth.CookieName)
	if err != nil {
		return "", CookieNotFound
	}

	return cookie.Value, nil
}

// TODO: implement auth header for API use
func checkAuthHeader(r *http.Request) (string, error) {
	return "", errors.New("not yet implemented")
}

func getEmail(claims map[string]interface{}) string {
	email := claims["email"]
	emailString, _ := email.(string)
	return emailString
}

func Authorized(t token.TokenSource) apollo.Constructor {
	return apollo.Constructor(func(next apollo.Handler) apollo.Handler {
		return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			tokenString, err := checkCookie(r)
			// if err == CookieNotFound {
			// 	tokenString, err = checkAuthHeader(r)
			// }

			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			claims, err := t.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			email := getEmail(claims)
			if email == "" {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			userCtx := context.WithValue(ctx, "email", email)
			next.ServeHTTP(userCtx, w, r)
		})
	})
}
