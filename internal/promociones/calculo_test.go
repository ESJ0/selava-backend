package promociones

import (
	"github.com/ESJ0/selava-backend/internal/models"
	"testing"
	"time"
)

func TestAplicacionesYSobrantes(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name                                                        string
		mode                                                        models.ModalidadCobro
		required, price, measure, normal, total, covered, remainder float64
		applications                                                int64
	}{
		{"unidad_insuficiente", models.ModalidadCobroUnidad, 2, 40, 1, 25, 0, 0, 0, 0},
		{"unidad_un_paquete", models.ModalidadCobroUnidad, 2, 40, 2, 25, 40, 2, 0, 1},
		{"unidad_cinco", models.ModalidadCobroUnidad, 2, 40, 5, 25, 105, 4, 1, 2},
		{"unidad_siete", models.ModalidadCobroUnidad, 2, 40, 7, 25, 145, 6, 1, 3},
		{"libra_un_paquete", models.ModalidadCobroLibra, 5, 50, 5, 12, 50, 5, 0, 1},
		{"libra_varios", models.ModalidadCobroLibra, 5, 50, 12.5, 12, 130, 10, 2.5, 2},
		{"libra_milesimas", models.ModalidadCobroLibra, 0.1, 1, 0.3, 12, 3, 0.3, 0, 3},
		{"metro_un_paquete", models.ModalidadCobroMetro, 2, 60, 2, 40, 60, 2, 0, 1},
		{"metro_varios", models.ModalidadCobroMetro, 2, 60, 5.5, 40, 180, 4, 1.5, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := models.Promocion{ID: 1, PromocionRequest: models.PromocionRequest{Nombre: "Paquete", TipoPrendaID: 3, ModalidadCobro: tc.mode, MedidaRequerida: tc.required, PrecioPromocional: tc.price, Activo: true, FechaInicio: now.Add(-time.Hour), FechaFin: now.Add(time.Hour)}}
			m, _ := Escalar(tc.measure, 1000)
			normal, _ := Escalar(tc.normal, 100)
			a, err := Aplicar(p, tc.mode, 3, m, normal, 0, now)
			if tc.applications == 0 {
				if err == nil {
					t.Fatal("acepto medida insuficiente")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if a.Aplicaciones != tc.applications || a.SubtotalTotal != tc.total || a.MedidaCubierta != tc.covered || a.Sobrante != tc.remainder {
				t.Fatalf("desglose incorrecto: %+v", a)
			}
		})
	}
}
func TestElegibilidad(t *testing.T) {
	now := time.Now()
	base := models.Promocion{PromocionRequest: models.PromocionRequest{TipoPrendaID: 1, ModalidadCobro: models.ModalidadCobroUnidad, MedidaRequerida: 2, PrecioPromocional: 40, Activo: true, FechaInicio: now.Add(-time.Hour), FechaFin: now.Add(time.Hour)}}
	for _, name := range []string{"inactiva", "futura", "vencida", "otra_modalidad", "otro_tipo", "unidad_decimal", "precio_overflow"} {
		t.Run(name, func(t *testing.T) {
			p := base
			switch name {
			case "inactiva":
				p.Activo = false
			case "futura":
				p.FechaInicio = now.Add(time.Hour)
			case "vencida":
				p.FechaFin = now.Add(-time.Hour)
			case "otra_modalidad":
				p.ModalidadCobro = models.ModalidadCobroLibra
			case "otro_tipo":
				p.TipoPrendaID = 2
			case "unidad_decimal":
				p.MedidaRequerida = 1.5
			case "precio_overflow":
				p.PrecioPromocional = 99999999.99
			}
			_, err := Aplicar(p, models.ModalidadCobroUnidad, 1, 5000, 2500, 12500, now)
			if err == nil {
				t.Fatal("acepto promocion invalida")
			}
		})
	}
	// Extremos inclusivos de vigencia.
	base.FechaInicio = now
	base.FechaFin = now
	if _, err := Aplicar(base, models.ModalidadCobroUnidad, 1, 2000, 2500, 5000, now); err != nil {
		t.Fatal(err)
	}
}
func TestPrecisionYRedondeo(t *testing.T) {
	for _, tc := range []struct {
		v float64
		s int64
	}{{1.005, 100}, {0.0001, 1000}, {-1, 100}} {
		if _, err := Escalar(tc.v, tc.s); err == nil {
			t.Fatal(tc)
		}
	}
	got, err := Subtotal(101, 500)
	if err != nil || got != 51 {
		t.Fatal(got, err)
	}
	if _, err := Subtotal(MaxCentavos, 99999999999); err == nil {
		t.Fatal("overflow permitido")
	}
}
