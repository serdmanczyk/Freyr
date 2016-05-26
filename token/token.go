// Package token defines methods and types for using and validating Freyr
// tokens in JWT format.
package token

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/serdmanczyk/freyr/models"
	"time"
)

var (
	nilClaims = Claims{}
)

// ValidationError is used for errors related to validating tokens.
type ValidationError error

// TODO: instead of exporting errors for comparison, refactor to better method.

// Exported ValidationError 'constants'
var (
	ErrorInvalidToken     ValidationError = errors.New("Invalid Token")
	ErrorTokenExpired                     = errors.New("Token has expired")
	ErrorInvalidAlgorithm                 = errors.New("Token signed with invalid algorithm")
)

// Claims is a type used to set claim values in a JWT.
type Claims map[string]interface{}

// Source is an interface representing types that can generate and validate JWT tokens.
// Typically a TokenSource holds a secret for signing or has access to keyed secrets
// e.g. user secrets.
type Source interface {
	GenerateToken(exp time.Time, claims Claims) (string, error)
	ValidateToken(string) (Claims, error)
}

// JWTTokenGen is a type used for generating signed JSON Web Tokens.
type JWTTokenGen []byte

// GenerateToken returns a signed JSON Web Token with the set claims and
// expiration date.
func (t JWTTokenGen) GenerateToken(exp time.Time, claims Claims) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	for k, v := range claims {
		token.Claims[k] = v
	}
	token.Claims["exp"] = exp.Format(time.RFC3339)

	return token.SignedString([]byte(t))
}

// ValidateToken verifies if a string is a valid JSON Web token, and was
// signed with the key used the TokenGen's key.  If valid, it returns the
// token's claims.
func (t JWTTokenGen) ValidateToken(tokenString string) (Claims, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrorInvalidAlgorithm
		}

		err := checkExpired(token)
		if err != nil {
			return nil, err
		}

		return []byte(t), nil
	})
	if err != nil {
		return nil, err
	}

	delete(parsedToken.Claims, "exp")

	return parsedToken.Claims, nil
}

func checkExpired(t *jwt.Token) error {
	expiration, ok := t.Claims["exp"].(string)
	if !ok {
		return ErrorInvalidToken
	}

	timeExpires, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		return ErrorInvalidToken
	}

	if timeExpires.Unix() < time.Now().Unix() {
		return ErrorTokenExpired
	}

	return nil
}

// GenerateWebToken generates a JWT to be used as a session token by a user
// accessing the API via a web browser.
func GenerateWebToken(t Source, exp time.Time, userEmail string) (string, error) {
	return t.GenerateToken(exp, Claims{
		"email": userEmail,
		"exp":   exp.Format(time.RFC3339),
	})
}

// GenerateDeviceToken generates a JWT to be used by a Spark webhook
// registered with a core sending readings.
func GenerateDeviceToken(t Source, exp time.Time, coreid, userEmail string) (string, error) {
	return t.GenerateToken(exp, Claims{
		"email":  userEmail,
		"coreid": coreid,
		"exp":    exp.Format(time.RFC3339),
	})
}

// ValidateUserToken validates a JWT signed with a particular user's secret.
func ValidateUserToken(store models.SecretStore, jwtTokenString string) (Claims, error) {
	parsedToken, err := jwt.Parse(jwtTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrorInvalidAlgorithm
		}

		err := checkExpired(token)
		if err != nil {
			return nil, err
		}

		email, ok := token.Claims["email"].(string)
		if !ok {
			return nil, ErrorInvalidToken
		}

		secret, err := store.GetSecret(email)
		if err != nil {
			return nil, err
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nilClaims, err
	}

	return parsedToken.Claims, nil
}
