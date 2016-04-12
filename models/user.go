package models

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
)

var (
	UserAlreadyExists = errors.New("Entry already exists for user")
)

type UserStore interface {
	GetUser(email string) (User, error)
	AddUser(User) error
}

type User struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Gender     string `json:"gender"`
	Locale     string `json:"locale"`
}

func UserFromJson(r io.Reader) (*User, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if err := json.Unmarshal(buf, &user); err != nil {
		return nil, err
	}

	return user, nil
}
