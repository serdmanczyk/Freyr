package routes

import (
	"encoding/json"
	"errors"
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"net/http"
	"time"
	"io"
)

var (
	NoReading = errors.New("reading not present in request")
)

func StringsEmpty(strs ...string) bool {
	for _, s := range strs {
		if s == "" {
			return true
		}
	}
	return false
}

func loadReading(ctx context.Context, r *http.Request) (models.Reading, error) {
	email := getEmail(ctx)
	coreid := r.FormValue("coreid")
	published := r.FormValue("published_at")
	dataStr := r.FormValue("data")

	if StringsEmpty(email, coreid, published, dataStr) {
		return models.Reading{}, NoReading
	}

	readingData := make(map[string]float32)
	err := json.Unmarshal([]byte(dataStr), &readingData)
	if err != nil {
		return models.Reading{}, err
	}

	posted, err := time.Parse(models.JsonTime, published)
	if err != nil {
		return models.Reading{}, err
	}

	reading := models.Reading{
		UserEmail:   email,
		CoreId:      coreid,
		Posted:      posted,
		Temperature: readingData["temperature"],
		Humidity:    readingData["humidity"],
		Moisture:    readingData["moisture"],
		Light:       readingData["light"],
		Battery:     readingData["battery"],
	}
	return reading, nil
}

func PostReading(s models.ReadingStore) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		reading, err := loadReading(ctx, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.StoreReading(reading); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	})
}

func GetReadings(s models.ReadingStore) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		start, end, core, err := getReadingsParams(ctx, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		readings, err := s.GetReadings(core, start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err == nil && len(readings) == 0 {
			http.Error(w, "[]", http.StatusNotFound)
			return
		}

		bytes, err := json.Marshal(readings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(bytes))
	})
}

func getReadingsParams(ctx context.Context, r *http.Request) (start, end time.Time, core string, err error) {
	startDate := r.FormValue("start")
	if startDate == "" {
		err = errors.New("start date missing from query")
		return
	}

	start, err = time.Parse(time.RFC3339, startDate)
	if err != nil {
		return
	}

	endDate := r.FormValue("end")
	if endDate == "" {
		err = errors.New("end date missing from query")
		return
	}

	end, err = time.Parse(time.RFC3339, endDate)
	if err != nil {
		return
	}

	core = r.FormValue("core")
	if core == "" {
		err = errors.New("core id missing from query")
		return
	}

	return
}
