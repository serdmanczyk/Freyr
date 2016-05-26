package routes

import (
	"encoding/json"
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func formData(r models.Reading) string {
	v := url.Values{}
	v.Set("event", "post_reading")
	v.Set("data", r.DataJSON())
	v.Set("coreid", r.CoreID)
	v.Set("published_at", r.Posted.Format(models.JSONTime))

	return v.Encode()
}

func TestPostReading(t *testing.T) {
	userEmail := "johndoe@stupidname.com"
	coreid := "78348972452498"

	fS := &fake.ReadingStore{}

	postTime := time.Unix(5, 0).In(time.UTC)
	reading := fake.RandReading(userEmail, coreid, postTime)

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

	if postReadingResp.Code != http.StatusCreated {
		t.Fatalf("Incorrect response code; expected %d, got %d", http.StatusCreated, postReadingResp.Code)
	}

	storedReadings, err := fS.GetReadings(coreid, postTime.Add(time.Second*-1), postTime.Add(time.Second*1))
	if err != nil {
		t.Fatal(err)
	}

	if len(storedReadings) != 1 {
		t.Fatalf("Invalid number of readings returned; expected 1, got %d", len(storedReadings))
	}

	if !storedReadings[0].Compare(reading) {
		t.Fatalf("Stored reading doesn't match; expected %v, got %v", reading, storedReadings[0])
	}
}

func TestGetReadings(t *testing.T) {
	userEmail := "johndoe@stupidname.com"
	coreid := "78348972452498"

	fS := &fake.ReadingStore{}

	start := time.Now().In(time.UTC)
	timeStamp := start
	for i := 0; i < 100; i++ {
		err := fS.StoreReading(fake.RandReading(userEmail, coreid, timeStamp))
		if err != nil {
			t.Fatal(err)
		}
		timeStamp = timeStamp.Add(time.Second)
	}

	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", timeStamp.Format(time.RFC3339))
	query.Add("core", coreid)
	reqURL := "/get_readings?" + query.Encode()

	getReadingsReq, err := http.NewRequest("GET", reqURL, nil)
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
		t.Fatalf("Unexpected number of readings returned; expected 100 got %d", len(retReadings))
	}
}
