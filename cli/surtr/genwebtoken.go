package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
	"time"
)

var webTokenUsage = `Generate a user token, useable as a browser cookie.
	secret email [-expiration]`

var genWebToken = cli.Command{
	Name:  "genusertoken",
	Usage: webTokenUsage,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "expiration",
			Value: time.Now().Add(time.Hour).Format(time.RFC3339),
			Usage: "Expiration time of token",
		},
	},
	Action: prepareGenWebToken,
}

func prepareGenWebToken(c *cli.Context) {
	args := c.Args()
	if len(args) != 2 {
		fmt.Println(c.Command.Usage)
		return
	}

	secret, email := args[0], args[1]
	expiration := c.String("expiration")

	expiry, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		exit(err)
	}

	fmt.Println(generateWebToken(secret, email, expiry))
}

func generateWebToken(secret, email string, expiry time.Time) string {
	parsedSecret, err := models.SecretFromBase64(secret)
	if err != nil {
		exit(err)
	}

	tkGen := token.JtwTokenGen(parsedSecret)
	token, err := token.GenerateWebToken(tkGen, expiry, email)
	if err != nil {
		exit(err)
	}
	return token
}
