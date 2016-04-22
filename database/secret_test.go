// +build integration

package database

import (
	"github.com/serdmanczyk/freyr/models"
	"testing"
)

func TestGetSecretDoesntExist(t *testing.T) {
	testUser := models.User{
		Email:      "kilgoretrout@stagman.com",
		Name:       "Kilgore Trout",
		GivenName:  "Kilgore",
		FamilyName: "Trout",
		Gender:     "male",
		Locale:     "us",
	}

	_, err := db.GetSecret(testUser.Email)
	if err.Error() != models.SecretDoesntExist.Error() {
		t.Errorf("Unknown error retreiving secret for non-existent user: %s", err.Error())
	}

	err = db.StoreUser(testUser)
	if err != nil {
		t.Fatalf("Failed adding user: ", err.Error())
	}

	_, err = db.GetSecret(testUser.Email)
	if err == nil {
		t.Errorf("No error reported retreiving un-set secret for new user: should be %s", models.SecretDoesntExist.Error())
	}

	if err != models.SecretDoesntExist {
		t.Errorf("Unknown error retreiving un-set secret for new user: %s", err.Error())
	}
}

func TestSetGetSecret(t *testing.T) {
	testUser := models.User{
		Email:      "billypilgrim@us.army.mil",
		Name:       "Billy Pilgrim",
		GivenName:  "Billy",
		FamilyName: "Pilgrim",
		Gender:     "male",
		Locale:     "us",
	}

	err := db.StoreUser(testUser)
	if err != nil {
		t.Fatalf("Failed adding user: ", err.Error())
	}

	secret, err := models.NewSecret()
	if err != nil {
		t.Errorf("Error generating new secret: %s", err.Error())
	}

	err = db.StoreSecret(testUser.Email, secret)
	if err != nil {
		t.Errorf("Error setting secret: %s", err.Error())
	}

	dbSecret, err := db.GetSecret(testUser.Email)
	if err != nil {
		t.Errorf("Error getting secret: %s", err.Error())
	}

	if dbSecret.Encode() != secret.Encode() {
		t.Errorf("Secret from database doesn't match; expected %s, got: %s", secret, dbSecret)
	}
}
