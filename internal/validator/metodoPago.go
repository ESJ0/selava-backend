package validator

import (
	"github.com/ESJ0/selava-backend/internal/models"
	"strings"
)

func ValidateMetodoPagoCreate(req *models.MetodoPagoCreateRequest) ValidationErrors {
	return validateNombreMetodoPago(req.Nombre)
}
func ValidateMetodoPagoUpdate(req *models.MetodoPagoUpdateRequest) ValidationErrors {
	if req.Nombre == nil {
		return nil
	}
	return validateNombreMetodoPago(*req.Nombre)
}
func validateNombreMetodoPago(nombre string) ValidationErrors {
	if strings.TrimSpace(nombre) == "" {
		return ValidationErrors{{Field: "nombre", Message: "es requerido"}}
	}
	if len(nombre) > 50 {
		return ValidationErrors{{Field: "nombre", Message: "no puede superar los 50 caracteres"}}
	}
	return nil
}
