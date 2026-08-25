package config

import (
	"net/url"
	"testing"
)

func TestDSNEscapaCredencialesEspeciales(t *testing.T) {
	cfg := Config{DBHost: "127.0.0.1", DBPort: "5433", DBUser: "postgres", DBPassword: "p@ss:/?#[]", DBName: "lavanderia"}
	parsed, err := url.Parse(cfg.DSN())
	if err != nil {
		t.Fatalf("DSN inválido: %v", err)
	}
	password, ok := parsed.User.Password()
	if !ok || password != cfg.DBPassword {
		t.Fatalf("la contraseña no se preservó")
	}
	if parsed.Hostname() != cfg.DBHost || parsed.Port() != cfg.DBPort {
		t.Fatalf("host inesperado: %s", parsed.Host)
	}
}
