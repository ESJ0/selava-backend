package validator

import "github.com/ESJ0/selava-backend/internal/models"

func ValidatePedidoCreate(req *models.PedidoCreateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.ClienteID <= 0 {
		errs = append(errs, FieldError{Field: "cliente_id", Message: "debe ser un entero positivo"})
	}
	return errs
}
