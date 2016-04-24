package routes

import (
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

func GenerateSecret(s models.SecretStore) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		email := getEmail(ctx)

		if r.Method != "GET" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if _, err := s.GetSecret(email); err == nil {
			http.Error(w, "Secret already exists", http.StatusBadRequest)
			return
		}

		secret, err := models.NewSecret()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = s.StoreSecret(email, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		io.WriteString(w, secret.Encode())
		return
	})
}

func RotateSecret(s models.SecretStore) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		email := getEmail(ctx)

		if r.Method != "POST" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		if _, err := s.GetSecret(email); err != nil {
			// TODO: Do more parsing of secret to return if error was actually an internal error
			http.Error(w, "No base secret generated, how did we get here?", http.StatusBadRequest)
		}

		secret, err := models.NewSecret()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = s.StoreSecret(email, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		io.WriteString(w, secret.Encode())
		return
	})
}
