// +build integration

package database

import (
	"github.com/serdmanczyk/freyr/models"
	"math"
	"math/rand"
	"testing"
	"time"
)

var randGen = rand.New(rand.NewSource(time.Now().Unix()))

func floatBetween(a, b float32) float32 {
	if b <= a {
		return 0.0
	}
	return (randGen.Float32() * (b - a)) + a
}

func randReading(userEmail, core string, posted time.Time) models.Reading {
	return models.Reading{
		UserEmail:   userEmail,
		CoreId:      core,
		Posted:      posted,
		Temperature: floatBetween(10.0, 20.0),
		Humidity:    floatBetween(30.0, 60.0),
		Moisture:    floatBetween(20.0, 96.0),
		Light:       floatBetween(0.0, 120.0),
		Battery:     floatBetween(0.0, 100.0),
	}
}

func cmpFloat(a, b float32) bool {
	if math.Abs(float64(a-b)) < 0.1 {
		return true
	}

	return false
}

func compare(a, b models.Reading) bool {
	if a.UserEmail != b.UserEmail {
		return false
	}

	if a.CoreId != b.CoreId {
		return false
	}

	if !cmpFloat(a.Temperature, b.Temperature) {
		return false
	}

	if !cmpFloat(a.Humidity, b.Humidity) {
		return false
	}

	if !cmpFloat(a.Moisture, b.Moisture) {
		return false
	}

	if !cmpFloat(a.Light, b.Light) {
		return false
	}

	if !cmpFloat(a.Battery, b.Battery) {
		return false
	}

	return true
}

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
	reading := randReading(userEmail, coreId, start)

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

	if !compare(reading, retreading) {
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
			err := db.StoreReading(randReading(comb.email, comb.coreid, timeStamp))
			if err != nil {
				t.Fatal(err)
			}
			timeStamp = timeStamp.Add(time.Second)
		}
	}

	maxTimeStamp := timeStamp.Add(time.Second)
	latestInputUserOneCoreOne := randReading(userOneEmail, userOneCoreOne, maxTimeStamp)
	latestInputUserOneCoreTwo := randReading(userOneEmail, userOneCoreTwo, maxTimeStamp)
	latestInputUserTwoCoreOne := randReading(userTwoEmail, userTwoCoreOne, maxTimeStamp)

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

	if !compare(latestInputUserOneCoreOne, latestOutputUserOneCoreOne) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", latestInputUserOneCoreOne, latestOutputUserOneCoreOne)
	}

	latestOutputUserOneCoreTwo, ok := userOneOutMap[userOneCoreTwo]
	if !ok {
		t.Error("Core missing in return from latest")
	}

	if !compare(latestInputUserOneCoreTwo, latestOutputUserOneCoreTwo) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", latestInputUserOneCoreTwo, latestOutputUserOneCoreTwo)
	}

	latestOutputUserTwo, err := db.GetLatestReadings(userTwoEmail)
	if err != nil {
		t.Fatal(err)
	}

	if len(latestOutputUserTwo) != 1 {
		t.Fatalf("Expected one readings returned from latest, got %d", len(latestOutputUserTwo))
	}

	if !compare(latestInputUserTwoCoreOne, latestOutputUserTwo[0]) {
		t.Fatalf("Incorrect reading returned; expected %v, got %v", latestInputUserOneCoreTwo, latestOutputUserOneCoreTwo)
	}
}
