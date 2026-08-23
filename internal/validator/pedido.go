package validator

import (
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidatePedidoCreate(req *models.PedidoCreateRequest) ValidationErrors {
	var errs ValidationErrors

	if req.ClienteID <= 0 {
		errs = append(errs, FieldError{Field: "cliente_id", Message: "debe ser un entero positivo"})
	}

	if len(req.Prendas) == 0 {
		errs = append(errs, FieldError{Field: "prendas", Message: "el pedido debe incluir al menos una prenda"})
		return errs
	}

	for i, prenda := range req.Prendas {
		errs = append(errs, validatePrendaCreate(i, prenda)...)
	}

	return errs
}

func validatePrendaCreate(index int, prenda models.PrendaCreateRequest) ValidationErrors {
	var errs ValidationErrors

	if prenda.TipoPrendaID <= 0 {
		errs = append(errs, FieldError{
			Field:   fmt.Sprintf("prendas[%d].tipo_prenda_id", index),
			Message: "debe ser un entero positivo",
		})
	}

	if prenda.Cantidad <= 0 {
		errs = append(errs, FieldError{
			Field:   fmt.Sprintf("prendas[%d].cantidad", index),
			Message: "debe ser mayor a cero",
		})
	}

	if prenda.Color != nil && *prenda.Color == "" {
		errs = append(errs, FieldError{
			Field:   fmt.Sprintf("prendas[%d].color", index),
			Message: "no puede estar vacío",
		})
	}

	return errs
}
