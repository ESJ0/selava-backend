package validator

import (
	"github.com/ESJ0/selava-backend/internal/models"
	"strings"
	"testing"
)

func validPrenda() models.PrendaCreateRequest {
	return models.PrendaCreateRequest{TipoPrendaID: 1, Cantidad: 1, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: 1}}}
}
func TestPedidoColorMaximoTreinta(t *testing.T) {
	p := validPrenda()
	color := strings.Repeat("a", 31)
	p.Color = &color
	errs := ValidatePedidoCreate(&models.PedidoCreateRequest{ClienteID: 1, Prendas: []models.PrendaCreateRequest{p}})
	if !hasValidatorField(errs, "prendas[0].color") {
		t.Fatalf("expected color error: %+v", errs)
	}
}
func TestPedidoRequiereServicioValido(t *testing.T) {
	p := validPrenda()
	p.Servicios = []models.PrendaServicioCreateRequest{{ServicioID: 0}}
	errs := ValidatePedidoCreate(&models.PedidoCreateRequest{ClienteID: 1, Prendas: []models.PrendaCreateRequest{p}})
	if !hasValidatorField(errs, "prendas[0].servicios[0].servicio_id") {
		t.Fatalf("expected service error: %+v", errs)
	}
}
func TestPedidoRechazaServicioDuplicado(t *testing.T) {
	p := validPrenda()
	p.Servicios = append(p.Servicios, models.PrendaServicioCreateRequest{ServicioID: 1})
	errs := ValidatePedidoCreate(&models.PedidoCreateRequest{ClienteID: 1, Prendas: []models.PrendaCreateRequest{p}})
	if len(errs) == 0 {
		t.Fatal("expected duplicate service error")
	}
}
func hasValidatorField(errs ValidationErrors, field string) bool {
	for _, err := range errs {
		if err.Field == field {
			return true
		}
	}
	return false
}
