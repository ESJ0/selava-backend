package repository

import (
	"math"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestReporteVentasAgregaPagosDelPeriodoInclusivo(t *testing.T) {
	f := newSprint3Fixture(t)
	pedido1 := f.order(t, 2)
	pedido2 := f.order(t, 1)

	var efectivoID, tarjetaID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO metodos_pago(nombre) VALUES('Efectivo reporte') RETURNING id`).Scan(&efectivoID))
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO metodos_pago(nombre) VALUES('Tarjeta reporte') RETURNING id`).Scan(&tarjetaID))

	type pagoFecha struct {
		pedidoID int
		metodoID int
		monto    float64
		fecha    time.Time
	}
	pagos := []pagoFecha{
		{pedidoID: pedido1.ID, metodoID: efectivoID, monto: 1, fecha: time.Date(2026, 8, 31, 23, 59, 59, 0, time.UTC)},
		{pedidoID: pedido1.ID, metodoID: efectivoID, monto: 5, fecha: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
		{pedidoID: pedido1.ID, metodoID: tarjetaID, monto: 6.25, fecha: time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)},
		{pedidoID: pedido2.ID, metodoID: efectivoID, monto: 10, fecha: time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC)},
		{pedidoID: pedido1.ID, metodoID: efectivoID, monto: 2, fecha: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
	}
	pagoRepo := NewPagoRepository(f.db)
	for _, item := range pagos {
		pago, err := pagoRepo.Create(f.ctx, item.pedidoID, &models.PagoCreateRequest{MetodoPagoID: item.metodoID, Monto: item.monto}, f.userID)
		mustSprint3(t, err)
		_, err = f.db.Exec(f.ctx, `UPDATE pagos SET fecha_pago=$1 WHERE id=$2`, item.fecha, pago.ID)
		mustSprint3(t, err)
	}

	reporte, err := NewReporteVentasRepository(f.db).GetByPeriodo(
		f.ctx,
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
	)
	mustSprint3(t, err)

	if reporte.CantidadVentas != 2 || reporte.CantidadPagos != 3 || math.Abs(reporte.TotalVentas-21.25) > 0.00001 {
		t.Fatalf("resumen inesperado: %+v", reporte)
	}
	if reporte.FechaInicio != "2026-09-01" || reporte.FechaFin != "2026-09-30" {
		t.Fatalf("periodo inesperado: %s - %s", reporte.FechaInicio, reporte.FechaFin)
	}
	if len(reporte.VentasPorDia) != 3 {
		t.Fatalf("ventas por dia = %+v", reporte.VentasPorDia)
	}
	if len(reporte.VentasPorMetodoPago) != 2 {
		t.Fatalf("ventas por metodo = %+v", reporte.VentasPorMetodoPago)
	}
	if reporte.VentasPorMetodoPago[0].MetodoPago.Nombre != "Efectivo reporte" || reporte.VentasPorMetodoPago[0].TotalVentas != 15 {
		t.Fatalf("agregado de efectivo inesperado: %+v", reporte.VentasPorMetodoPago[0])
	}
}
