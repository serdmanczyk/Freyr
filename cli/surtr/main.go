package main

import (
	"github.com/codegangsta/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "surtr"
	app.Usage = "cli for testing against freyr server"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		genUserToken,
		genDeviceToken,
	}

	app.Run(os.Args)
}
