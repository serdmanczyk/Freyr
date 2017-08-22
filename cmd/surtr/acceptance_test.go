// +build acceptance

package main

import (
	"bytes"
	"flag"
	"github.com/serdmanczyk/freyr/client"
	"github.com/serdmanczyk/freyr/envflags"
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
	"log"
	"os"
	"testing"
	"time"
)

type Config struct {
	Domain    string `flag:"domain" env:"SURTR_DOMAIN"`
	TestUser  string `flag:"user" env:"SURTR_USER"`
	SecretKey string `flag:"secretkey" env:"SURTR_SECRET"`
}

func TestAcceptance(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		t.Log(buf.String())
		log.SetOutput(os.Stderr)
	}()

	var c Config

	envflags.SetFlags(&c)
	flag.Parse()

	if envflags.ConfigEmpty(&c) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	time.Sleep(time.Second)

	// First test, test secret generation facilities
	// generate a web token with the system secret just
	// to get a secret for the user.
	token, err := token.GenerateWebToken(token.JWTTokenGen(c.SecretKey), time.Now().Add(time.Hour), c.TestUser)
	if err != nil {
		t.Fatal(err)
	}

	webSignator := client.WebSignator{Token: token}

	userSecret, err := client.GetSecret(webSignator, c.Domain)
	if err != nil {
		t.Fatal(err)
	}

	apiSignator := client.APISignator{
		UserEmail: c.TestUser,
		Secret:    userSecret,
	}

	newUserSecret, err := client.RotateSecret(apiSignator, c.Domain)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.RotateSecret(apiSignator, c.Domain)
	if err == nil {
		t.Fatal("Secret should have rotated, should have gotten unauthorized response")
	}

	apiSignator = client.APISignator{
		UserEmail: c.TestUser,
		Secret:    newUserSecret,
	}

	// Test Two: send a slew of readings and test that the
	// get latest call returns the correct latest readings
	sentLatest := make(map[string]models.Reading)
	startTime := time.Now().In(time.UTC).Round(time.Second)
	for _, coreId := range []string{"123123123123", "456456456456"} {
		postTime := startTime
		var reading models.Reading
		for i := 0; i < 100; i++ {
			reading = fake.RandReading(c.TestUser, coreId, postTime)
			err = client.PostReading(apiSignator, c.Domain, reading)
			if err != nil {
				t.Fatalf("Error posting reading: %s", err.Error())
			}
			postTime = postTime.Add(time.Second)
		}
		sentLatest[coreId] = reading
	}

	returnedLatest, err := client.GetLatest(apiSignator, c.Domain)
	if err != nil {
		t.Fatalf("Error retreiving latest readings: %s", err.Error())
	}

	returned := make(map[string]models.Reading)
	for _, reading := range returnedLatest {
		returned[reading.CoreID] = reading
	}

	for _, sentReading := range sentLatest {
		returnedReading, ok := returned[sentReading.CoreID]
		if !ok {
			t.Fatal("Core reading missing from results in latest call")
		}
		if !sentReading.Compare(returnedReading) {
			t.Fatal("Latest reading returned for core does not match sent reading %v %v", returnedReading, sentReading)
		}
	}

	for _, returnedReading := range returned {
		_, ok := sentLatest[returnedReading.CoreID]
		if !ok {
			t.Fatal("Core returned from call to latest should not be present")
		}
	}

	// Test Three: send a deterministic list of readings and test
	// that a call to get readings between those dates returns the
	// correct readings.
	startTime = time.Now().In(time.UTC).Round(time.Second)
	postTime := startTime
	const numReadings = 100

	sentReadings := make(map[time.Time]models.Reading, numReadings)
	coreId := "890890890890"
	for i := 0; i < numReadings; i++ {
		reading := fake.RandReading(c.TestUser, coreId, postTime)
		err = client.PostReading(apiSignator, c.Domain, reading)
		if err != nil {
			t.Fatalf("Error posting reading: ", err.Error())
		}
		sentReadings[postTime] = reading
		postTime = postTime.Add(time.Second)
	}

	readings, err := client.GetReadings(apiSignator, c.Domain, coreId, startTime, postTime)
	if err != nil {
		t.Fatalf("Error calling get on readings by date span: %s", err.Error())
	}

	returnedReadings := make(map[time.Time]models.Reading, numReadings)
	for _, reading := range readings {
		if _, ok := returnedReadings[reading.Posted]; ok {
			t.Fatal("Multiple readings for same date returned; readings should be unique per core per date")
		}
		returnedReadings[reading.Posted] = reading
	}

	for posted, sentReading := range sentReadings {
		returnedReading, ok := returnedReadings[posted]
		if !ok {
			t.Fatal("Missing reading in API response for date: " + posted.Format(time.RFC3339))
		}
		if !sentReading.Compare(returnedReading) {
			t.Fatalf("Returned reading in API response doesn't match sent for date: %s", posted.Format(time.RFC3339))
		}
	}

	for posted, returnedReading := range returnedReadings {
		_, ok := sentReadings[posted]
		if !ok {
			t.Fatalf("Returned reading in API response shouldn't exist %v", returnedReading)
		}
	}

	err = client.DeleteReadings(apiSignator, c.Domain, coreId, startTime, postTime)
	if err != nil {
		t.Fatalf("Error deleting readings: %s", err.Error())
	}

	// Re-use readings we sent earlier for PostReadings call
	var readingsList []models.Reading
	for _, reading := range sentReadings {
		readingsList = append(readingsList, reading)
	}

	jobID, err := client.PostReadings(apiSignator, c.Domain, readingsList)
	if err != nil {
		t.Fatalf("Error posting multiple readings: %s", err.Error())
	}

	err = client.WaitForJob(apiSignator, c.Domain, jobID, time.Minute)
	if err != nil {
		t.Fatalf("Failed waiting for job %d: %s", jobID, err.Error())
	}

	readings, err = client.GetReadings(apiSignator, c.Domain, coreId, startTime, postTime)
	if err != nil {
		t.Fatalf("Error calling get on readings by date span: %s", err.Error())
	}

	for _, reading := range readings {
		if _, ok := sentReadings[reading.Posted]; !ok {
			t.Fatal("Reading was not added in post multiple call: %v", reading)
		}
	}
}
