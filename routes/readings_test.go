package routes

import (
	"encoding/json"
	"errors"
	//"github.com/serdmanczyk/freyr/middleware"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var randGen = rand.New(rand.NewSource(time.Now().Unix()))

// TODO: move fakeStores and reading comparators into own package
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

type fakeReadingStore struct {
	readings []models.Reading
}

func filterReadings(input []models.Reading, filter func(models.Reading) bool) (output []models.Reading) {
	for _, r := range input {
		if filter(r) {
			output = append(output, r)
		}
	}

	return
}

func (f *fakeReadingStore) StoreReading(reading models.Reading) error {
	f.readings = append(f.readings, reading)
	return nil
}

func (f *fakeReadingStore) GetLatestReadings(userEmail string) (readings []models.Reading, err error) {
	err = errors.New("don't need this")
	return
}

func (f *fakeReadingStore) GetReadings(core string, start, end time.Time) ([]models.Reading, error) {
	filtered := filterReadings(f.readings, func(r models.Reading) bool {
		if r.CoreId == core && r.Posted.After(start) && r.Posted.Before(end) {
			return true
		}

		return false
	})

	return filtered, nil
}

func dataJSON(r models.Reading) string {
	data := map[string]float32{
		"temperature": r.Temperature,
		"humidity":    r.Humidity,
		"moisture":    r.Moisture,
		"light":       r.Light,
		"battery":     r.Battery,
	}

	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func formData(r models.Reading) string {
	v := url.Values{}
	v.Set("event", "post_reading")
	v.Set("data", dataJSON(r))
	v.Set("coreid", r.CoreId)
	v.Set("published_at", r.Posted.Format(models.JsonTime))
	return v.Encode()
}

func TestPostReading(t *testing.T) {
	userEmail := "johndoe@stupidname.com"
	coreid := "78348972452498"

	fS := &fakeReadingStore{}

	postTime := time.Unix(5, 0).In(time.UTC)
	reading := randReading(userEmail, coreid, postTime)

	formEncoded := formData(reading)
	body := strings.NewReader(formEncoded)
	postReadingReq, err := http.NewRequest("POST", "/post_reading", body)
	if err != nil {
		t.Fatal(err)
	}

	postReadingReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// typically handled by middleware
	postReadingReq.ParseForm()
	postReadingResp := httptest.NewRecorder()

	emailCtx := context.WithValue(context.Background(), "email", userEmail)
	handler := PostReading(fS)
	handler.ServeHTTP(emailCtx, postReadingResp, postReadingReq)

	storedReadings, err := fS.GetReadings(coreid, postTime.Add(time.Second*-1), postTime.Add(time.Second*1))
	if err != nil {
		t.Fatal(err)
	}

	if len(storedReadings) != 1 {
		t.Fatalf("Invalid number of readings returned; expected 1, got %d", len(storedReadings))
	}

	if !compare(storedReadings[0], reading) {
		t.Fatalf("Stored reading doesn't match; expected %v, got %v", reading, storedReadings[0])
	}
}

func TestGetReadings(t *testing.T) {
	userEmail := "johndoe@stupidname.com"
	coreid := "78348972452498"

	fS := &fakeReadingStore{}

	start := time.Now().In(time.UTC)
	timeStamp := start
	for i := 0; i < 100; i++ {
		err := fS.StoreReading(randReading(userEmail, coreid, timeStamp))
		if err != nil {
			t.Fatal(err)
		}
		timeStamp = timeStamp.Add(time.Second)
	}

	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", timeStamp.Format(time.RFC3339))
	query.Add("core", coreid)
	reqUrl := "/get_readings?" + query.Encode()

	getReadingsReq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	getReadingsResp := httptest.NewRecorder()

	emailCtx := context.WithValue(context.Background(), "email", userEmail)
	handler := GetReadings(fS)
	handler.ServeHTTP(emailCtx, getReadingsResp, getReadingsReq)

	var retReadings []models.Reading

	if getReadingsResp.Code != http.StatusOK {
		t.Fatalf("Incorrect response code; expected %d, got %d", http.StatusOK, getReadingsResp.Code)
	}

	err = json.NewDecoder(getReadingsResp.Body).Decode(&retReadings)
	if err != nil {
		t.Fatal(err)
	}

	if len(retReadings) != 100 {
		t.Fatal("Unexpected number of readings returned; expected 100 got %d", len(retReadings))
	}
}
