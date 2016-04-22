// +build integration

package database

import (
	"github.com/serdmanczyk/freyr/models"
	"testing"
	"time"
)

func TestReadings(t *testing.T) {
	userEmail := "odin@asgard.unv"
	coreId := "123123123142"

	testUser := models.User{
		Email: userEmail,
	}

	err := db.StoreUser(testUser)
	if err != nil {
		t.Fatal(err)
	}

	reading := models.Reading{
		UserEmail:   userEmail,
		CoreId:      coreId,
		Posted:      time.Now(),
		Temperature: 30.0,
		Humidity:    50.0,
		Moisture:    30.0,
		Light:       90.2,
		Battery:     70.0,
	}

	err = db.StoreReading(reading)
	if err != nil {
		t.Fatal(err)
	}

	ref := time.Now()
	start, end := ref.Add(-time.Second), ref.Add(time.Second)

	readings, err := db.GetReadings(coreId, start, end)
	if err != nil {
		t.Fatal(err)
	}

	if len(readings) > 1 {
		t.Fatalf("Expected only one reading return, got %d", len(readings))
	}

	retreading := readings[0]

	if retreading.UserEmail != reading.UserEmail {
		t.Fatalf("Incorrect user email returned; expected %s, got %s", retreading.UserEmail, reading.UserEmail)
	}

	if retreading.CoreId != reading.CoreId {
		t.Fatalf("Incorrect coreid returned; expected %s, got %s", retreading.CoreId, reading.CoreId)
	}

	if retreading.Posted.Unix() != reading.Posted.Unix() {
		t.Fatalf("Incorrect posted time returned; expected %s, got %s", retreading.Posted, reading.Posted)
	}
}

func TestDBGetLatest(t *testing.T) {
	userEmail := "thor@asgard.unv"
	coreOneId := "123123123142"
	coreTwoId := "456456456456"

	testUser := models.User{
		Email: userEmail,
	}

	err := db.StoreUser(testUser)
	if err != nil {
		t.Fatal(err)
	}

	for _, reading := range []models.Reading{
		{
			UserEmail:   userEmail,
			CoreId:      coreOneId,
			Posted:      time.Unix(1461307493, 0),
			Temperature: 30.0,
			Humidity:    50.0,
			Moisture:    30.0,
			Light:       90.2,
			Battery:     70.0,
		},
		{
			UserEmail:   userEmail,
			CoreId:      coreOneId,
			Posted:      time.Unix(1461307499, 0),
			Temperature: 30.0,
			Humidity:    50.0,
			Moisture:    30.0,
			Light:       90.2,
			Battery:     70.0,
		},
		{
			UserEmail:   userEmail,
			CoreId:      coreOneId,
			Posted:      time.Unix(1461307503, 0),
			Temperature: 30.0,
			Humidity:    51.0,
			Moisture:    33.0,
			Light:       90.2,
			Battery:     70.0,
		},
		{
			UserEmail:   userEmail,
			CoreId:      coreTwoId,
			Posted:      time.Unix(1461307493, 0),
			Temperature: 30.0,
			Humidity:    50.0,
			Moisture:    30.0,
			Light:       90.2,
			Battery:     70.0,
		},
		{
			UserEmail:   userEmail,
			CoreId:      coreTwoId,
			Posted:      time.Unix(1461307499, 0),
			Temperature: 30.0,
			Humidity:    50.0,
			Moisture:    30.0,
			Light:       90.2,
			Battery:     70.0,
		},
		{
			UserEmail:   userEmail,
			CoreId:      coreTwoId,
			Posted:      time.Unix(1461307503, 0),
			Temperature: 30.0,
			Humidity:    51.0,
			Moisture:    33.0,
			Light:       90.4,
			Battery:     70.0,
		},
	} {
		err = db.StoreReading(reading)
		if err != nil {
			t.Fatal(err)
		}
	}

	readings, err := db.GetLatestReadings(userEmail)
	if err != nil {
		t.Fatal(err)
	}

	if len(readings) != 2 {
		t.Fatalf("Should have only 2 max readings, got %d", len(readings))
	}

	for _, reading := range readings {
		if reading.UserEmail != userEmail {
			t.Fatalf("Latest reading for user contains incorrect email; expected %s got %s", userEmail, reading.UserEmail)
		}
	}
}
