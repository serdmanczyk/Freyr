package routes

import (
	"encoding/json"
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"net/http"
)

func getEmail(ctx context.Context) string {
	email, _ := ctx.Value("email").(string)
	return email
}

func User(s models.UserStore) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		email := getEmail(ctx)

		if r.Method != "GET" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		user, err := s.GetUser(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		err = json.NewEncoder(w).Encode(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	})
}
