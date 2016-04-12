package database

import (
	"database/sql"
	"github.com/serdmanczyk/gardenspark/models"
)

func (db GspkDb) GetUserSecret(userEmail string) (models.Secret, error) {
	var secret string

	err := db.QueryRow("select secret from users where email = $1;", userEmail).Scan(&secret)
	if err == sql.ErrNoRows || secret == "" {
		return models.Secret(""), models.SecretDoesntExist
	}

	return models.Secret(secret), err
}

func (db GspkDb) SetUserSecret(userEmail string, secret models.Secret) error {
	_, err := db.Exec("update users set secret = $1 where email = $2;", string(secret), userEmail)

	return err
}
