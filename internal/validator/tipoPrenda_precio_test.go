package validator

import (
	"github.com/ESJ0/selava-backend/internal/models"
	"math"
	"testing"
)

func TestTipoPrendaPrecioRequeridoYPrecision(t *testing.T) {
	for _, tc := range []struct {
		name   string
		precio *float64
		valid  bool
	}{
		{"omitido", nil, false}, {"cero", precioPtr(0), false},
		{"negativo", precioPtr(-1), false}, {"tres decimales", precioPtr(25.001), false},
		{"fraccion pequena", precioPtr(0.000001), false}, {"limite", precioPtr(100000000), false},
		{"NaN", precioPtr(math.NaN()), false}, {"infinito", precioPtr(math.Inf(1)), false},
		{"centavo", precioPtr(0.01), true}, {"entero", precioPtr(25), true},
		{"decimal", precioPtr(19.99), true}, {"maximo", precioPtr(99999999.99), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateTipoPrendaCreate(&models.TipoPrendaCreateRequest{Nombre: "Camisa", Precio: tc.precio})
			if errs.HasErrors() == tc.valid {
				t.Fatalf("valido=%v errores=%v", tc.valid, errs)
			}
			if tc.precio != nil {
				update := ValidateTipoPrendaUpdate(&models.TipoPrendaUpdateRequest{Precio: tc.precio})
				if update.HasErrors() == tc.valid {
					t.Fatalf("update valido=%v errores=%v", tc.valid, update)
				}
			}
		})
	}
	if errs := ValidateTipoPrendaUpdate(&models.TipoPrendaUpdateRequest{}); errs.HasErrors() {
		t.Fatal(errs)
	}
}
func precioPtr(value float64) *float64 { return &value }

func TestTipoPrendaModalidadCobro(t *testing.T) {
	price := 25.0
	for _, mode := range []models.ModalidadCobro{models.ModalidadCobroUnidad, models.ModalidadCobroLibra, models.ModalidadCobroMetro} {
		if errs := ValidateTipoPrendaCreate(&models.TipoPrendaCreateRequest{Nombre: "Prenda", Precio: &price, ModalidadCobro: mode}); errs.HasErrors() {
			t.Fatalf("modalidad %s: %v", mode, errs)
		}
	}
	invalid := models.ModalidadCobro("KILO")
	if errs := ValidateTipoPrendaUpdate(&models.TipoPrendaUpdateRequest{ModalidadCobro: &invalid}); !hasValidatorField(errs, "modalidad_cobro") {
		t.Fatal("modalidad invalida aceptada")
	}
}
