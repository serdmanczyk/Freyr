package middleware

import (
	"github.com/cyclopsci/apollo"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/oauth"
	"github.com/serdmanczyk/freyr/token"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	okResponse = "Everything is OK"
	testEmail  = "Ardvark@comeatme.bro"
)

// TODO: stole this from token package tests, figure out way to share
type fakeSecretStore map[string]models.Secret

func (s fakeSecretStore) GetSecret(userEmail string) (models.Secret, error) {
	secret, ok := s[userEmail]
	if !ok {
		return models.Secret([]byte{}), models.SecretDoesntExist
	}
	return secret, nil
}

func (s fakeSecretStore) StoreSecret(userEmail string, secret models.Secret) error {
	s[userEmail] = secret
	return nil
}

func testToken(tok token.TokenSource) (string, error) {
	expiry := time.Now().Add(time.Second * 1)
	claims := map[string]interface{}{"email": testEmail}
	return tok.GenerateToken(expiry, claims)
}

func happyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	email, _ := ctx.Value("email").(string)
	w.Header().Add("Email", email)
	io.WriteString(w, okResponse)
}

func TestUserAuthorized(t *testing.T) {
	secret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	tokGen := token.JtwTokenGen(secret)
	uA := NewUserAuthorizer(tokGen)

	handler := apollo.New(Authorize(uA)).ThenFunc(happyHandler)

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

func TestUserNotAuthorized(t *testing.T) {
	secret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	tokGen := token.JtwTokenGen(secret)
	uA := NewUserAuthorizer(tokGen)

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

	ss := fakeSecretStore{userEmail: secret}
	aa := NewApiAuthorizer(ss)

	authorizeRequest, err := http.NewRequest("GET", "/authorize", nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	authorizeRequest.Header.Add(AuthTypeHeader, ApiAuthTypeValue)
	authorizeRequest.Header.Add(AuthUserHeader, userEmail)
	n := time.Now().Unix()
	unixStamp := strconv.FormatInt(n, 10)
	authorizeRequest.Header.Add(ApiAuthDateHeader, unixStamp)

	_, signingString := apiSigningString(authorizeRequest)
	if signingString == "" {
		t.Fatal("Failure generating signing string")
	}

	signature := secret.Sign(signingString)
	authorizeRequest.Header.Add(ApiSignatureHeader, signature)

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

	ss := fakeSecretStore{userEmail: secret}
	da := NewDeviceAuthorizer(ss)

	body := strings.NewReader("event=post_reading&data=%7B%20%22temperature%22%3A%2019.800%2C%20%22humidity%22%3A%2057.300%2C%20%22moisture%22%3A%200000%2C%20%22light%22%3A%201.000%20%7D&published_at=2016-04-20T04%3A32%3A52.962Z&coreid=" + coreId)
	authorizeRequest, err := http.NewRequest("POST", "/authorize", body)
	if err != nil {
		t.Errorf(err.Error())
	}

	authorizeRequest.Header.Add(AuthTypeHeader, DeviceAuthTypeValue)
	authorizeRequest.Header.Add(AuthUserHeader, userEmail)
	authorizeRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tokGen := token.JtwTokenGen(secret)
	token, err := tokGen.GenerateToken(time.Now().Add(time.Second), token.Claims{
		"coreid": coreId,
		"email":  userEmail,
	})
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
