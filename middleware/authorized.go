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
	ApiAuthTypeValue    = "API"
	DeviceAuthTypeValue = "DEVICE"
	AuthTypeHeader      = "X-FREYR-AUTHTYPE"
	TokenHeader         = "X-FREYR-TOKEN"
	AuthUserHeader      = "X-FREYR-USER"
	ApiAuthDateHeader   = "X-FREYR-DATETIME"
	ApiSignatureHeader  = "X-FREYR-SIGNATURE"
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

type WebAuthorizer struct {
	tokenStore token.JtwTokenGen
}

func NewWebAuthorizer(tS token.JtwTokenGen) *WebAuthorizer {
	return &WebAuthorizer{tokenStore: tS}
}

func (u *WebAuthorizer) Authorize(ctx context.Context, r *http.Request) context.Context {
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
	authType := r.Header.Get(AuthTypeHeader)
	if authType != ApiAuthTypeValue {
		return nil
	}

	signature := r.Header.Get(ApiSignatureHeader)
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
	datetime := r.Header.Get(ApiAuthDateHeader)
	user := r.Header.Get(AuthUserHeader)

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

	userEmail = user
	if r.Method == "POST" {
		signinString = r.Method+r.URL.RawPath+datetime+user+strconv.FormatInt(r.ContentLength, 10)
	} else {
		signinString = r.Method+r.URL.RawPath+datetime+user
	}

	return
}

func SignRequest(s models.Secret, userEmail string, r *http.Request) {
	r.Header.Add(AuthTypeHeader, ApiAuthTypeValue)
	r.Header.Add(AuthUserHeader, userEmail)
	n := time.Now().Unix()
	unixStamp := strconv.FormatInt(n, 10)
	r.Header.Add(ApiAuthDateHeader, unixStamp)

	_, signingString := apiSigningString(r)

	signature := s.Sign(signingString)
	r.Header.Add(ApiSignatureHeader, signature)
}

type DeviceAuthorizer struct {
	secretStore models.SecretStore
}

func NewDeviceAuthorizer(ss models.SecretStore) *DeviceAuthorizer {
	return &DeviceAuthorizer{secretStore: ss}
}

func (d *DeviceAuthorizer) Authorize(ctx context.Context, r *http.Request) context.Context {
	authType := r.Header.Get(AuthTypeHeader)
	if authType != DeviceAuthTypeValue {
		return nil
	}

	jwtTokenString := r.Header.Get(TokenHeader)
	if jwtTokenString == "" {
		return nil
	}

	requestUserEmail := r.Header.Get(AuthUserHeader)
	if requestUserEmail == "" {
		return nil
	}

	requestCoreId := r.PostFormValue("coreid")
	if requestCoreId == "" {
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
