package token

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var (
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

		return []byte(t), nil
	})
	if err != nil {
		return nil, err
	}

	err = checkExpired(parsedToken)
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
