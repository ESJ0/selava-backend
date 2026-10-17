package validator

import (
	"strings"
	"testing"

	"github.com/ESJ0/selava-backend/internal/models"
)

func registroClienteValido() *models.ClienteRegistroRequest {
	return &models.ClienteRegistroRequest{
		Nombre:   "Ana",
		Apellido: "Lopez",
		Telefono: "5555-1234",
		Email:    "ana@example.com",
		Password: "secreto-seguro",
	}
}

func TestValidateClienteRegistroAceptaPayloadValido(t *testing.T) {
	if errs := ValidateClienteRegistro(registroClienteValido()); errs.HasErrors() {
		t.Fatalf("payload valido rechazado: %v", errs)
	}
}

func TestValidateClienteRegistroRequiereEmailYPasswordSegura(t *testing.T) {
	req := registroClienteValido()
	req.Email = " "
	req.Password = "corta"

	errs := ValidateClienteRegistro(req)
	if len(errs) != 2 || errs[0].Field != "email" || errs[1].Field != "password" {
		t.Fatalf("errores inesperados: %#v", errs)
	}
}

func TestValidateClienteRegistroRespetaLimiteBcrypt(t *testing.T) {
	req := registroClienteValido()
	req.Password = strings.Repeat("a", 73)

	errs := ValidateClienteRegistro(req)
	if len(errs) != 1 || errs[0].Field != "password" {
		t.Fatalf("se esperaba error de password: %#v", errs)
	}
}

func TestValidateClienteRegistroRechazaPasswordDeEspacios(t *testing.T) {
	req := registroClienteValido()
	req.Password = "        "

	errs := ValidateClienteRegistro(req)
	if len(errs) != 1 || errs[0].Field != "password" {
		t.Fatalf("se esperaba error de password: %#v", errs)
	}
}

func TestValidateClienteLoginRequiereCredenciales(t *testing.T) {
	errs := ValidateClienteLogin(&models.ClienteLoginRequest{})
	if len(errs) != 2 {
		t.Fatalf("se esperaban dos errores, se obtuvo %#v", errs)
	}
}
