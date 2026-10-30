package validator

import (
	"strings"
	"unicode/utf8"

	"github.com/ESJ0/selava-backend/internal/models"
)

const (
	minPasswordRunes = 8
	maxPasswordBytes = 72 // limite aceptado por bcrypt
)

func ValidateClienteRegistro(req *models.ClienteRegistroRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateNombre("nombre", req.Nombre)...)
	errs = append(errs, validateNombre("apellido", req.Apellido)...)
	errs = append(errs, validateTelefono(req.Telefono)...)
	if strings.TrimSpace(req.Email) == "" {
		errs = append(errs, FieldError{Field: "email", Message: "es requerido"})
	} else {
		errs = append(errs, validateEmail(req.Email)...)
	}
	errs = append(errs, validateDireccion(req.Direccion)...)
	errs = append(errs, validatePassword(req.Password)...)
	return errs
}

func ValidateClienteLogin(req *models.ClienteLoginRequest) ValidationErrors {
	var errs ValidationErrors
	if strings.TrimSpace(req.Email) == "" {
		errs = append(errs, FieldError{Field: "email", Message: "es requerido"})
	} else {
		errs = append(errs, validateEmail(req.Email)...)
	}
	if req.Password == "" {
		errs = append(errs, FieldError{Field: "password", Message: "es requerida"})
	}
	return errs
}

func validatePassword(password string) ValidationErrors {
	var errs ValidationErrors
	if password == "" {
		return append(errs, FieldError{Field: "password", Message: "es requerida"})
	}
	if strings.TrimSpace(password) == "" {
		return append(errs, FieldError{Field: "password", Message: "no puede contener solo espacios"})
	}
	if utf8.RuneCountInString(password) < minPasswordRunes {
		errs = append(errs, FieldError{Field: "password", Message: "debe tener al menos 8 caracteres"})
	}
	if len([]byte(password)) > maxPasswordBytes {
		errs = append(errs, FieldError{Field: "password", Message: "no puede superar los 72 bytes"})
	}
	return errs
}

func ValidatePassword(password string) ValidationErrors { return validatePassword(password) }

func ValidateClientePerfil(req *models.ClientePerfilUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateNombre("nombre", req.Nombre)...)
	errs = append(errs, validateNombre("apellido", req.Apellido)...)
	errs = append(errs, validateTelefono(req.Telefono)...)
	if strings.TrimSpace(req.Email) == "" {
		errs = append(errs, FieldError{Field: "email", Message: "es requerido"})
	} else {
		errs = append(errs, validateEmail(req.Email)...)
	}
	errs = append(errs, validateDireccion(req.Direccion)...)
	return errs
}

func ValidateClientePassword(req *models.ClientePasswordUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.PasswordActual == "" {
		errs = append(errs, FieldError{Field: "password_actual", Message: "es requerida"})
	}
	for _, item := range validatePassword(req.PasswordNueva) {
		item.Field = "password_nueva"
		errs = append(errs, item)
	}
	if req.PasswordActual != "" && req.PasswordActual == req.PasswordNueva {
		errs = append(errs, FieldError{Field: "password_nueva", Message: "debe ser diferente a la contraseña actual"})
	}
	return errs
}
