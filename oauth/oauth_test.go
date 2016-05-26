package oauth

import (
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/token"
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

func TestHandleThreeLegged(t *testing.T) {
	oauth := &fake.Oauth{Email: testEmail}
	tokensource := token.JWTTokenGen(testKey)

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

	callbackURL := "/oauth2callback?state=" + state[0] + "&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := fake.UserStore{testEmail: models.User{Email: testEmail}}
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
	oauth := &fake.Oauth{Email: testEmail}
	tokensource := token.JWTTokenGen(testKey)

	callbackURL := "/oauth2callback?state=shutyomouth&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := fake.UserStore{testEmail: models.User{Email: testEmail}}
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != http.StatusForbidden {
		t.Fatalf("Response should be %d, got: %d", http.StatusForbidden, callbackResponse.Code)
	}
}

func TestExpiredToken(t *testing.T) {
	oauth := &fake.Oauth{Email: testEmail}
	tokensource := token.JWTTokenGen(testKey)

	claims := map[string]interface{}{}
	expiredToken, err := tokensource.GenerateToken(time.Now().Add(time.Second*-1), claims)
	if err != nil {
		t.Fatal(err)
	}

	callbackURL := "/oauth2callback?state=" + expiredToken + "&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := fake.UserStore{testEmail: models.User{Email: testEmail}}
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != http.StatusForbidden {
		t.Fatalf("Response should be %d, got: %d", http.StatusForbidden, callbackResponse.Code)
	}

	if !strings.Contains(callbackResponse.Body.String(), token.ErrorTokenExpired.Error()) {
		t.Fatalf("Error response incorrect, expected %s, got: %s", token.ErrorTokenExpired.Error(), callbackResponse.Body.String())
	}
}

func TestInvalidClaims(t *testing.T) {
	oauth := &fake.Oauth{Email: testEmail}
	tokensource := token.JWTTokenGen(testKey)

	claims := map[string]interface{}{
		"wacko": "blammo",
	}
	expiredToken, err := tokensource.GenerateToken(time.Now().Add(time.Second*5), claims)
	if err != nil {
		t.Fatal(err)
	}

	callbackURL := "/oauth2callback?state=" + expiredToken + "&code=jibbajabba"
	callbackRequest, err := http.NewRequest("GET", callbackURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	callbackResponse := httptest.NewRecorder()

	userStore := fake.UserStore{testEmail: models.User{Email: testEmail}}
	callbackHandler := HandleOAuth2Callback(oauth, tokensource, userStore)
	callbackHandler.ServeHTTP(callbackResponse, callbackRequest)

	if callbackResponse.Code != http.StatusForbidden {
		t.Fatalf("Response should be %d, got: %d", http.StatusForbidden, callbackResponse.Code)
	}

	if !strings.Contains(callbackResponse.Body.String(), ErrorInvalidClaims.Error()) {
		t.Fatalf("Error response incorrect, expected %s, got: %s", ErrorInvalidClaims.Error(), callbackResponse.Body.String())
	}
}
