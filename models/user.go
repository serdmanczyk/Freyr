package models

import (
	"errors"
)

var (
	// ErrorUserAlreadyExists is returned from a UserStore when a user is
	// attempted to be stored but already exists in the store.
	ErrorUserAlreadyExists = errors.New("Entry already exists for user")
)

// UserStore is an interface represeting types that can store and retrieve user descriptions
type UserStore interface {
	GetUser(email string) (User, error)
	StoreUser(User) error
}

// User is a struct represnting a distinct user of the application and basic
// data describing that user.
type User struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Gender     string `json:"gender"`
	Locale     string `json:"locale"`
}
