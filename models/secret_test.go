package models

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
	"testing"
)

const (
	dummySigningContents = "unfenestratedhearthsteadnonjournalisticallywhillikers"
)

func TestSecret(t *testing.T) {
	secret, err := NewSecret()
	if err != nil {
		t.Fatalf("Error creating Secret %s", err.Error())
	}
	if len(secret) < 32 {
		t.Fatal("Secret is too small")
	}
}

func TestValidateSignature(t *testing.T) {
	secret, err := NewSecret()
	if err != nil {
		t.Fatalf("Error creating Secret %s", err.Error())
	}

	signature := secret.Sign(dummySigningContents)
	goodVerified := secret.Verify(dummySigningContents, signature)
	if !goodVerified {
		t.Fatal("String should be verified but isn't")
	}

	badVerified := secret.Verify("bullshit", signature)
	if badVerified {
		t.Fatal("String should not be verified but is")
	}
}

func TestBase64Encoding(t *testing.T) {
	secret, err := NewSecret()
	if err != nil {
		t.Fatalf("Error creating Secret %s", err.Error())
	}

	originalSignature := secret.Sign(dummySigningContents)

	encoded := secret.Encode()
	if encoded == "" {
		t.Fatal("Encoded secret is empty")
	}

	decoded, err := SecretFromBase64(encoded)
	if err != nil {
		t.Fatal("Encoded secret did not re-encode properly")
	}

	newSignature := decoded.Sign(dummySigningContents)
	if originalSignature != newSignature {
		t.Fatal("Signatures inconsistent")
	}

	if !decoded.Verify(dummySigningContents, originalSignature) {
		t.Fatal("Inconsistency in verification between decoded/re-encoded secret")
	}

	if !secret.Verify(dummySigningContents, newSignature) {
		t.Fatal("Inconsistency in verification between decoded/re-encoded secret")
	}
}

func TestBuffer(t *testing.T) {
	signingString := "blahblahblah"
	secret, err := NewSecret()
	if err != nil {
		t.Fatalf("Error creating Secret %s", err.Error())
	}

	signature := secret.Sign(signingString)

	secretString := secret.Encode()

	var dest bytes.Buffer

	reader := strings.NewReader(secretString)

	decoder := base64.NewDecoder(base64.URLEncoding, reader)
	_, err = io.Copy(&dest, decoder)
	if err != nil {
		t.Fatal(err)
	}

	transcribed := Secret(dest.Bytes())
	transSign := transcribed.Sign(signingString)
	if signature != transSign {
		t.Fatalf("Signatures do not match between original and transcribed secret: %s %s", signature, transSign)
	}
}

func BenchmarkSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSecret()
	}
}
