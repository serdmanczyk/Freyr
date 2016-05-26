// Package surtr defines the cli interface for making calls against Freyr
package main

import (
	"encoding/json"
	"github.com/jwaldrip/odin/cli"
	"github.com/serdmanczyk/freyr/client"
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
	"os"
	"time"
)

var (
	defaultExp = time.Now().Add(time.Hour).Format(time.RFC3339)
	startTime  = time.Now().In(time.UTC).Format(time.RFC3339)
)

func main() {
	surtr := cli.New("0.1.0", "Freyr client and testing tool", func(c cli.Command) {})

	token := surtr.DefineSubCommand("gentoken", "Generate a token", func(c cli.Command) {
		c.ErrPrintln("Please define type of token to generate, [web, core]")
	})

	get := surtr.DefineSubCommand("get", "get commands", func(c cli.Command) {
		c.ErrPrintln("Define what you want to get [latest, between]")
	})

	post := surtr.DefineSubCommand("post", "post commands", func(c cli.Command) {
		c.ErrPrintln("Define what you want to post [reading]")
	})

	delete := surtr.DefineSubCommand("delete", "deletecommands", func(c cli.Command) {
		c.ErrPrintln("Define what you want to delete [readings]")
	})

	surtr.DefineSubCommand("rotatesecret", "rotate user secret", rotateSecret, "domain", "secret", "email")

	delete.DefineSubCommand("readings", "delete readings", deleteBetween, "domain", "secret", "email", "coreid", "start", "end")

	webToken := token.DefineSubCommand("web", "generate web token", genWebToken, "email", "secret")
	webToken.DefineStringFlag("exp", defaultExp, "expiration date of token")

	coreToken := token.DefineSubCommand("core", "generate core token", genCoreToken, "email", "secret", "coreid")
	coreToken.DefineStringFlag("exp", defaultExp, "expiration date of token")

	get.DefineSubCommand("latest", "get latest readings for cores", getLatest, "domain", "secret", "email")
	get.DefineSubCommand("between", "get latest readings for cores", getBetween, "domain", "secret", "email", "coreid", "start", "end")

	pr := post.DefineSubCommand("reading", "post a reading", postReading, "domain", "secret", "email", "coreid")
	pr.DefineStringFlag("posted", startTime, "Time of reading's posting")
	pr.DefineFloat64Flag("temperature", float64(fake.FloatBetween(10.0, 30.0)), "temperature to post")
	pr.DefineFloat64Flag("humidity", float64(fake.FloatBetween(30.0, 90.0)), "humidity percentage to post")
	pr.DefineFloat64Flag("moisture", float64(fake.FloatBetween(0.0, 90.0)), "moisture level to post")
	pr.DefineFloat64Flag("light", float64(fake.FloatBetween(0.0, 120.0)), "light level to post")
	pr.DefineFloat64Flag("battery", float64(fake.FloatBetween(0.0, 100.0)), "battery levl to post")
	pr.DefineInt64Flag("number", 1, "Number of readings to post")
	pr.DefineInt64Flag("step", int64(time.Minute*15/time.Second), "Step, in time, between readings")
	pr.AliasFlag('p', "posted")
	pr.AliasFlag('t', "temperature")
	pr.AliasFlag('h', "humidity")
	pr.AliasFlag('m', "moisture")
	pr.AliasFlag('l', "light")
	pr.AliasFlag('b', "battery")
	pr.AliasFlag('n', "number")
	pr.AliasFlag('s', "step")

	surtr.Start()
}

func genWebToken(c cli.Command) {
	email := c.Param("email").String()
	base64secret := c.Param("secret").String()
	expString := c.Flag("exp").String()

	parsedSecret, err := models.SecretFromBase64(base64secret)
	if err != nil {
		panic(err)
	}

	exp, err := time.Parse(time.RFC3339, expString)
	if err != nil {
		panic(err)
	}

	tG := token.JWTTokenGen(parsedSecret)

	webToken, err := token.GenerateWebToken(tG, exp, email)
	if err != nil {
		panic(err)
	}
	c.Println(webToken)
}

