package models

import (
	"encoding/json"
	"math"
	"time"
)

const (
	epsilon = 0.1
	// JSONTime is the time format encountered from Spark API
	JSONTime = "2006-01-02T15:04:05.000Z"
)

// ReadingStore is an interface for any type that defines methods for storing
// and accessing readings.
type ReadingStore interface {
	StoreReading(reading Reading) error
	GetLatestReadings(userEmail string) ([]Reading, error)
	GetReadings(core string, start, end time.Time) ([]Reading, error)
	DeleteReadings(core string, start, end time.Time) error
}

// Reading represents a distinct reading of environment attributes sent by a
// user's Spark 'Core' or other device at specific point in time.
type Reading struct {
	UserEmail   string    `json:"user"`
	CoreID      string    `json:"coreid"`
	Posted      time.Time `json:"posted"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Moisture    float64   `json:"moisture"`
	Light       float64   `json:"light"`
	Battery     float64   `json:"battery"`
}

// ReadingFromJSON is a convenience method for building a Reading from a
// request sent by a Particle webhook.  Potentially deprecated.
func ReadingFromJSON(userEmail, coreID string, posted time.Time, JSONStr string) (Reading, error) {
	var reading Reading

	if err := json.Unmarshal([]byte(JSONStr), &reading); err != nil {
		return reading, err
	}

	reading.UserEmail = userEmail
	reading.CoreID = coreID
	reading.Posted = posted

	return reading, nil
}

// DataJSON formats just the environmental attributes of the readings into a JSON string.
// This is primarily used to mock the JSON sent in a Particle webhook for testing.
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

// Compare checks if two readings are the same by checking equality of string
// fields and that float values are within acceptible margins of each other
// (because float comparisons tricky).
func (r Reading) Compare(b Reading) bool {
	if r.UserEmail != b.UserEmail {
		return false
	}

	if r.CoreID != b.CoreID {
		return false
	}

	if !r.Posted.Equal(b.Posted) {
		return false
	}

	for _, pair := range []struct {
		a, b float64
	}{
		{r.Temperature, b.Temperature},
		{r.Humidity, b.Humidity},
		{r.Moisture, b.Moisture},
		{r.Light, b.Light},
		{r.Battery, b.Battery},
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

// FilterReadings takes a slice of readings and a function defining a filter,
// and returns a subslice of the Readings containing all the matching items.
func FilterReadings(input []Reading, filter func(Reading) bool) (output []Reading) {
	for _, r := range input {
		if filter(r) {
			output = append(output, r)
		}
	}

	return
}
