// +build integration

package database

import (
	"github.com/serdmanczyk/freyr/fake"
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

	start := time.Unix(1461300000, 0)
	reading := fake.RandReading(userEmail, coreId, start)

	err = db.StoreReading(reading)
	if err != nil {
		t.Fatal(err)
	}

	readings, err := db.GetReadings(coreId, start.Add(time.Second*-1), start.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	if len(readings) > 1 {
		t.Fatalf("Expected only one reading return, got %d", len(readings))
	}

	retreading := readings[0]

	if !reading.Compare(retreading) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", reading, retreading)
	}
}

func TestDBGetLatest(t *testing.T) {
	userOneEmail := "thor@asgard.unv"
	userOneCoreOne := "123123123142"
	userOneCoreTwo := "456456456456"
	userTwoEmail := "fenrir@hel.unv"
	userTwoCoreOne := "789789789798789"

	for _, userEmail := range []string{userOneEmail, userTwoEmail} {
		err := db.StoreUser(models.User{
			Email: userEmail,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	timeStamp := time.Unix(1461307000, 0)
	for _, comb := range []struct {
		email, coreid string
	}{
		{email: userOneEmail, coreid: userOneCoreOne},
		{email: userOneEmail, coreid: userOneCoreTwo},
		{email: userTwoEmail, coreid: userTwoCoreOne},
	} {
		for i := 0; i < 100; i++ {
			err := db.StoreReading(fake.RandReading(comb.email, comb.coreid, timeStamp))
			if err != nil {
				t.Fatal(err)
			}
			timeStamp = timeStamp.Add(time.Second)
		}
	}

	maxTimeStamp := timeStamp.Add(time.Second)
	latestInputUserOneCoreOne := fake.RandReading(userOneEmail, userOneCoreOne, maxTimeStamp)
	latestInputUserOneCoreTwo := fake.RandReading(userOneEmail, userOneCoreTwo, maxTimeStamp)
	latestInputUserTwoCoreOne := fake.RandReading(userTwoEmail, userTwoCoreOne, maxTimeStamp)

	for _, reading := range []models.Reading{
		latestInputUserOneCoreOne,
		latestInputUserOneCoreTwo,
		latestInputUserTwoCoreOne,
	} {
		err := db.StoreReading(reading)
		if err != nil {
			t.Fatal(err)
		}
	}

	latestOutputUserOne, err := db.GetLatestReadings(userOneEmail)
	if err != nil {
		t.Fatal(err)
	}

	if len(latestOutputUserOne) != 2 {
		t.Fatalf("Expected one readings returned from latest, got %d", len(latestOutputUserOne))
	}

	userOneOutMap := make(map[string]models.Reading)

	for _, reading := range latestOutputUserOne {
		userOneOutMap[reading.CoreId] = reading
	}

	latestOutputUserOneCoreOne, ok := userOneOutMap[userOneCoreOne]
	if !ok {
		t.Error("Core missing in return from latest")
	}

	if !latestInputUserOneCoreOne.Compare(latestOutputUserOneCoreOne) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", latestInputUserOneCoreOne, latestOutputUserOneCoreOne)
	}

	latestOutputUserOneCoreTwo, ok := userOneOutMap[userOneCoreTwo]
	if !ok {
		t.Error("Core missing in return from latest")
	}

	if !latestInputUserOneCoreTwo.Compare(latestOutputUserOneCoreTwo) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", latestInputUserOneCoreTwo, latestOutputUserOneCoreTwo)
	}

	latestOutputUserTwo, err := db.GetLatestReadings(userTwoEmail)
	if err != nil {
		t.Fatal(err)
	}

	if len(latestOutputUserTwo) != 1 {
		t.Fatalf("Expected one readings returned from latest, got %d", len(latestOutputUserTwo))
	}

	if !latestInputUserTwoCoreOne.Compare(latestOutputUserTwo[0]) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", latestInputUserOneCoreTwo, latestOutputUserOneCoreTwo)
	}
}

func TestDeleteReadings(t *testing.T) {
	userEmail := "loki@niflheim.unv"
	core := "6666666666"

	err := db.StoreUser(models.User{
		Email: userEmail,
	})

	start := time.Now()
	end := start.Add(time.Hour * 24 * 7)
	step := time.Second * 60 * 15
	readingGen := fake.ReadingGen(userEmail, core, start, step)

	for now := start; now.Before(end); now = now.Add(step) {
		err := db.StoreReading(readingGen())
		if err != nil {
			t.Fatal(err)
		}
	}

	readings, err := db.GetReadings(core, start, end)
	if err != nil {
		t.Fatal(err)
	}

	if len(readings) == 0 {
		t.Fatal("No readings inserted into database")
	}

	err = db.DeleteReadings(core, start, end)
	if err != nil {
		t.Fatal(err)
	}

	readings, err = db.GetReadings(core, start, end)
	if err != nil {
		t.Fatal(err)
	}

	if len(readings) != 0 {
		t.Fatalf("Readings still remain after delete: %d, %v", len(readings), readings)
	}
}
