package validator

import (
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidateServicioCreate(req *models.ServicioCreateRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateNombreServicio(req.Nombre)...)
	errs = append(errs, validatePrecioServicio(req.PrecioBase)...)
	return errs
}

func ValidateServicioUpdate(req *models.ServicioUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.Nombre != nil {
		errs = append(errs, validateNombreServicio(*req.Nombre)...)
	}
	if req.PrecioBase != nil {
		errs = append(errs, validatePrecioServicio(*req.PrecioBase)...)
	}
	return errs
}

func validateNombreServicio(nombre string) ValidationErrors {
	var errs ValidationErrors
	if strings.TrimSpace(nombre) == "" {
		return append(errs, FieldError{Field: "nombre", Message: "es requerido"})
	}
	if len(nombre) > 100 {
		errs = append(errs, FieldError{Field: "nombre", Message: "no puede superar los 100 caracteres"})
	}
	return errs
}

func validatePrecioServicio(precio float64) ValidationErrors {
	if precio <= 0 {
		return ValidationErrors{{Field: "precio_base", Message: "debe ser mayor que 0"}}
	}
	return nil
}
