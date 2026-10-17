package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCuentaClienteNoExponePasswordHash(t *testing.T) {
	cuenta := CuentaCliente{
		ID:           1,
		ClienteID:    7,
		Email:        "cliente@selava.test",
		PasswordHash: "hash-que-no-debe-salir",
		Activo:       true,
	}

	payload, err := json.Marshal(cuenta)
	if err != nil {
		t.Fatalf("serializando cuenta: %v", err)
	}
	if strings.Contains(string(payload), "password") || strings.Contains(string(payload), cuenta.PasswordHash) {
		t.Fatalf("la respuesta JSON expuso credenciales: %s", payload)
	}
}

func TestCuentaClienteTieneEmailVerificado(t *testing.T) {
	cuenta := CuentaCliente{}
	if cuenta.TieneEmailVerificado() {
		t.Fatal("una cuenta sin fecha no debe aparecer verificada")
	}

	ahora := time.Now()
	cuenta.EmailVerificadoAt = &ahora
	if !cuenta.TieneEmailVerificado() {
		t.Fatal("una cuenta con fecha debe aparecer verificada")
	}
}
