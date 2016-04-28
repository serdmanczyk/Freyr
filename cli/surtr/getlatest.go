package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/serdmanczyk/freyr/cli/client"
	"os"
)

var GetLatestUsage = `filll in here`

var GetLatest = cli.Command{
	Name:   "getlatest",
	Usage:  GetLatestUsage,
	Action: getLatest,
}

func getLatest(c *cli.Context) {
	args := c.Args()
	if len(args) != 3 {
		fmt.Println(c.Command.Usage)
		return
	}

	domain, email, secret := args[0], args[1], args[2]

	signator, err := client.NewApiSignator(email, secret)
	if err != nil {
		exit(err)
	}

	readings, err := client.GetLatest(signator, domain)
	if err != nil {
		exit(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(readings)
	if err != nil {
		exit(err)
	}
}
