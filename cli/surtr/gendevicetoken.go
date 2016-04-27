package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
	"time"
)

var genDeviceTokenUsage = `Generate a device token, to accompany requests sent by sensor devices.
	email secret coreid [-expiration]`

var genDeviceToken = cli.Command{
	Name:  "gendevicetoken",
	Usage: genDeviceTokenUsage,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "expiration",
			Value: time.Now().Add(time.Hour).Format(time.RFC3339),
			Usage: "Expiration time of token",
		},
	},
	Action: prepareGenDeviceToken,
}

func prepareGenDeviceToken(c *cli.Context) {
	args := c.Args()
	if len(args) != 3 {
		fmt.Println(c.Command.Usage)
		return
	}

	email, secret, coreid := args[0], args[1], args[2]
	expiration := c.String("expiration")

	expiry, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		exit(err)
	}

	fmt.Println(generateDeviceToken(email, secret, coreid, expiry))
}

func generateDeviceToken(email, secret, coreid string, expiry time.Time) string {
	parsedSecret, err := models.SecretFromBase64(secret)
	if err != nil {
		exit(err)
	}

	tkGen := token.JtwTokenGen(parsedSecret)

	token, err := tkGen.GenerateToken(expiry, token.Claims{
		"email": email,
		"core":  coreid,
	})
	if err != nil {
		exit(err)
	}
	return token
}
