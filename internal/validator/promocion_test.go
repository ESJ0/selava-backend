package validator

import (
	"github.com/ESJ0/selava-backend/internal/models"
	"testing"
	"time"
)

func TestValidacionPromociones(t *testing.T) {
	base := models.PromocionRequest{Nombre: "Dos por cuarenta", TipoPrendaID: 1, ModalidadCobro: models.ModalidadCobroUnidad, MedidaRequerida: 2, PrecioPromocional: 40, FechaInicio: time.Now(), FechaFin: time.Now().Add(time.Hour), Activo: true}
	if errs := ValidatePromocion(&base); errs.HasErrors() {
		t.Fatal(errs)
	}
	for _, name := range []string{"nombre", "tipo", "precio_cero", "precio_negativo", "precio_precision", "medida_cero", "medida_negativa", "unidad_decimal", "medida_precision", "fechas", "modalidad"} {
		t.Run(name, func(t *testing.T) {
			req := base
			switch name {
			case "nombre":
				req.Nombre = " "
			case "tipo":
				req.TipoPrendaID = 0
			case "precio_cero":
				req.PrecioPromocional = 0
			case "precio_negativo":
				req.PrecioPromocional = -1
			case "precio_precision":
				req.PrecioPromocional = 1.005
			case "medida_cero":
				req.MedidaRequerida = 0
			case "medida_negativa":
				req.MedidaRequerida = -1
			case "unidad_decimal":
				req.MedidaRequerida = 1.5
			case "medida_precision":
				req.MedidaRequerida = 0.0001
			case "fechas":
				req.FechaFin = req.FechaInicio.Add(-time.Hour)
			case "modalidad":
				req.ModalidadCobro = "OTRA"
			}
			if !ValidatePromocion(&req).HasErrors() {
				t.Fatal("acepto solicitud invalida")
			}
		})
	}
	for _, mode := range []models.ModalidadCobro{models.ModalidadCobroLibra, models.ModalidadCobroMetro} {
		req := base
		req.ModalidadCobro = mode
		req.MedidaRequerida = 0.125
		if errs := ValidatePromocion(&req); errs.HasErrors() {
			t.Fatal(errs)
		}
	}
}
