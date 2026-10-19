package validator

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

func ValidateTipoPrendaCreate(req *models.TipoPrendaCreateRequest) ValidationErrors {
	var errs ValidationErrors
	errs = append(errs, validateTipoPrendaNombre(req.Nombre)...)
	if req.Precio == nil {
		errs = append(errs, FieldError{Field: "precio", Message: "es requerido"})
	} else {
		errs = append(errs, validatePrecioTipoPrenda(*req.Precio)...)
	}
	if req.Descripcion != nil {
		errs = append(errs, validateTipoPrendaDescripcion(*req.Descripcion)...)
	}
	return errs
}

func ValidateTipoPrendaUpdate(req *models.TipoPrendaUpdateRequest) ValidationErrors {
	var errs ValidationErrors
	if req.Precio != nil {
		errs = append(errs, validatePrecioTipoPrenda(*req.Precio)...)
	}
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

func validatePrecioTipoPrenda(precio float64) ValidationErrors {
	if math.IsNaN(precio) || math.IsInf(precio, 0) || precio <= 0 || precio > 99999999.99 {
		return ValidationErrors{{Field: "precio", Message: "debe ser mayor que 0 y no superar 99999999.99"}}
	}
	// Validar el decimal representable sin tolerancias que admitan fracciones de centavo.
	decimal, ok := new(big.Rat).SetString(strconv.FormatFloat(precio, 'f', -1, 64))
	if !ok || !decimal.Mul(decimal, big.NewRat(100, 1)).IsInt() {
		return ValidationErrors{{Field: "precio", Message: "debe tener como maximo 2 decimales"}}
	}
	return nil
}
