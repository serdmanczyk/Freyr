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

func FloatBetween(a, b float64) float64 {
	if b <= a {
		return 0.0
	}
	return (randGen.Float64() * (b - a)) + a
}

func RandReading(userEmail, core string, posted time.Time) models.Reading {
	return models.Reading{
		UserEmail:   userEmail,
		CoreId:      core,
		Posted:      posted,
		Temperature: FloatBetween(10.0, 20.0),
		Humidity:    FloatBetween(30.0, 60.0),
		Moisture:    FloatBetween(20.0, 96.0),
		Light:       FloatBetween(0.0, 120.0),
		Battery:     FloatBetween(0.0, 100.0),
	}
}

type ReadingStore struct {
	readings []models.Reading
}

func (f *ReadingStore) StoreReading(reading models.Reading) error {
	f.readings = append(f.readings, reading)
	return nil
}

func (f *ReadingStore) GetLatestReadings(userEmail string) (readings []models.Reading, err error) {
	err = errors.New("don't need this yet")
	return
}

func (f *ReadingStore) GetReadings(core string, start, end time.Time) ([]models.Reading, error) {
	filtered := models.FilterReadings(f.readings, func(r models.Reading) bool {
		if r.CoreId == core && r.Posted.After(start) && r.Posted.Before(end) {
			return true
		}

		return false
	})

	return filtered, nil
}

// Translate from linear domain a-b to domain a-c
func Translator(a, b, c, d float64) func(float64) float64 {
	factor := (d - c) / (b - a)

	return func(x float64) float64 {
		return (x-a)*factor + c
	}
}

// Generate a Sine wave via Fourier Sine series
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

func ReadingGen(userEmail, coreId string, current time.Time, step time.Duration) func() models.Reading {
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
			CoreId:      coreId,
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
