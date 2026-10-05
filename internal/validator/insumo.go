package validator

import (
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidateInsumoCreate(req *models.InsumoCreateRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateInsumoNombre(req.Nombre)...)
	errs = append(errs, validateInsumoDescripcion(req.Descripcion)...)
	errs = append(errs, validateInsumoUnidadMedida(req.UnidadMedida)...)
	errs = append(errs, validateInsumoStock("stock_actual", req.StockActual)...)
	errs = append(errs, validateInsumoStock("stock_minimo", req.StockMinimo)...)
	return errs
}

func ValidateInsumoUpdate(req *models.InsumoUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.Nombre != nil {
		errs = append(errs, validateInsumoNombre(*req.Nombre)...)
	}
	if req.Descripcion != nil {
		errs = append(errs, validateInsumoDescripcion(req.Descripcion)...)
	}
	if req.UnidadMedida != nil {
		errs = append(errs, validateInsumoUnidadMedida(*req.UnidadMedida)...)
	}
	if req.StockActual != nil {
		errs = append(errs, validateInsumoStock("stock_actual", *req.StockActual)...)
	}
	if req.StockMinimo != nil {
		errs = append(errs, validateInsumoStock("stock_minimo", *req.StockMinimo)...)
	}
	return errs
}

func validateInsumoNombre(nombre string) ValidationErrors {
	if strings.TrimSpace(nombre) == "" {
		return ValidationErrors{{Field: "nombre", Message: "es requerido"}}
	}
	if len(nombre) > 100 {
		return ValidationErrors{{Field: "nombre", Message: "no puede superar los 100 caracteres"}}
	}
	return nil
}

func validateInsumoDescripcion(descripcion *string) ValidationErrors {
	if descripcion != nil && len(*descripcion) > 255 {
		return ValidationErrors{{Field: "descripcion", Message: "no puede superar los 255 caracteres"}}
	}
	return nil
}

func validateInsumoUnidadMedida(unidad string) ValidationErrors {
	if strings.TrimSpace(unidad) == "" {
		return ValidationErrors{{Field: "unidad_medida", Message: "es requerida"}}
	}
	if len(unidad) > 20 {
		return ValidationErrors{{Field: "unidad_medida", Message: "no puede superar los 20 caracteres"}}
	}
	return nil
}

func validateInsumoStock(field string, stock float64) ValidationErrors {
	if stock < 0 {
		return ValidationErrors{{Field: field, Message: "no puede ser negativo"}}
	}
	return nil
}
