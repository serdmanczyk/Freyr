package models

import (
	"encoding/json"
	"math"
	"time"
)

const (
	epsilon  = 0.1
	JsonTime = "2006-01-02T15:04:05.000Z"
)

type ReadingStore interface {
	StoreReading(reading Reading) error
	GetLatestReadings(userEmail string) ([]Reading, error)
	GetReadings(core string, start, end time.Time) ([]Reading, error)
}

type Reading struct {
	UserEmail   string    `json:"user"`
	CoreId      string    `json:"coreid"`
	Posted      time.Time `json:"posted"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Moisture    float64   `json:"moisture"`
	Light       float64   `json:"light"`
	Battery     float64   `json:"battery"`
}

func ReadingFromJSON(userEmail, coreId string, posted time.Time, jsonStr string) (Reading, error) {
	var reading Reading

	if err := json.Unmarshal([]byte(jsonStr), &reading); err != nil {
		return reading, err
	}

	reading.UserEmail = userEmail
	reading.CoreId = coreId
	reading.Posted = posted

	return reading, nil
}

func (r Reading) DataJSON() string {
	data := map[string]float64{
		"temperature": r.Temperature,
		"humidity":    r.Humidity,
		"moisture":    r.Moisture,
		"light":       r.Light,
		"battery":     r.Battery,
	}

	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func (a Reading) Compare(b Reading) bool {
	if a.UserEmail != b.UserEmail {
		return false
	}

	if a.CoreId != b.CoreId {
		return false
	}

	if !a.Posted.Equal(b.Posted) {
		return false
	}

	for _, pair := range []struct {
		a, b float64
	}{
		{a.Temperature, b.Temperature},
		{a.Humidity, b.Humidity},
		{a.Moisture, b.Moisture},
		{a.Light, b.Light},
		{a.Battery, b.Battery},
	} {
		if !floatCompare(pair.a, pair.b) {
			return false
		}
	}

	return true
}

func floatCompare(a, b float64) bool {
	if math.Abs(float64(a-b)) < epsilon {
		return true
	}

	return false
}

func FilterReadings(input []Reading, filter func(Reading) bool) (output []Reading) {
	for _, r := range input {
		if filter(r) {
			output = append(output, r)
		}
	}

	return
}
