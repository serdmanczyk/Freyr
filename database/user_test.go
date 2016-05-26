// +build integration

package database

import (
	"github.com/serdmanczyk/freyr/models"
	"reflect"
	"testing"
)

func TestStoreUser(t *testing.T) {
	testUser := models.User{
		Email:      "jackharkness@torchwood.co.uk",
		Name:       "Jack Harkness",
		GivenName:  "Jack",
		FamilyName: "Harkness",
		Gender:     "male",
		Locale:     "us",
	}

	err := db.StoreUser(testUser)
	if err != nil {
		t.Fatalf("Failed adding user: %s", err.Error())
	}

	dbUser, err := db.GetUser(testUser.Email)
	if err != nil {
		t.Fatalf("Failed getting user: %s", err.Error())
	}

	if !reflect.DeepEqual(testUser, dbUser) {
		t.Fatalf("User did not match inserted; got %v expected %v", dbUser, testUser)
	}
}

func TestStoreUserTwice(t *testing.T) {
	testUser := models.User{
		Email:      "thedoctor@gallifrey.time",
		Name:       "doctor",
		GivenName:  "doctor",
		FamilyName: "doctor",
		Gender:     "male",
		Locale:     "gall",
	}
	err := db.StoreUser(testUser)
	if err != nil {
		t.Fatalf("Failed adding user: %s", err.Error())
	}

	err = db.StoreUser(testUser)
	if err != nil && err != models.ErrorUserAlreadyExists {
		t.Fatalf("Incorrect error on double insert; expected %s got %s ", models.ErrorUserAlreadyExists, err)
	}
}