func genCoreToken(c cli.Command) {
	email := c.Param("email").String()
	base64secret := c.Param("secret").String()
	coreid := c.Param("coreid").String()
	expString := c.Flag("exp").String()

	parsedSecret, err := models.SecretFromBase64(base64secret)
	if err != nil {
		panic(err)
	}

	exp, err := time.Parse(time.RFC3339, expString)
	if err != nil {
		panic(err)
	}

	tG := token.JWTTokenGen(parsedSecret)

	deviceToken, err := token.GenerateDeviceToken(tG, exp, coreid, email)
	if err != nil {
		panic(err)
	}
	c.Println(deviceToken)
}

func getBetween(c cli.Command) {
	domain := c.Param("domain").String()
	secret := c.Param("secret").String()
	email := c.Param("email").String()
	coreid := c.Param("coreid").String()
	start := c.Param("start").String()
	end := c.Param("end").String()

	signator, err := client.NewAPISignator(email, secret)
	if err != nil {
		panic(err)
	}

	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		panic(err)
	}

	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		panic(err)
	}

	readings, err := client.GetReadings(signator, domain, coreid, startTime, endTime)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(readings)
	if err != nil {
		panic(err)
	}
}

func deleteBetween(c cli.Command) {
	domain := c.Param("domain").String()
	secret := c.Param("secret").String()
	email := c.Param("email").String()
	coreid := c.Param("coreid").String()
	start := c.Param("start").String()
	end := c.Param("end").String()

	signator, err := client.NewAPISignator(email, secret)
	if err != nil {
		panic(err)
	}

	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		panic(err)
	}

	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		panic(err)
	}

	err = client.DeleteReadings(signator, domain, coreid, startTime, endTime)
	if err != nil {
		panic(err)
	}
}

func getLatest(c cli.Command) {
	domain := c.Param("domain").String()
	email := c.Param("email").String()
	secret := c.Param("secret").String()

	signator, err := client.NewAPISignator(email, secret)
	if err != nil {
		panic(err)
	}

	readings, err := client.GetLatest(signator, domain)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(readings)
	if err != nil {
		panic(err)
	}
}

func postReading(c cli.Command) {
	domain := c.Param("domain").String()
	email := c.Param("email").String()
	secret := c.Param("secret").String()
	coreid := c.Param("coreid").String()
	posted := c.Flag("posted").String()

	postedTime, err := time.Parse(time.RFC3339, posted)
	if err != nil {
		panic(err)
	}

	signator, err := client.NewAPISignator(email, secret)
	if err != nil {
		panic(err)
	}

	temperature := c.Flag("temperature").Get().(float64)
	humidity := c.Flag("humidity").Get().(float64)
	moisture := c.Flag("moisture").Get().(float64)
	light := c.Flag("light").Get().(float64)
	battery := c.Flag("battery").Get().(float64)
	number := c.Flag("number").Get().(int64)
	step := c.Flag("step").Get().(int64)

	readingGen := fake.ReadingGen(email, coreid, postedTime, time.Duration(step)*time.Second)
	var reading models.Reading

	if number == 1 {
		reading = models.Reading{
			UserEmail:   email,
			CoreID:      coreid,
			Posted:      postedTime,
			Temperature: temperature,
			Humidity:    humidity,
			Moisture:    moisture,
			Light:       light,
			Battery:     battery,
		}

	} else {
		reading = readingGen()
	}

	for i := 0; i < int(number); i++ {
		err = client.PostReading(signator, domain, reading)
		if err != nil {
			panic(err)
		}

		//postedTime = postedTime.Add(time.Second)
		//reading = fake.RandReading(email, coreid, postedTime)
		reading = readingGen()
	}
}

func rotateSecret(c cli.Command) {
	domain := c.Param("domain").String()
	email := c.Param("email").String()
	secret := c.Param("secret").String()

	signator, err := client.NewAPISignator(email, secret)
	if err != nil {
		panic(err)
	}

	newsecret, err := client.RotateSecret(signator, domain)
	if err != nil {
		panic(err)
	}

	c.Println(newsecret.Encode())
}
