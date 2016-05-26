package models

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var (
	nullSecret = Secret([]byte{})
	// ErrorSecretDoesntExist is returned when a SecretStore doesn't
	// find a requested secret
	ErrorSecretDoesntExist = errors.New("No secret exists for given criterion")
)

// SecretStore is an interface for any type that can store and retrieve user
// secrets.
type SecretStore interface {
	GetSecret(userEmail string) (Secret, error)
	StoreSecret(userEmail string, secret Secret) error
}

// Secret defines a random value that is used for signing content to verify
// ownership.
type Secret []byte

// SecretFromBase64 is a convenience method for decoding a secret encoded as
// a base64 string.
func SecretFromBase64(base64String string) (Secret, error) {
	decoded, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return nullSecret, err
	}

	return decoded, nil
}

// Encode returns the secret encoded as a base64 string
func (s Secret) Encode() string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// Sign returns a base64 encoded signature of the input string using the
// secret and HMAC 256 scheme.
func (s Secret) Sign(input string) string {
	mac := hmac.New(sha256.New, []byte(s))
	mac.Write([]byte(input))
	signed := mac.Sum(nil)
	return base64.URLEncoding.EncodeToString(signed)
}

// Verify takes the input unsigned string and signature and validates the
// signature was generated using this secret
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

// NewSecret creates a new 32 byte (256 bit) secret using crypto/rand
func NewSecret() (Secret, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return nullSecret, err
	}

	return Secret(b), nil
}
