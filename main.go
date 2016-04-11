package main

import (
	"flag"
	"github.com/codegangsta/negroni"
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/gardenspark/database"
	"github.com/serdmanczyk/gardenspark/envflags"
	"github.com/serdmanczyk/gardenspark/middleware"
	"github.com/serdmanczyk/gardenspark/oauth"
	"github.com/serdmanczyk/gardenspark/token"
	"golang.org/x/net/context"
	"io"
	"log"
	"net/http"
	"os"
)

type Config struct {
	OauthClientId     string `flag:"oauthClientId" env:"GSPK_OAUTHID"`
	OauthClientSecret string `flag:"oauthClientSecret" env:"GSPK_OAUTHSECRET"`
	Domain            string `flag:"domain" env:"GSPK_DOMAIN"`
	SecretKey         string `flag:"secretkey" env:"GSPK_SECRET"`
	DBHost            string `flag:"dbhost" env:"GSPK_DBHOST"`
	DBUser            string `flag:"dbuser" env:"GSPK_DBUSER"`
	DBPassword        string `flag:"dbpassw" env:"GSPK_DBPASSW"`
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

	authed := apollo.New(middleware.Authorized(tokenSource))

	mux := http.NewServeMux()
	mux.Handle("/", authed.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		email, _ := ctx.Value("email").(string)
		io.WriteString(w, "hey there: "+email)
		return
	}))
	mux.Handle("/authorize", oauth.HandleAuthorize(googleOauth, tokenSource))
	mux.Handle("/oauth2callback", oauth.HandleOAuth2Callback(googleOauth, tokenSource, dbConn))

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.UseHandler(mux)
	n.Run(":8080")
}
