package oauth

import (
	"github.com/serdmanczyk/gardenspark/models"
	"github.com/serdmanczyk/gardenspark/token"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	testKey   = "tokenkeytokenkeytokenkey"
	testEmail = "Ardvark@comeatme.bro"
)

type fakeUserStore struct{}

func newfakeUserStore() *fakeUserStore {
	return &fakeUserStore{}
}

func (f *fakeUserStore) HasUser() bool {
	return true
}

func (f *fakeUserStore) GetUser(email string) (models.User, error) {
	return models.User{}, nil
}

func (f *fakeUserStore) AddUser(models.User) error {
	return nil
}

type fakeOauth struct{}

func (f *fakeOauth) GetRedirectURL(csrftoken string) string {
	return "/someurl?state=" + csrftoken
}

func (f *fakeOauth) GetCallbackCsrfToken(r *http.Request) string {
	return r.FormValue("state")
}

func (f *fakeOauth) GetExchangeToken(r *http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (f *fakeOauth) GetUserData(tok *oauth2.Token) (*models.User, error) {
	return &models.User{Email: testEmail}, nil
}

func TestHandleThreeLegged(t *testing.T) {
	oauth := &fakeOauth{}
	tokensource := token.JtwTokenGen{[]byte(testKey)}

	authorizeRequest, err := http.NewRequest("GET", "/authorize", nil)
	if err != nil {
		t.Fatal(err)
	}
	authorizeResponse := httptest.NewRecorder()

	authorizeHandler := HandleAuthorize(oauth, tokensource)
	authorizeHandler.ServeHTTP(authorizeResponse, authorizeRequest)

	if authorizeResponse.Code != 302 {
		t.Fatalf("Response should be 302, got: %d", authorizeResponse.Code)
	}

	authorizeLocation, ok := authorizeResponse.HeaderMap["Location"]
	if !ok || len(authorizeLocation) != 1 {
		t.Fatalf("Location header not set and/or incorrect number of values")
	}

	u, err := url.Parse(authorizeLocation[0])
	if err != nil {
		t.Fatalf("Redirect header invalid: %s", err.Error())
	}

	state, ok := u.Query()["state"]
	if !ok || len(state) != 1 {
		t.Fatalf("State parameter not set in redirect url or set incorrect number of times")
	}

	claims, err := tokensource.ValidateToken(state[0])
	if err != nil {
		t.Fatalf("State header set in redirect url invalid: %s", err.Error())
	}

	if !checkOauthClaim(claims) {
		t.Fatalf("Passed claim in csrf token invalid: %v", claims)
	}

	callbackUrl := "/oauth2callback?state=" + state[0] + "&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := newfakeUserStore()
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != 302 {
		t.Fatalf("Response should be 302, got: %d:%s", callbackResponse.Code, callbackResponse.Body.String())
	}

	callbackLocation, ok := callbackResponse.HeaderMap["Location"]
	if !ok || len(callbackLocation) != 1 {
		t.Fatalf("Location header not set and/or incorrect number of values")
	}

	if callbackLocation[0] != "/" {
		t.Fatalf("Callback should redirect to '/', got: %s", callbackLocation[0])
	}

	cookie, ok := callbackResponse.Header()["Set-Cookie"]
	if !ok || len(cookie) != 1 {
		t.Fatalf("Cookie not set in response, or set improper amount of times")
	}

	header := http.Header{}
	header.Add("Cookie", cookie[0])
	request := http.Request{Header: header}
	parsedCookie, err := request.Cookie(CookieName)
	if err != nil {
		t.Fatalf("Cookie not properly set in response: %s", parsedCookie)
	}

	claims, err = tokensource.ValidateToken(parsedCookie.Value)
	if err != nil {
		t.Fatalf("Cookie not properly set in response: %s", err.Error())
	}

	email, ok := claims["email"]
	if !ok {
		t.Fatalf("email not set in claim")
	}
	if email != testEmail {
		t.Fatalf("incorrect email in claim")
	}
}

// TODO: refactor following into table test??

func TestRejectToken(t *testing.T) {
	oauth := &fakeOauth{}
	tokensource := token.JtwTokenGen{[]byte(testKey)}

	callbackUrl := "/oauth2callback?state=shutyomouth&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := newfakeUserStore()
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != http.StatusForbidden {
		t.Fatalf("Response should be %d, got: %d", http.StatusForbidden, callbackResponse.Code)
	}
}

func TestExpiredToken(t *testing.T) {
	oauth := &fakeOauth{}
	tokensource := token.JtwTokenGen{[]byte(testKey)}

	claims := map[string]interface{}{}
	expiredToken, err := tokensource.GenerateToken(time.Now().Add(time.Second*-1), claims)
	if err != nil {
		t.Fatal(err)
	}

	callbackUrl := "/oauth2callback?state=" + expiredToken + "&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := newfakeUserStore()
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != http.StatusForbidden {
		t.Fatalf("Response should be %d, got: %d", http.StatusForbidden, callbackResponse.Code)
	}

	if !strings.Contains(callbackResponse.Body.String(), token.TokenExpired.Error()) {
		t.Fatalf("Error response incorrect, expected %s, got: %s", token.TokenExpired.Error(), callbackResponse.Body.String())
	}
}

func TestInvalidClaims(t *testing.T) {
	oauth := &fakeOauth{}
	tokensource := token.JtwTokenGen{[]byte(testKey)}

	claims := map[string]interface{}{
		"wacko": "blammo",
	}
	expiredToken, err := tokensource.GenerateToken(time.Now().Add(time.Second*5), claims)
	if err != nil {
		t.Fatal(err)
	}

	callbackUrl := "/oauth2callback?state=" + expiredToken + "&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackUrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := newfakeUserStore()
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != http.StatusForbidden {
		t.Fatalf("Response should be %d, got: %d", http.StatusForbidden, callbackResponse.Code)
	}

	if !strings.Contains(callbackResponse.Body.String(), InvalidClaims.Error()) {
		t.Fatalf("Error response incorrect, expected %s, got: %s", InvalidClaims.Error(), callbackResponse.Body.String())
	}
}
