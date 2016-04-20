package models

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var (
	nullSecret        = Secret([]byte{})
	SecretDoesntExist = errors.New("No secret exists for given criterion")
)

type SecretStore interface {
	GetSecret(userEmail string) (Secret, error)
	StoreSecret(userEmail string, secret Secret) error
}

type Secret []byte

func SecretFromBase64(base64String string) (Secret, error) {
	decoded, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return nullSecret, err
	}

	return decoded, nil
}

func (s Secret) Encode() string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

func (s Secret) Sign(input string) string {
	mac := hmac.New(sha256.New, []byte(s))
	mac.Write([]byte(input))
	signed := mac.Sum(nil)
	return base64.URLEncoding.EncodeToString(signed)
}

func (s Secret) Verify(input, signature string) bool {
	if input == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s))
	mac.Write([]byte(input))
	expected := mac.Sum(nil)
	decoded, err := base64.URLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	// Listen to your mother and do constant time comparison
	return hmac.Equal(decoded, expected)
}

func NewSecret() (Secret, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return nullSecret, err
	}

	return Secret(b), nil
}
