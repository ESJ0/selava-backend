package validator

import (
	"regexp"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

// Precompiladas al iniciar el programa: evita el costo de compilar el
// regex en cada request, importante porque esto corre en el flujo de venta.
var (
	nombreRegex   = regexp.MustCompile(`^[\p{L}\s'-]+$`)
	telefonoRegex = regexp.MustCompile(`^\+?[0-9\s-]{8,15}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

const (
	maxNombreLen    = 100
	maxTelefonoLen  = 20
	maxEmailLen     = 150
	maxDireccionLen = 255
)

func ValidateClienteCreate(req *models.ClienteCreateRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateNombre("nombre", req.Nombre)...)
	errs = append(errs, validateNombre("apellido", req.Apellido)...)
	errs = append(errs, validateTelefono(req.Telefono)...)
	errs = append(errs, validateEmail(req.Email)...)
	errs = append(errs, validateDireccion(req.Direccion)...)
	return errs
}

func ValidateClienteUpdate(req *models.ClienteUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.Nombre != nil {
		errs = append(errs, validateNombre("nombre", *req.Nombre)...)
	}
	if req.Apellido != nil {
		errs = append(errs, validateNombre("apellido", *req.Apellido)...)
	}
	if req.Telefono != nil {
		errs = append(errs, validateTelefono(*req.Telefono)...)
	}
	if req.Email != nil {
		errs = append(errs, validateEmail(*req.Email)...)
	}
	if req.Direccion != nil {
		errs = append(errs, validateDireccion(*req.Direccion)...)
	}
	return errs
}

func validateNombre(campo, valor string) ValidationErrors {
	var errs ValidationErrors
	valor = strings.TrimSpace(valor)
	if valor == "" {
		errs = append(errs, FieldError{Field: campo, Message: "es requerido"})
		return errs
	}
	if len(valor) > maxNombreLen {
		errs = append(errs, FieldError{Field: campo, Message: "no puede superar los 100 caracteres"})
	}
	if !nombreRegex.MatchString(valor) {
		errs = append(errs, FieldError{Field: campo, Message: "solo puede contener letras, espacios, apóstrofes y guiones"})
	}
	return errs
}

func validateTelefono(valor string) ValidationErrors {
	var errs ValidationErrors
	valor = strings.TrimSpace(valor)
	if valor == "" {
		errs = append(errs, FieldError{Field: "telefono", Message: "es requerido"})
		return errs
	}
	if len(valor) > maxTelefonoLen {
		errs = append(errs, FieldError{Field: "telefono", Message: "no puede superar los 20 caracteres"})
	}
	if !telefonoRegex.MatchString(valor) {
		errs = append(errs, FieldError{Field: "telefono", Message: "debe tener entre 8 y 15 dígitos, puede incluir +, espacios o guiones"})
	}
	return errs
}

// El email es opcional en el modelo, así que un valor vacío no genera error.
func validateEmail(valor string) ValidationErrors {
	var errs ValidationErrors
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return errs
	}
	if len(valor) > maxEmailLen {
		errs = append(errs, FieldError{Field: "email", Message: "no puede superar los 150 caracteres"})
	}
	if !emailRegex.MatchString(valor) {
		errs = append(errs, FieldError{Field: "email", Message: "formato de email inválido"})
	}
	return errs
}

func validateDireccion(valor string) ValidationErrors {
	var errs ValidationErrors
	if len(valor) > maxDireccionLen {
		errs = append(errs, FieldError{Field: "direccion", Message: "no puede superar los 255 caracteres"})
	}
	return errs
}
