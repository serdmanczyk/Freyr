package database

import (
	"github.com/serdmanczyk/freyr/models"
	"time"
)

func (db DB) GetLatestReadings(userEmail string) ([]models.Reading, error) {
	var readings []models.Reading

	rows, err := db.Query(`
	select readings.useremail, readings.posted, readings.coreid, readings.posted, readings.temperature,
		readings.humidity, readings.moisture, readings.light, readings.battery
	from readings inner join 
	    (select coreid, max(posted) from
	        (select * from readings where useremail = $1) as userreadings
	    group by coreid) as maxposted
	    on readings.coreid = maxposted.coreid and readings.posted = maxposted.max`, userEmail)
	if err != nil {
		return readings, err
	}
	defer rows.Close()

	for rows.Next() {
		reading := models.Reading{}

		err := rows.Scan(&reading.UserEmail, &reading.Posted, &reading.CoreId, &reading.Posted,
			&reading.Temperature, &reading.Humidity, &reading.Moisture, &reading.Light, &reading.Battery)
		if err != nil {
			return readings, err
		}

		readings = append(readings, reading)
	}

	if err := rows.Err(); err != nil {
		return readings, err
	}

	return readings, err
}

func (db DB) StoreReading(reading models.Reading) error {
	_, err := db.Exec(`insert into readings
		(useremail, posted, coreid, temperature, humidity, moisture, light, battery)
		values ($1, $2, $3, $4, $5, $6, $7, $8);`,
		reading.UserEmail, reading.Posted, reading.CoreId,
		reading.Temperature, reading.Humidity, reading.Moisture, reading.Light, reading.Battery)
	if err != nil {
		return err
	}

	return nil
}

func (db DB) GetReadings(core string, start, end time.Time) ([]models.Reading, error) {
	var readings []models.Reading

	rows, err := db.Query(`select
		useremail, posted, coreid, temperature, humidity, moisture, light, battery
		from readings where coreid = $1 and posted between $2 and $3`, core, start, end)
	if err != nil {
		return readings, err
	}
	defer rows.Close()

	for rows.Next() {
		reading := models.Reading{}

		err := rows.Scan(&reading.UserEmail, &reading.Posted, &reading.CoreId, &reading.Temperature,
			&reading.Humidity, &reading.Moisture, &reading.Light, &reading.Battery)
		if err != nil {
			return readings, err
		}

		readings = append(readings, reading)
	}

	if err := rows.Err(); err != nil {
		return readings, err
	}

	return readings, err
}
