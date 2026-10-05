package validator

import (
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestPedidoListadoDefaultsAndInclusiveDates(t *testing.T) {
	defaults, errs := ParsePedidoList(models.PedidoListRequest{})
	if errs.HasErrors() || defaults.Pagina != 1 || defaults.Limite != 20 || defaults.Orden != "recientes" || defaults.EstadoID != 0 {
		t.Fatalf("defaults=%+v errores=%v", defaults, errs)
	}
	filter, errs := ParsePedidoList(models.PedidoListRequest{Buscar: "  Ana  ", EstadoID: "3", FechaDesde: "2026-10-05", FechaHasta: "2026-10-05", Pagina: "2", Limite: "50", Orden: "antiguos"})
	if errs.HasErrors() || filter.Buscar != "Ana" || filter.EstadoID != 3 || filter.Pagina != 2 || filter.Limite != 50 || filter.Orden != "antiguos" {
		t.Fatalf("filter=%+v errores=%v", filter, errs)
	}
	if filter.FechaDesde.UTC().Format(time.RFC3339) != "2026-10-05T06:00:00Z" || filter.FechaHastaExclusiva.UTC().Format(time.RFC3339) != "2026-10-06T06:00:00Z" {
		t.Fatalf("límites de Guatemala incorrectos: %+v", filter)
	}
}

func TestPedidoListadoInvalidFilters(t *testing.T) {
	for _, tc := range []struct {
		name string
		req  models.PedidoListRequest
	}{
		{"search", models.PedidoListRequest{Buscar: strings.Repeat("á", 101)}},
		{"status zero", models.PedidoListRequest{EstadoID: "0"}},
		{"status text", models.PedidoListRequest{EstadoID: "abc"}},
		{"status overflow", models.PedidoListRequest{EstadoID: "2147483648"}},
		{"page negative", models.PedidoListRequest{Pagina: "-1"}},
		{"page huge", models.PedidoListRequest{Pagina: "1000001"}},
		{"limit zero", models.PedidoListRequest{Limite: "0"}},
		{"limit huge", models.PedidoListRequest{Limite: "101"}},
		{"date impossible", models.PedidoListRequest{FechaDesde: "2026-02-30"}},
		{"year zero", models.PedidoListRequest{FechaHasta: "0000-01-01"}},
		{"date format", models.PedidoListRequest{FechaHasta: "05/10/2026"}},
		{"reverse range", models.PedidoListRequest{FechaDesde: "2026-10-06", FechaHasta: "2026-10-05"}},
		{"sort injection", models.PedidoListRequest{Orden: "DESC; DROP TABLE pedidos"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, errs := ParsePedidoList(tc.req); !errs.HasErrors() {
				t.Fatal("se aceptaron filtros inválidos")
			}
		})
	}
}

func TestPedidoListadoOpenEndedDates(t *testing.T) {
	for _, req := range []models.PedidoListRequest{{FechaDesde: "2026-10-05"}, {FechaHasta: "2026-10-05"}} {
		if _, errs := ParsePedidoList(req); errs.HasErrors() {
			t.Fatal(errs)
		}
	}
}
