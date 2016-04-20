package token

import (
	"encoding/base64"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/serdmanczyk/gardenspark/models"
	"time"
)

var (
	nilClaims        = Claims{}
	InvalidToken     = errors.New("Invalid Token")
	TokenExpired     = errors.New("Token has expired")
	InvalidAlgorithm = errors.New("Token signed with invalid algorithm")
)

type Claims map[string]interface{}

type TokenSource interface {
	GenerateToken(exp time.Time, claims Claims) (string, error)
	ValidateToken(string) (Claims, error)
}

type JtwTokenGen []byte

func (t JtwTokenGen) GenerateToken(exp time.Time, claims Claims) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	for k, v := range claims {
		token.Claims[k] = v
	}
	token.Claims["exp"] = exp.Format(time.RFC3339)

	return token.SignedString([]byte(t))
}

func (t JtwTokenGen) ValidateToken(tokenString string) (Claims, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, InvalidAlgorithm
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
		return InvalidToken
	}

	timeExpires, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		return InvalidToken
	}

	if timeExpires.Unix() < time.Now().Unix() {
		return TokenExpired
	}

	return nil
}

func GenerateUserToken(secret models.Secret, exp time.Time, userEmail string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = map[string]interface{}{
		"user": userEmail,
		"exp":  exp.Format(time.RFC3339),
	}

	bytes, err := base64.URLEncoding.DecodeString(string(secret))
	if err != nil {
		return "", nil
	}

	return token.SignedString([]byte(bytes))
}

func ValidateUserToken(store models.SecretStore, jwtTokenString string) (Claims, error) {
	parsedToken, err := jwt.Parse(jwtTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, InvalidAlgorithm
		}

		err := checkExpired(token)
		if err != nil {
			return nil, err
		}

		email, ok := token.Claims["email"].(string)
		if !ok {
			return nil, InvalidToken
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
