package main

import (
	"flag"
	"github.com/codegangsta/negroni"
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/database"
	"github.com/serdmanczyk/freyr/envflags"
	"github.com/serdmanczyk/freyr/middleware"
	"github.com/serdmanczyk/freyr/oauth"
	"github.com/serdmanczyk/freyr/routes"
	"github.com/serdmanczyk/freyr/token"
	"log"
	"net/http"
	"os"
)

type Config struct {
	OauthClientId     string `flag:"oauthClientId" env:"FREYR_OAUTHID"`
	OauthClientSecret string `flag:"oauthClientSecret" env:"FREYR_OAUTHSECRET"`
	Domain            string `flag:"domain" env:"FREYR_DOMAIN"`
	SecretKey         string `flag:"secretkey" env:"FREYR_SECRET"`
	DBHost            string `flag:"dbhost" env:"FREYR_DBHOST"`
	DBUser            string `flag:"dbuser" env:"FREYR_DBUSER"`
	DBPassword        string `flag:"dbpassw" env:"FREYR_DBPASSW"`
}

func main() {
	var c Config

	envflags.SetFlags(&c)
	flag.Parse()

	if envflags.ConfigEmpty(&c) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	googleOauth := oauth.NewGoogleOauth(c.OauthClientId, c.OauthClientSecret, c.Domain)
	tokenSource := token.JtwTokenGen(c.SecretKey)
	dbConn, err := database.DbConn(c.DBHost, c.DBUser, c.DBPassword)
	if err != nil {
		log.Fatalf("Error initializing database conn: %s", err)
	}

	apiAuth := middleware.NewApiAuthorizer(dbConn)
	apiAuthed := apollo.New(middleware.Authorize(apiAuth))

	webAuth := middleware.NewWebAuthorizer(tokenSource)
	deviceAuth := middleware.NewDeviceAuthorizer(dbConn)

	webAuthed := apollo.New(middleware.Authorize(webAuth, apiAuth))
	webApiAuthed := apollo.New(middleware.Authorize(webAuth, apiAuth))
	apiDeviceAuthed := apollo.New(middleware.Authorize(apiAuth, deviceAuth))

	mux := http.NewServeMux()
	mux.Handle("/", webApiAuthed.Then(routes.Hello()))
	mux.Handle("/secret", webAuthed.Then(routes.GenerateSecret(dbConn)))
	mux.Handle("/latest", webApiAuthed.Then(routes.GetLatestReadings(dbConn)))
	mux.Handle("/readings", webApiAuthed.Then(routes.GetReadings(dbConn)))

	mux.Handle("/reading", apiDeviceAuthed.Then(routes.PostReading(dbConn)))

	mux.Handle("/rotate_secret", apiAuthed.Then(routes.RotateSecret(dbConn)))

	mux.Handle("/authorize", oauth.HandleAuthorize(googleOauth, tokenSource))
	mux.Handle("/oauth2callback", oauth.HandleOAuth2Callback(googleOauth, tokenSource, dbConn))

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.UseHandler(mux)
	n.Run(":8080")
}
