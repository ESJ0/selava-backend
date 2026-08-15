package validator

import (
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidateTipoPrendaCreate(req *models.TipoPrendaCreateRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateTipoPrendaNombre(req.Nombre)...)
	if req.Descripcion != nil {
		errs = append(errs, validateTipoPrendaDescripcion(*req.Descripcion)...)
	}
	return errs
}

func ValidateTipoPrendaUpdate(req *models.TipoPrendaUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.Nombre != nil {
		errs = append(errs, validateTipoPrendaNombre(*req.Nombre)...)
	}
	if req.Descripcion != nil {
		errs = append(errs, validateTipoPrendaDescripcion(*req.Descripcion)...)
	}
	return errs
}

func validateTipoPrendaNombre(nombre string) ValidationErrors {
	if strings.TrimSpace(nombre) == "" {
		return ValidationErrors{{Field: "nombre", Message: "es requerido"}}
	}
	if len(nombre) > 50 {
		return ValidationErrors{{Field: "nombre", Message: "no puede superar los 50 caracteres"}}
	}
	return nil
}

func validateTipoPrendaDescripcion(descripcion string) ValidationErrors {
	if len(descripcion) > 255 {
		return ValidationErrors{{Field: "descripcion", Message: "no puede superar los 255 caracteres"}}
	}
	return nil
}
