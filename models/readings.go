package models

import (
	"encoding/json"
	"time"
)

const (
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
	Temperature float32   `json:"temperature"`
	Humidity    float32   `json:"humidity"`
	Moisture    float32   `json:"moisture"`
	Light       float32   `json:"light"`
	Battery     float32   `json:"battery"`
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
