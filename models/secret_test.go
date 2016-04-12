package models

import (
	"testing"
)

func TestSecret(t *testing.T) {
	secret, err := NewSecret()
	if err != nil {
		t.Fatal("Error creating Secret %s", err.Error())
	}
	if len(secret) < 44 {
		t.Fatal("Secret is too small")
	}
}

func BenchmarkSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSecret()
	}
}
