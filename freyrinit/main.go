package main

import (
	"flag"
	"github.com/serdmanczyk/freyr/models"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"text/template"
)

var (
	domainName  = flag.String("domainName", "localhost", "domain name for running website")
	oauthId     = flag.String("oauthId", "", "Google Oauth Id")
	oauthSecret = flag.String("oauthSecret", "", "Google Oauth Secret")
	demoUser    = flag.String("demouser", "noone@nothing.com", "Demo user account email")
	dbUser      = flag.String("dbuser", "fakeuser", "Postgres database username")
	dbPass      = flag.String("dbpass", "changeme", "Postgres database password")
	force       = flag.Bool("force", false, "Overrite settings files if they exist")
	clean       = flag.Bool("clean", false, "Delete existing .env files")
)

type TemplateConfig struct {
	DomainName  string
	OauthId     string
	OauthSecret string
	DemoUser    string
	DbUser      string
	DbPass      string
	Secret      string
}

type envFile struct {
	name     string
	template string
}

func main() {
	flag.Parse()

	secret, err := models.NewSecret()
	if err != nil {
		log.Fatal(err)
	}

	c := TemplateConfig{
		DomainName:  *domainName,
		OauthId:     *oauthId,
		OauthSecret: *oauthSecret,
		DemoUser:    *demoUser,
		DbUser:      *dbUser,
		DbPass:      *dbPass,
		Secret:      secret.Encode(),
	}

	if Empty(&c) {
		log.Fatal("Must defined a value for all non-defaulted string flags; run 'freyrinit -h' for more info")
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if !isProjectDir(dir) {
		log.Fatal("This must only be run in freyr/freyrinit directory")
	}

	projectdir := filepath.Clean(filepath.Join(dir, ".."))

	for _, f := range []envFile{
		{".env", freyrEnv},
		{"postgres/demo_user.sql", demoUserSQL},
		{"postgres/.env", postgresEnv},
		{"cmd/surtr/.env", surtrEnv},
		{"nginx/conf/nginx.conf", nginxConf},
	} {
		path := filepath.Join(projectdir, f.name)

		if *clean {
			os.Remove(path)
			continue
		}

		if _, err := os.Lstat(path); err == nil && !*force {
			log.Fatalf("%s already exists, cancelling init\n", path)
		}

		writeTemplateFile(path, f.template, c)
	}
}

func writeTemplateFile(path, templ string, c TemplateConfig) {
	tmpl, err := template.New(path).Parse(templ)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tmpl.Execute(f, c)
}

func isProjectDir(path string) bool {
	var dir string
	for _, expected := range []string{
		"freyrinit", "freyr",
	} {
		path = filepath.Clean(path)
		path, dir = filepath.Split(path)
		if dir != expected {
			return false
		}
	}
	return true
}

func Empty(s interface{}) bool {
	val := reflect.ValueOf(s).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		zero := reflect.Zero(field.Type())
		if zero.String() == field.String() {
			return true
		}
	}

	return false
}
