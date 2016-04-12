// +build integration

package database

import (
	"github.com/serdmanczyk/gardenspark/models"
	"reflect"
	"testing"
)

func TestAddUser(t *testing.T) {
	testUser := models.User{
		Email:      "jackharkness@torchwood.co.uk",
		Name:       "Jack Harkness",
		GivenName:  "Jack",
		FamilyName: "Harkness",
		Gender:     "male",
		Locale:     "us",
	}

	err := gspkDb.AddUser(testUser)
	if err != nil {
		t.Fatalf("Failed adding user: ", err.Error())
	}

	dbUser, err := gspkDb.GetUser(testUser.Email)
	if err != nil {
		t.Fatalf("Failed getting user: ", err.Error())
	}

	if !reflect.DeepEqual(testUser, dbUser) {
		t.Fatalf("User did not match inserted; got %v expected %v", dbUser, testUser)
	}
}

func TestAddUserTwice(t *testing.T) {
	testUser := models.User{
		Email:      "thedoctor@gallifrey.time",
		Name:       "doctor",
		GivenName:  "doctor",
		FamilyName: "doctor",
		Gender:     "male",
		Locale:     "gall",
	}
	err := gspkDb.AddUser(testUser)
	if err != nil {
		t.Fatalf("Failed adding user: ", err.Error())
	}

	err = gspkDb.AddUser(testUser)
	if err != nil && err != models.UserAlreadyExists {
		t.Fatalf("Incorrect error on double insert; expected %s got %s ", models.UserAlreadyExists, err)
	}
}
