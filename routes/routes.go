package routes

import (
	"encoding/json"
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/bifrost"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
)

func getEmail(ctx context.Context) string {
	email, _ := ctx.Value("email").(string)
	return email
}

// User handles HTTP requests for a user's info.
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

func Jobs(j bifrost.JobDispatcher) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		strJobID := r.FormValue("jobID")
		if strJobID == "" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		jobID, err := strconv.ParseUint(strJobID, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tracker, err := j.JobStatus(uint(jobID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		err = json.NewEncoder(w).Encode(tracker.Status())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
