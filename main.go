// Freyr is a golang web server used to store and retrieve plant environmental
// readings from a user's sensors.  It handles user log-in with oauth and authorizes
// using HMAC signed JWTs.
//
// More at: http://serdmanczyk.github.io.
//
// Github: http://github.com/serdmanczyk/freyr/
package main

import (
	"flag"
	"github.com/codegangsta/negroni"
	"github.com/cyclopsci/apollo"
	_ "github.com/lib/pq"
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

// Config represent the basic configuration needed by Freyr to operate.
type Config struct {
	OauthClientID     string `flag:"oauthClientId" env:"FREYR_OAUTHID"`
	OauthClientSecret string `flag:"oauthClientSecret" env:"FREYR_OAUTHSECRET"`
	Domain            string `flag:"domain" env:"FREYR_DOMAIN"`
	SecretKey         string `flag:"secretkey" env:"FREYR_SECRET"`
	DBHost            string `flag:"dbhost" env:"FREYR_DBHOST"`
	DBUser            string `flag:"dbuser" env:"FREYR_DBUSER"`
	DBPassword        string `flag:"dbpassw" env:"FREYR_DBPASSW"`
	DemoUser          string `flag:"demouser" env:"FREYR_DEMOUSER"`
}

func main() {
	var c Config

	envflags.SetFlags(&c)
	flag.Parse()

	if envflags.ConfigEmpty(&c) {
		flag.PrintDefaults()
		os.Exit(1)
	}

	googleOauth := oauth.NewGoogleOauth(c.OauthClientID, c.OauthClientSecret, c.Domain)
	tokenSource := token.JWTTokenGen(c.SecretKey)
	dbConn, err := database.DBConn("postgres", c.DBHost, c.DBUser, c.DBPassword)
	if err != nil {
		log.Fatalf("Error initializing database conn: %s", err)
	}

	webAuth := middleware.NewWebAuthorizer(tokenSource)
	apiAuth := middleware.NewAPIAuthorizer(dbConn)
	deviceAuth := middleware.NewDeviceAuthorizer(dbConn)

	apiAuthed := apollo.New(middleware.Authorize(apiAuth))
	webAuthed := apollo.New(middleware.Authorize(webAuth))
	webAPIAuthed := apollo.New(middleware.Authorize(webAuth, apiAuth))
	apiDeviceAuthed := apollo.New(middleware.Authorize(apiAuth, deviceAuth))

	mux := http.NewServeMux()
	mux.Handle("/user", webAPIAuthed.Then(routes.User(dbConn)))
	mux.Handle("/secret", webAuthed.Then(routes.GenerateSecret(dbConn)))
	mux.Handle("/latest", webAPIAuthed.Then(routes.GetLatestReadings(dbConn)))
	mux.Handle("/readings", webAPIAuthed.Then(routes.GetReadings(dbConn)))

	mux.Handle("/reading", apiDeviceAuthed.Then(routes.PostReading(dbConn)))

	mux.Handle("/delete_readings", apiAuthed.Then(routes.DeleteReadings(dbConn)))
	mux.Handle("/rotate_secret", apiAuthed.Then(routes.RotateSecret(dbConn)))

	mux.Handle("/authorize", oauth.HandleAuthorize(googleOauth, tokenSource))
	mux.Handle("/oauth2callback", oauth.HandleOAuth2Callback(googleOauth, tokenSource, dbConn))
	mux.Handle("/logout", oauth.LogOut())
	mux.Handle("/demo", oauth.SetDemoUser(c.DemoUser, tokenSource, dbConn))

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.UseHandler(mux)
	n.Run(":8080")
}
