package routes

import (
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/gardenspark/models"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

const (
	signatureHeader = "X-GSPK-Signature"
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

func getHMACSignature(r *http.Request) string {
	signature := r.Header.Get(signatureHeader)
	return signature
}

func GenerateToken(s models.SecretStore) apollo.Handler {
	return apollo.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		//email := getEmail(ctx)

		//requestSignature := r.Header.Get(signatureHeader)

		//requestSigningString := "pending imp"
		// requestSigningString := signingString(r)

		//if requestSignature == "" && current_secret == "" || current_secret.ValidateSignature(requestSigningString, requestSignature) {
		//	newSecret, err := models.NewSecret()
		//	if err != nil {
		//		http.Error(w, err.Error(), http.StatusInternalServerError)
		//		return
		//	}
		//
		//	err = s.StoreSecret(email, newSecret)
		//	if err != nil {
		//		http.Error(w, err.Error(), http.StatusInternalServerError)
		//		return
		//	}
		//
		//	io.WriteString(w, string(newSecret))
		//	return
		//}

		http.Error(w, "Invalid request... sorry that's all you get", http.StatusForbidden)
		return
	})
}
