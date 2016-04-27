package main

import (
	"github.com/codegangsta/cli"
	"os"
	"fmt"
)

func main() {
	app := cli.NewApp()
	app.Name = "surtr"
	app.Usage = "cli for testing against freyr server"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		genWebToken,
		genDeviceToken,
		PostReading,
	}

	app.Run(os.Args)
}

func exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}
