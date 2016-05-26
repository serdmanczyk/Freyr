// Package middleware defines HTTP middleware such as those used for user
// validation.
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

// Constant definitions for authorization type headers and
// header values.
const (
	APIAuthTypeValue    = "API"
	DeviceAuthTypeValue = "DEVICE"
	AuthTypeHeader      = "X-FREYR-AUTHTYPE"
	TokenHeader         = "X-FREYR-TOKEN"
	AuthUserHeader      = "X-FREYR-USER"
	APIAuthDateHeader   = "X-FREYR-DATETIME"
	APISignatureHeader  = "X-FREYR-SIGNATURE"
)

// Authorizer is an interface that represents types capable of validating
// if an HTTP request is authorized, and returning a context indicating the
// authorized user
type Authorizer interface {
	Authorize(ctx context.Context, r *http.Request) context.Context
}

// Authorize returns a piece of middleware that will verify if the request
// is authorized by and of the Authorizers passed in before calling subsequent
// handlers.
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

// WebAuthorizer is used to verify if requests are signed in a manner
// expected users accessing the api via a web browser.
type WebAuthorizer struct {
	tokenStore token.JWTTokenGen
}

// NewWebAuthorizer generates a new *WebAuthorizer
func NewWebAuthorizer(tS token.JWTTokenGen) *WebAuthorizer {
	return &WebAuthorizer{tokenStore: tS}
}

// Authorize validates the request contains a cookie placed by the system's
// oauth handler and that the cookie has a valid signature signed by the
// master token.JtwTokenGen
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

// APIAuthorizer is a type used to validate requests were signed in
// the manner expected of API style requests.
type APIAuthorizer struct {
	secretStore models.SecretStore
}

// NewAPIAuthorizer returns a new APIAuthorizer
func NewAPIAuthorizer(ss models.SecretStore) *APIAuthorizer {
	return &APIAuthorizer{secretStore: ss}
}

// Authorize validates that a request has been signed with a user's secret
func (a *APIAuthorizer) Authorize(ctx context.Context, r *http.Request) context.Context {
	authType := r.Header.Get(AuthTypeHeader)
	if authType != APIAuthTypeValue {
		return nil
	}

	signature := r.Header.Get(APISignatureHeader)
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
	datetime := r.Header.Get(APIAuthDateHeader)
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
		signinString = r.Method + r.URL.RawPath + datetime + user + strconv.FormatInt(r.ContentLength, 10)
	} else {
		signinString = r.Method + r.URL.RawPath + datetime + user
	}

	return
}

// SignRequest builds a signing-string from the request's content and signs
// it with the given secret.
func SignRequest(s models.Secret, userEmail string, r *http.Request) {
	r.Header.Add(AuthTypeHeader, APIAuthTypeValue)
	r.Header.Add(AuthUserHeader, userEmail)
	n := time.Now().Unix()
	unixStamp := strconv.FormatInt(n, 10)
	r.Header.Add(APIAuthDateHeader, unixStamp)

	_, signingString := apiSigningString(r)

	signature := s.Sign(signingString)
	r.Header.Add(APISignatureHeader, signature)
}

// DeviceAuthorizer is a type used to verify that a request was signed in the
// manner specified for requests from a device.
type DeviceAuthorizer struct {
	secretStore models.SecretStore
}

// NewDeviceAuthorizer returns a new *DeviceAuthorizer
func NewDeviceAuthorizer(ss models.SecretStore) *DeviceAuthorizer {
	return &DeviceAuthorizer{secretStore: ss}
}

// Authorize validates a a valid JWT signature header is present signed by
// the user's secret and that the content in the signature matches the headers
// describing the user and core on who's behalf the request was made.
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

	requestCoreID := r.PostFormValue("coreid")
	if requestCoreID == "" {
		return nil
	}

	claims, err := token.ValidateUserToken(d.secretStore, jwtTokenString)
	if err != nil {
		return nil
	}

	tokenCoreID, ok := claims["coreid"].(string)
	if !ok {
		return nil
	}

	tokenUserEmail, ok := claims["email"].(string)
	if !ok {
		return nil
	}

	if tokenCoreID != requestCoreID || tokenUserEmail != requestUserEmail {
		return nil
	}

	return context.WithValue(ctx, "email", requestUserEmail)
}
