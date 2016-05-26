package fake

import (
	"errors"
	"github.com/serdmanczyk/freyr/models"
	"math"
	"math/rand"
	"time"
)

var randGen *rand.Rand

func init() {
	randGen = rand.New(rand.NewSource(time.Now().Unix()))
}

// FloatBetween returns a random float value between a and b
func FloatBetween(a, b float64) float64 {
	if b <= a {
		return 0.0
	}
	return (randGen.Float64() * (b - a)) + a
}

// RandReading returns a reading with specified userEmail, core, and posted time
// and other reading values set to random values within a reasonable range.
func RandReading(userEmail, core string, posted time.Time) models.Reading {
	return models.Reading{
		UserEmail:   userEmail,
		CoreID:      core,
		Posted:      posted,
		Temperature: FloatBetween(10.0, 20.0),
		Humidity:    FloatBetween(30.0, 60.0),
		Moisture:    FloatBetween(20.0, 96.0),
		Light:       FloatBetween(0.0, 120.0),
		Battery:     FloatBetween(0.0, 100.0),
	}
}

// ReadingStore is a fake implementation of the models.ReadingStore interface
// via an in memory slice.  It is used for unit tests of libraries that accept
// a models.ReadingStore interface.
type ReadingStore struct {
	readings []models.Reading
}

// StoreReading appends the reading to its slice of readings
func (f *ReadingStore) StoreReading(reading models.Reading) error {
	f.readings = append(f.readings, reading)
	return nil
}

// GetLatestReadings is a placeholder to fulfill the models.ReadingStore interface
func (f *ReadingStore) GetLatestReadings(userEmail string) (readings []models.Reading, err error) {
	err = errors.New("don't need this yet")
	return
}

// DeleteReadings is a placeholder to fulfill the models.ReadingStore interface
func (f *ReadingStore) DeleteReadings(core string, start, end time.Time) error {
	return errors.New("don't need this yet")
}

// GetReadings returns readings in its slice of readings that lie between
// the specified start and end time.
func (f *ReadingStore) GetReadings(core string, start, end time.Time) ([]models.Reading, error) {
	filtered := models.FilterReadings(f.readings, func(r models.Reading) bool {
		if r.CoreID == core && r.Posted.After(start) && r.Posted.Before(end) {
			return true
		}

		return false
	})

	return filtered, nil
}

// Translator returns a function that translates input floats
// from linear domain a-b to domain c-d.
func Translator(a, b, c, d float64) func(float64) float64 {
	factor := (d - c) / (b - a)

	return func(x float64) float64 {
		return (x-a)*factor + c
	}
}

// FourierSineGen returns a function that generates sequential values in a
// fourier sine series given the input parameters.
func FourierSineGen(current, period, step, min, max float64, params ...float64) func() float64 {
	var delta float64
	for _, a := range params {
		delta += a
	}

	trans := Translator(-delta, delta, min, max)

	return func() float64 {
		var y float64
		for d, A := range params {
			f := ((1 / period) * float64(d))
			y += A * math.Sin(current*2*math.Pi*f)
		}

		current += step
		return trans(y)
	}
}

// ReadingGen returns a function; the returned function generates readings
// with the input userEmail and coreid that start at the input current time
// and iterate in values of step.  The values for temperature, humidity, etc.
// follow a fourier sine series that stays within reasonable values.  The main
// point of this is to generate test values that will yield a visuably reasonable
// graph.
func ReadingGen(userEmail, coreID string, current time.Time, step time.Duration) func() models.Reading {
	cfloat := float64(current.Unix())
	dayF := float64(time.Hour * 24)
	stepF := float64(step)

	tempGen := FourierSineGen(cfloat, dayF, stepF, 10.0, 20.0, 1, 0.75, 0.75)
	humGen := FourierSineGen(cfloat, dayF, stepF, 30.0, 60.0, 0.35, 0.75, 0.25)
	moistGen := FourierSineGen(cfloat, dayF, stepF, 20.0, 96.0, 0.5, 0.25, 0.45)
	lightGen := FourierSineGen(cfloat, dayF, stepF, 0.0, 120.0, 0.25, 0.1, 0.2)
	battGen := FourierSineGen(cfloat, dayF, stepF, 10.0, 90.0, 1, 0.35, 0.5)

	return func() models.Reading {
		reading := models.Reading{
			UserEmail:   userEmail,
			CoreID:      coreID,
			Posted:      current,
			Temperature: tempGen(),
			Humidity:    humGen(),
			Moisture:    moistGen(),
			Light:       lightGen(),
			Battery:     battGen(),
		}

		current = current.Add(step)
		return reading
	}
}
