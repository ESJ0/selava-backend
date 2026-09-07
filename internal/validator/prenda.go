package validator

import (
	"fmt"
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
		if prendas[i].Cantidad < 1 {
			errs = append(errs, FieldError{Field: campo("cantidad"), Message: "debe ser un entero positivo"})
		}
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
		if prendas[i].Cantidad == 0 {
			prendas[i].Cantidad = 1
		}
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
