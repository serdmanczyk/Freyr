package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
	"time"
)

var userUsage = `Generate a user token, useable as a browser cookie.
	secret email [-expiration]`

var genUserToken = cli.Command{
	Name:  "genusertoken",
	Usage: userUsage,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "expiration",
			Value: time.Now().Add(time.Hour).Format(time.RFC3339),
			Usage: "Expiration time of token",
		},
	},
	Action: prepareGenUserToken,
}

func prepareGenUserToken(c *cli.Context) {
	args := c.Args()
	if len(args) != 2 {
		fmt.Println(c.Command.Usage)
		return
	}

	secret, email := args[0], args[1]
	expiration := c.String("expiration")

	expiry, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		panic(err)
	}

	fmt.Println(generateUserToken(secret, email, expiry))
}

func generateUserToken(secret, email string, expiry time.Time) string {
	parsedSecret, err := models.SecretFromBase64(secret)
	if err != nil {
		panic(err)
	}

	tkGen := token.JtwTokenGen(parsedSecret)

	token, err := tkGen.GenerateToken(expiry, token.Claims{
		"email": email,
	})
	if err != nil {
		panic(err)
	}
	return token
}
