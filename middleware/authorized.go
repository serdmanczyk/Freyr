package middleware

import (
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/oauth"
	"github.com/serdmanczyk/freyr/token"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"time"
)

const (
	apiAuthTypeValue    = "API"
	deviceAuthTypeValue = "DEVICE"
	authTypeHeader      = "X-FREYR-AUTHTYPE"
	tokenHeader         = "X-FREYR-TOKEN"
	authUserHeader      = "X-FREYR-USER"
	apiAuthDateHeader   = "X-FREYR-DATETIME"
	apiSignatureHeader  = "X-FREYR-SIGNATURE"
)

type Authorizer interface {
	Authorize(ctx context.Context, r *http.Request) context.Context
}

func Authorize(auths ...Authorizer) apollo.Constructor {
	return apollo.Constructor(func(next apollo.Handler) apollo.Handler {
		return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			for _, auth := range auths {
				authCtx := auth.Authorize(ctx, r)
				if authCtx != nil {
					next.ServeHTTP(authCtx, w, r)
					return
				}
			}

			http.Error(w, "Request not authorized", http.StatusUnauthorized)
		})
	})
}

type UserAuthorizer struct {
	tokenStore token.JtwTokenGen
}

func NewUserAuthorizer(tS token.JtwTokenGen) *UserAuthorizer {
	return &UserAuthorizer{tokenStore: tS}
}

func (u *UserAuthorizer) Authorize(ctx context.Context, r *http.Request) context.Context {
	cookie, err := r.Cookie(oauth.CookieName)
	if err != nil {
		return nil
	}

	claims, err := u.tokenStore.ValidateToken(cookie.Value)
	if err != nil {
		return nil
	}

	userEmail, ok := claims["email"].(string)
	if !ok {
		return nil
	}

	return context.WithValue(ctx, "email", userEmail)
}

type ApiAuthorizer struct {
	secretStore models.SecretStore
}

func NewApiAuthorizer(ss models.SecretStore) *ApiAuthorizer {
	return &ApiAuthorizer{secretStore: ss}
}

func (a *ApiAuthorizer) Authorize(ctx context.Context, r *http.Request) context.Context {
	authType := r.Header.Get(authTypeHeader)
	if authType != apiAuthTypeValue {
		return nil
	}

	signature := r.Header.Get(apiSignatureHeader)
	if signature == "" {
		return nil
	}

	userEmail, signingString := apiSigningString(r)
	if userEmail == "" {
		return nil
	}

	userSecret, err := a.secretStore.GetSecret(userEmail)
	if err != nil {
		return nil
	}

	if !userSecret.Verify(signingString, signature) {
		return nil
	}

	return context.WithValue(ctx, "email", userEmail)
}

func apiSigningString(r *http.Request) (userEmail string, signinString string) {
	datetime := r.Header.Get(apiAuthDateHeader)
	user := r.Header.Get(authUserHeader)

	if datetime == "" || user == "" {
		return
	}

	timeInt, err := strconv.ParseInt(datetime, 10, 64)
	if err != nil {
		return
	}

	if timeInt < time.Now().Add(time.Second*-5).Unix() || timeInt > time.Now().Add(time.Second*5).Unix() {
		return
	}
	// TODO: Add support for POST (include content-length)

	userEmail, signinString = user, r.Method+r.URL.RawPath+datetime+user
	return
}

type DeviceAuthorizer struct {
	secretStore models.SecretStore
}

func NewDeviceAuthorizer(ss models.SecretStore) *DeviceAuthorizer {
	return &DeviceAuthorizer{secretStore: ss}
}

func (d *DeviceAuthorizer) Authorize(ctx context.Context, r *http.Request) context.Context {
	authType := r.Header.Get(authTypeHeader)
	if authType != deviceAuthTypeValue {
		return nil
	}

	jwtTokenString := r.Header.Get(tokenHeader)
	if jwtTokenString == "" {
		return nil
	}

	requestUserEmail := r.Header.Get(authUserHeader)
	if requestUserEmail == "" {
		return nil
	}

	err := r.ParseForm()
	if err != nil {
		return nil
	}

	requestCoreId := r.PostFormValue("coreid")
	if jwtTokenString == "" {
		return nil
	}

	claims, err := token.ValidateUserToken(d.secretStore, jwtTokenString)
	if err != nil {
		return nil
	}

	tokenCoreId, ok := claims["coreid"].(string)
	if !ok {
		return nil
	}

	tokenUserEmail, ok := claims["email"].(string)
	if !ok {
		return nil
	}

	if tokenCoreId != requestCoreId || tokenUserEmail != requestUserEmail {
		return nil
	}

	return context.WithValue(ctx, "email", requestUserEmail)
}
