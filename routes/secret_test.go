package routes

import (
	"github.com/serdmanczyk/freyr/fake"
	"github.com/serdmanczyk/freyr/models"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateSecret(t *testing.T) {
	userEmail := "Yggdrasil@nine.worlds"
	fakeSs := fake.SecretStore{}

	genReq, err := http.NewRequest("GET", "/generate_secret", nil)
	if err != nil {
		t.Fatal(err)
	}

	genResp := httptest.NewRecorder()

	emCtx := context.WithValue(context.Background(), "email", userEmail)

	secretHandler := GenerateSecret(fakeSs)
	secretHandler.ServeHTTP(emCtx, genResp, genReq)

	if genResp.Code != 200 {
		t.Fatalf("Response should be 200, got: %d:%s", genResp.Code, genResp.Body.String())
	}

	secretBody, err := ioutil.ReadAll(genResp.Body)
	if err != nil {
		t.Fatal(err)
	}

	storedSecret, err := fakeSs.GetSecret(userEmail)
	if err != nil {
		t.Fatal(err)
	}

	got, expected := string(secretBody), storedSecret.Encode()
	if got != expected {
		t.Fatalf("Incorrect secret returned; got %s, expected %s", got, expected)
	}

	genInvalidResp := httptest.NewRecorder()
	secretHandler.ServeHTTP(emCtx, genInvalidResp, genReq)

	if genInvalidResp.Code != 400 {
		t.Fatalf("Response should be 400, got: %d:%s", genResp.Code, genResp.Body.String())
	}
}

func TestRotateSecret(t *testing.T) {
	userEmail := "Yggdrasil@nine.worlds"
	fakeSs := fake.SecretStore{}

	rotReq, err := http.NewRequest("POST", "/rotate_secret", nil)
	if err != nil {
		t.Fatal(err)
	}

	rotInvalidResp := httptest.NewRecorder()

	emCtx := context.WithValue(context.Background(), "email", userEmail)

	secretHandler := RotateSecret(fakeSs)
	secretHandler.ServeHTTP(emCtx, rotInvalidResp, rotReq)

	if rotInvalidResp.Code != 400 {
		t.Fatalf("Response should be 400, got: %d:%s", rotInvalidResp.Code, rotInvalidResp.Body.String())
	}

	firstSecret, err := models.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	fakeSs.StoreSecret(userEmail, firstSecret)

	rotGoodResp := httptest.NewRecorder()

	secretHandler.ServeHTTP(emCtx, rotGoodResp, rotReq)

	if rotGoodResp.Code != 200 {
		t.Fatalf("Response should be 200, got: %d:%s", rotGoodResp.Code, rotGoodResp.Body.String())
	}

	secretBody, err := ioutil.ReadAll(rotGoodResp.Body)
	if err != nil {
		t.Fatal(err)
	}

	base64returned := string(secretBody)
	if base64returned == firstSecret.Encode() {
		t.Fatal("Got back old secret, should have been rotated")
	}

	currentSecret, err := fakeSs.GetSecret(userEmail)
	if err != nil {
		t.Fatal(err)
	}

	if base64returned != currentSecret.Encode() {
		t.Fatalf("Returned secret doesn't match secret in store; got %s, expected %s", base64returned, currentSecret.Encode())
	}
}
