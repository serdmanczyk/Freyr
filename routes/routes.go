package routes

import (
	"github.com/cyclopsci/apollo"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

func getEmail(ctx context.Context) string {
	email, _ := ctx.Value("email").(string)
	return email
}

func Hello() apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		email := getEmail(ctx)
		io.WriteString(w, "hey there: "+email)
		return
	})
}
