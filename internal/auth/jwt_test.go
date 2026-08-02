package auth

import (
	"errors"
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken("secret", 7, 2)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ValidateToken("secret", token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims.UsuarioID != 7 {
		t.Fatalf("expected usuario_id 7, got %d", claims.UsuarioID)
	}
	if claims.RolID != 2 {
		t.Fatalf("expected rol_id 2, got %d", claims.RolID)
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		t.Fatal("expected token timestamps to be present")
	}
}

func TestValidateTokenRejectsInvalidToken(t *testing.T) {
	_, err := ValidateToken("secret", "token-invalido")
	if !errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("expected ErrTokenInvalido, got %v", err)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken("secret", 7, 2)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	_, err = ValidateToken("otro-secret", token)
	if !errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("expected ErrTokenInvalido, got %v", err)
	}
}
