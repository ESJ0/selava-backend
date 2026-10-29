package validator

import (
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidateUsuarioCreate(req *models.UsuarioCreateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.RolID != 1 && req.RolID != 2 && req.RolID != 4 {
		errs = append(errs, FieldError{Field: "rol_id", Message: "debe ser Administrador, Recepcionista o Cliente"})
	}
	if strings.TrimSpace(req.Email) == "" {
		errs = append(errs, FieldError{Field: "email", Message: "es requerido"})
	} else {
		errs = append(errs, validateEmail(req.Email)...)
	}
	errs = append(errs, validatePassword(req.Password)...)
	if req.RolID == 2 || req.RolID == 4 {
		errs = append(errs, validateNombre("nombre", req.Nombre)...)
		errs = append(errs, validateNombre("apellido", req.Apellido)...)
		errs = append(errs, validateTelefono(req.Telefono)...)
	}
	if req.RolID == 4 {
		errs = append(errs, validateDireccion(req.Direccion)...)
	}
	return errs
}

func ValidateUsuarioUpdate(req *models.UsuarioUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.Nombre != nil {
		errs = append(errs, validateNombre("nombre", *req.Nombre)...)
	}
	if req.Apellido != nil {
		errs = append(errs, validateNombre("apellido", *req.Apellido)...)
	}
	if req.Email != nil {
		errs = append(errs, validateEmail(*req.Email)...)
	}
	if req.Telefono != nil {
		errs = append(errs, validateTelefono(*req.Telefono)...)
	}
	if req.Direccion != nil {
		errs = append(errs, validateDireccion(*req.Direccion)...)
	}
	return errs
}
