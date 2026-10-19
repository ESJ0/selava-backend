package validator

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/ESJ0/selava-backend/internal/models"
)

const (
	maxPrendaDescripcionLen = 255
	maxPrendaColorLen       = 30
)

func ValidatePrendasCreate(prendas []models.PrendaCreateRequest) ValidationErrors {
	var errs ValidationErrors
	if len(prendas) == 0 {
		return append(errs, FieldError{Field: "prendas", Message: "debe incluir al menos una prenda"})
	}

	for i := range prendas {
		campo := func(nombre string) string {
			return fmt.Sprintf("prendas[%d].%s", i, nombre)
		}

		if prendas[i].TipoPrendaID <= 0 {
			errs = append(errs, FieldError{Field: campo("tipo_prenda_id"), Message: "debe ser un entero positivo"})
		}
		errs = append(errs, validateMedidaPrenda(campo("cantidad"), prendas[i].Cantidad)...)
		if prendas[i].Descripcion != nil && len(*prendas[i].Descripcion) > maxPrendaDescripcionLen {
			errs = append(errs, FieldError{Field: campo("descripcion"), Message: "no puede superar los 255 caracteres"})
		}
		if prendas[i].Color != nil && len(*prendas[i].Color) > maxPrendaColorLen {
			errs = append(errs, FieldError{Field: campo("color"), Message: "no puede superar los 30 caracteres"})
		}
	}
	return errs
}

func NormalizePrendasCreate(prendas []models.PrendaCreateRequest) {
	for i := range prendas {
		if prendas[i].Descripcion != nil {
			valor := strings.TrimSpace(*prendas[i].Descripcion)
			prendas[i].Descripcion = &valor
		}
		if prendas[i].Color != nil {
			valor := strings.TrimSpace(*prendas[i].Color)
			prendas[i].Color = &valor
		}
	}
}

func validateMedidaPrenda(field string, medida float64) ValidationErrors {
	if math.IsNaN(medida) || math.IsInf(medida, 0) || medida <= 0 || medida > 99999999.999 {
		return ValidationErrors{{Field: field, Message: "debe ser mayor que 0 y no superar 99999999.999"}}
	}
	decimal, ok := new(big.Rat).SetString(strconv.FormatFloat(medida, 'f', -1, 64))
	if !ok || !decimal.Mul(decimal, big.NewRat(1000, 1)).IsInt() {
		return ValidationErrors{{Field: field, Message: "debe tener como maximo 3 decimales"}}
	}
	return nil
}
