package database

import (
	pq "github.com/lib/pq"
	"github.com/serdmanczyk/freyr/models"
)

// GetUser gets information from the database pertaining to the specified user.
func (db DB) GetUser(email string) (models.User, error) {
	var user models.User

	row := db.QueryRow("select email, full_name, family_name, given_name, gender, locale from users where email = $1;", email)
	err := row.Scan(&user.Email, &user.Name, &user.FamilyName, &user.GivenName, &user.Gender, &user.Locale)

	return user, err
}

// StoreUser inserts the specified user's information in the database.
func (db DB) StoreUser(user models.User) error {
	_, err := db.Exec("insert into users (email, full_name, family_name, given_name, gender, locale) values ($1, $2, $3, $4, $5, $6);", user.Email, user.Name, user.FamilyName, user.GivenName, user.Gender, user.Locale)

	if err, ok := err.(*pq.Error); ok {
		if err.Code == "23505" { // unique_violation
			return models.ErrorUserAlreadyExists
		}
	}

	return err
}
