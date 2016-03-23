package database

import (
	"database/sql"
	"fmt"
	pq "github.com/lib/pq"
	"github.com/serdmanczyk/gardenspark/models"
)

type GspkDb struct {
	*sql.DB
}

func DbConn(host, user, password string) (*GspkDb, error) {
	connstr := fmt.Sprintf("host=%s user=%s password=%s sslmode=disable", host, user, password)

	db, err := sql.Open("postgres", connstr)
	if err != nil {
		return nil, err
	}

	return &GspkDb{db}, nil
}

func (db *GspkDb) GetUser(email string) (models.User, error) {
	var user models.User

	row := db.QueryRow("select email, full_name, family_name, given_name, gender, locale from users where email = $1;", email)
	err := row.Scan(&user.Email, &user.Name, &user.FamilyName, &user.GivenName, &user.Gender, &user.Locale)

	return user, err
}

func (db *GspkDb) AddUser(user models.User) error {
	_, err := db.Exec("insert into users (email, full_name, family_name, given_name, gender, locale) values ($1, $2, $3, $4, $5, $6);", user.Email, user.Name, user.FamilyName, user.GivenName, user.Gender, user.Locale)

	if err, ok := err.(*pq.Error); ok {
		if err.Code == "23505" { // unique_violation
			return models.UserAlreadyExists
		}
	}

	return err
}

func (db *GspkDb) GetUserSecret(email string) (string, error) {
	var secret string

	err := db.QueryRow("select secret from users where email = $1;", email).Scan(&secret)

	return secret, err
}

func (db *GspkDb) SetUserSecret(email, secret string) error {
	_, err := db.Exec("update users set secret = $1 where email = $2;", secret, email)

	return err
}
