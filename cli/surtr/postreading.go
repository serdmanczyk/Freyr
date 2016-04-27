package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/serdmanczyk/freyr/cli/client"
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"time"
)

var PostReadingUsage = `Post a new reading to the server
	domain email secret coreid posted [-m -temperature, -humidity, -moisture, -light, -battery]

	If present, '-n N' will indicate N number of readings will be posted.  Readings will start from
	'posted' and increment by 1 second each.  If -n is specified, other flags are ignored and
	readings will be randomized`

var PostReading = cli.Command{
	Name:  "postreading",
	Usage: PostReadingUsage,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "n",
			Value: 1,
			Usage: "number of readings to post",
		},
		cli.Float64Flag{
			Name:  "temperature",
			Value: float64(fake.FloatBetween(10.0, 30.0)),
			Usage: "Temperature",
		},
		cli.Float64Flag{
			Name:  "humidity",
			Value: float64(fake.FloatBetween(30.0, 90.0)),
			Usage: "Humidity",
		},
		cli.Float64Flag{
			Name:  "moisture",
			Value: float64(fake.FloatBetween(0.0, 90.0)),
			Usage: "Moisture",
		},
		cli.Float64Flag{
			Name:  "light",
			Value: float64(fake.FloatBetween(0.0, 120.0)),
			Usage: "Light",
		},
		cli.Float64Flag{
			Name:  "battery",
			Value: float64(fake.FloatBetween(0.0, 100.0)),
			Usage: "Battery",
		},
	},
	Action: postReading,
}

func postReading(c *cli.Context) {
	args := c.Args()
	if len(args) != 5 {
		fmt.Println(c.Command.Usage)
		return
	}

	domain, email, secret, coreid, posted := args[0], args[1], args[2], args[3], args[4]

	postedTime, err := time.Parse(time.RFC3339, posted)
	if err != nil {
		exit(err)
	}

	frey, err := client.New(domain, email, secret)
	if err != nil {
		exit(err)
	}

	number := c.Int("n")
	if number > 0 {
		for i := 0; i < number; i++ {
			reading := fake.RandReading(email, coreid, postedTime)

			err = frey.PostReading(reading)
			if err != nil {
				exit(err)
			}
			// TODO: log here
			postedTime = postedTime.Add(time.Second)
		}
		return
	}

	temperature := float32(c.Float64("temperature"))
	humidity := float32(c.Float64("humidity"))
	moisture := float32(c.Float64("moisture"))
	light := float32(c.Float64("light"))
	battery := float32(c.Float64("battery"))

	reading := models.Reading{
		UserEmail:   email,
		CoreId:      coreid,
		Posted:      postedTime,
		Temperature: temperature,
		Humidity:    humidity,
		Moisture:    moisture,
		Light:       light,
		Battery:     battery,
	}

	err = frey.PostReading(reading)
	if err != nil {
		exit(err)
	}
}
