package validator

import (
	"math"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidateMovimientoInventarioCreate(req *models.MovimientoInventarioCreateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.InsumoID <= 0 {
		errs = append(errs, FieldError{Field: "insumo_id", Message: "debe ser un entero positivo"})
	}
	if req.TipoMovimiento != "entrada" && req.TipoMovimiento != "salida" {
		errs = append(errs, FieldError{Field: "tipo_movimiento", Message: "debe ser entrada o salida"})
	}
	if req.Cantidad <= 0 || math.IsNaN(req.Cantidad) || math.IsInf(req.Cantidad, 0) {
		errs = append(errs, FieldError{Field: "cantidad", Message: "debe ser mayor que cero"})
	} else {
		if req.Cantidad > 99999999.99 {
			errs = append(errs, FieldError{Field: "cantidad", Message: "excede la cantidad maxima permitida"})
		}
		if math.Abs(req.Cantidad*100-math.Round(req.Cantidad*100)) > 0.000001 {
			errs = append(errs, FieldError{Field: "cantidad", Message: "solo puede tener dos decimales"})
		}
	}
	if req.Motivo != nil && len(*req.Motivo) > 255 {
		errs = append(errs, FieldError{Field: "motivo", Message: "no puede superar los 255 caracteres"})
	}
	return errs
}
