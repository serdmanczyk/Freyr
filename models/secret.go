package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

var (
	SecretDoesntExist = errors.New("No secret exists for given criterion")
)

type SecretStore interface {
	GetSecret(userEmail string) (Secret, error)
	StoreSecret(userEmail string, secret Secret) error
}

type Secret string

func NewSecret() (Secret, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	str := base64.URLEncoding.EncodeToString(b)

	return Secret(str), nil
}
