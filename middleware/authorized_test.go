package middleware

import (
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/oauth"
	"github.com/serdmanczyk/freyr/token"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	okResponse = "Everything is OK"
	testEmail  = "Ardvark@comeatme.bro"
)

func happyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email, _ := ctx.Value("email").(string)
	w.Header().Add("Email", email)
	io.WriteString(w, okResponse)
}

func TestWebAuthorized(t *testing.T) {
	secret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	tokGen := token.JtwTokenGen(secret)
	uA := NewWebAuthorizer(tokGen)

	handler := apollo.New(Authorize(uA)).ThenFunc(happyHandler)

	authorizeRequest, err := http.NewRequest("GET", "/whatever", nil)
	if err != nil {
		t.Fatal(err)
	}

	token, err := token.GenerateWebToken(tokGen, time.Now().Add(time.Second*1), testEmail)
	if err != nil {
		t.Fatal(err)
	}

	authorizeRequest.Header.Add("Cookie", oauth.CookieName+"="+token)
	authorizeResponse := httptest.NewRecorder()

	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != http.StatusOK {
		t.Errorf("Response code incorrect.  Expected %d, got %d", http.StatusOK, authorizeResponse.Code)
	}

	if authorizeResponse.Body.String() != okResponse {
		t.Errorf("Response body not set.  Expected %s, got %s", okResponse, authorizeResponse.Body.String())
	}

	email := authorizeResponse.Header().Get("Email")
	if email != testEmail {
		t.Errorf("Email in signed token not made available to context, got %s expected %s", email, testEmail)
	}
}

func TestWebNotAuthorized(t *testing.T) {
	secret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	tokGen := token.JtwTokenGen(secret)
	uA := NewWebAuthorizer(tokGen)

	handler := apollo.New(Authorize(uA)).ThenFunc(happyHandler)

	authorizeRequest, err := http.NewRequest("GET", "/authorize", nil)
	if err != nil {
		t.Fatal(err)
	}

	authorizeResponse := httptest.NewRecorder()

	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != http.StatusUnauthorized {
		t.Errorf("Response code not set.  Expected %d, got %d", http.StatusUnauthorized, authorizeResponse.Code)
	}
}

func TestApiAuthorizer(t *testing.T) {
	secret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	userEmail := "badwolf@galifrey.unv"

	ss := fake.SecretStore{userEmail: secret}
	aa := NewApiAuthorizer(ss)

	authorizeRequest, err := http.NewRequest("GET", "/authorize", nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	SignRequest(secret, userEmail, authorizeRequest)

	handler := apollo.New(Authorize(aa)).ThenFunc(happyHandler)

	authorizeResponse := httptest.NewRecorder()

	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != http.StatusOK {
		t.Errorf("Response code incorrect.  Expected %d, got %d", http.StatusOK, authorizeResponse.Code)
	}

	if authorizeResponse.Body.String() != okResponse {
		t.Errorf("Response body not set.  Expected %s, got %s", okResponse, authorizeResponse.Body.String())
	}

	email := authorizeResponse.Header().Get("Email")
	if email != userEmail {
		t.Errorf("Email in signed token not made available to context, got %s expected %s", email, testEmail)
	}
}

func TestDeviceAuthorizer(t *testing.T) {
	secret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	userEmail := "badwolf@galifrey.unv"
	coreId := "53ff76065075535110341387"

	ss := fake.SecretStore{userEmail: secret}
	da := NewDeviceAuthorizer(ss)

	body := strings.NewReader("event=post_reading&data=%7B%20%22temperature%22%3A%2019.800%2C%20%22humidity%22%3A%2057.300%2C%20%22moisture%22%3A%200000%2C%20%22light%22%3A%201.000%20%7D&published_at=2016-04-20T04%3A32%3A52.962Z&coreid=" + coreId)
	authorizeRequest, err := http.NewRequest("POST", "/authorize", body)
	if err != nil {
		t.Errorf(err.Error())
	}

	authorizeRequest.Header.Add(AuthTypeHeader, DeviceAuthTypeValue)
	authorizeRequest.Header.Add(AuthUserHeader, userEmail)
	authorizeRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	token, err := token.GenerateDeviceToken(token.JtwTokenGen(secret), time.Now().Add(time.Second), coreId, userEmail)
	if err != nil {
		t.Fatal(err)
	}

	authorizeRequest.Header.Add(TokenHeader, token)

	authorizeResponse := httptest.NewRecorder()

	handler := apollo.New(Authorize(da)).ThenFunc(happyHandler)
	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != http.StatusOK {
		t.Errorf("Response code incorrect.  Expected %d, got %d", http.StatusOK, authorizeResponse.Code)
	}

	if authorizeResponse.Body.String() != okResponse {
		t.Errorf("Response body not set.  Expected %s, got %s", okResponse, authorizeResponse.Body.String())
	}

	email := authorizeResponse.Header().Get("Email")
	if email != userEmail {
		t.Errorf("Email in signed token not made available to context, got %s expected %s", email, testEmail)
	}
}
