package fake

import (
	"github.com/serdmanczyk/freyr/models"
)

type SecretStore map[string]models.Secret

func (s SecretStore) GetSecret(userEmail string) (models.Secret, error) {
	secret, ok := s[userEmail]
	if !ok {
		return models.Secret([]byte{}), models.SecretDoesntExist
	}
	return secret, nil
}

func (s SecretStore) StoreSecret(userEmail string, secret models.Secret) error {
	s[userEmail] = secret
	return nil
}
