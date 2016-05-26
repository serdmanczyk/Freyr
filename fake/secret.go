package fake

import (
	"github.com/serdmanczyk/freyr/models"
)

// SecretStore implements the models.SecretStore interface for use in unit tests
// of libraries that accept a models.SecretStore.  Implemented via an in memory map.
type SecretStore map[string]models.Secret

// GetSecret returns the secret for given userEmail.
func (s SecretStore) GetSecret(userEmail string) (models.Secret, error) {
	secret, ok := s[userEmail]
	if !ok {
		return models.Secret([]byte{}), models.ErrorSecretDoesntExist
	}
	return secret, nil
}

// StoreSecret updates/inserts a secret for the given userEmail.
func (s SecretStore) StoreSecret(userEmail string, secret models.Secret) error {
	s[userEmail] = secret
	return nil
}
