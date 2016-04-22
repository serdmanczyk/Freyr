// +build integration

package database

import (
	"os"
	"testing"
)

var db DB

func TestMain(m *testing.M) {
	ldb, err := DbConn("postgres", "testuser", "testpassword")
	if err != nil {
		panic("Error creating database connection: " + err.Error())
	}

	db = ldb
	_, err = db.Exec("TRUNCATE users, readings")
	if err != nil {
		panic("Coudn't connect to table! " + err.Error())
	}

	os.Exit(m.Run())
}
