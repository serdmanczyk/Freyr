// +build integration
package database

import (
	"os"
	"testing"
)

var gspkDb GspkDb

func TestMain(m *testing.M) {
	db, err := DbConn("testgspkpostgres", "testuser", "testpassword")
	if err != nil {
		panic("Error creating database connection: " + err.Error())
	}

	_, err = db.Exec("TRUNCATE users")
	if err != nil {
		panic("Coudn't connect to table! " + err.Error())
	}

	gspkDb = db
	os.Exit(m.Run())
}
