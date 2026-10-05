package validator

import (
	"math"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidatePagoCreate(req *models.PagoCreateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.MetodoPagoID <= 0 {
		errs = append(errs, FieldError{Field: "metodo_pago_id", Message: "debe ser un entero positivo"})
	}
	if req.Monto <= 0 || math.IsNaN(req.Monto) || math.IsInf(req.Monto, 0) {
		errs = append(errs, FieldError{Field: "monto", Message: "debe ser mayor que cero"})
	} else {
		if req.Monto > 99999999.99 {
			errs = append(errs, FieldError{Field: "monto", Message: "excede el monto maximo permitido"})
		}
		if math.Abs(req.Monto*100-math.Round(req.Monto*100)) > 0.000001 {
			errs = append(errs, FieldError{Field: "monto", Message: "solo puede tener dos decimales"})
		}
	}
	if req.Referencia != nil && len(*req.Referencia) > 100 {
		errs = append(errs, FieldError{Field: "referencia", Message: "no puede superar los 100 caracteres"})
	}
	return errs
}
