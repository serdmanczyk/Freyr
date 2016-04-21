package database

import (
	"database/sql"
	"github.com/serdmanczyk/freyr/models"
)

func (db DB) GetSecret(userEmail string) (secret models.Secret, err error) {
	var secretString string

	secret = models.Secret([]byte{})

	err = db.QueryRow("select secret from users where email = $1;", userEmail).Scan(&secretString)
	if err == sql.ErrNoRows || secretString == "" {
		err = models.SecretDoesntExist
		return
	}

	secret, err = models.SecretFromBase64(secretString)
	return
}

func (db DB) StoreSecret(userEmail string, secret models.Secret) error {
	_, err := db.Exec("update users set secret = $1 where email = $2;", secret.Encode(), userEmail)

	return err
}
