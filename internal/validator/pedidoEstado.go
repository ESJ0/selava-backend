package validator

import "github.com/ESJ0/selava-backend/internal/models"

func ValidatePedidoEstadoUpdate(req *models.PedidoEstadoUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.EstadoID <= 0 {
		errs = append(errs, FieldError{Field: "estado_id", Message: "debe ser un entero positivo"})
	}
	if req.Observaciones != nil && len(*req.Observaciones) > 255 {
		errs = append(errs, FieldError{Field: "observaciones", Message: "no puede superar los 255 caracteres"})
	}
	return errs
}
