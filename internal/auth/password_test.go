package auth

import "testing"

func TestHashPasswordCreatesVerifiableHash(t *testing.T) {
	hash, err := HashPassword("clave-segura")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "clave-segura" {
		t.Fatal("expected hash to differ from plain password")
	}
	if !CheckPassword(hash, "clave-segura") {
		t.Fatal("expected password to match generated hash")
	}
}

func TestCheckPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("clave-segura")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if CheckPassword(hash, "otra-clave") {
		t.Fatal("expected wrong password to be rejected")
	}
}
