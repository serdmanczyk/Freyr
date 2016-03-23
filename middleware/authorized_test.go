package middleware

import (
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/gardenspark/oauth"
	"github.com/serdmanczyk/gardenspark/token"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testKey    = "abracadabra"
	okResponse = "Everything is OK"
	testEmail  = "Ardvark@comeatme.bro"
)

func testToken(tok token.TokenSource) (string, error) {
	expiry := time.Now().Add(time.Minute * 1)
	claims := map[string]interface{}{"email": testEmail}
	return tok.GenerateToken(expiry, claims)
}

func happyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email, _ := ctx.Value("email").(string)
	w.Header().Add("Email", email)
	io.WriteString(w, okResponse)
}

func TestAuthorized(t *testing.T) {
	tokGen := token.JtwTokenGen{[]byte(testKey)}

	handler := apollo.New(Authorized(tokGen)).ThenFunc(happyHandler)

	authorizeRequest, err := http.NewRequest("GET", "/whatever", nil)
	if err != nil {
		t.Fatal(err)
	}

	token, err := testToken(tokGen)
	if err != nil {
		t.Fatal(err)
	}

	authorizeRequest.Header.Add("Cookie", oauth.CookieName+"="+token)
	authorizeResponse := httptest.NewRecorder()

	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != http.StatusOK {
		t.Errorf("Response code not set.  Expected %d, got %d", http.StatusForbidden, authorizeResponse.Code)
	}

	if authorizeResponse.Body.String() != okResponse {
		t.Errorf("Response body not set.  Expected %s, got %s", okResponse, authorizeResponse.Body.String())
	}

	email := authorizeResponse.Header()["Email"]
	if len(email) != 1 || email[0] != testEmail {
		t.Errorf("Email in signed token not made available to context, got %s expected %s", email, testEmail)
	}
}

func TestNotAuthorized(t *testing.T) {
	tokGen := token.JtwTokenGen{[]byte(testKey)}

	handler := apollo.New(Authorized(tokGen)).ThenFunc(happyHandler)

	authorizeRequest, err := http.NewRequest("GET", "/authorize", nil)
	if err != nil {
		t.Fatal(err)
	}

	authorizeResponse := httptest.NewRecorder()

	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != http.StatusForbidden {
		t.Errorf("Response code not set.  Expected %d, got %d", http.StatusForbidden, authorizeResponse.Code)
	}

	if authorizeResponse.Body.String() != CookieNotFound.Error()+"\n" {
		t.Errorf("Response body not set.  Expected %s, got %s", CookieNotFound, authorizeResponse.Body.String())
	}
}
