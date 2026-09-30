package repository

import (
	"math"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
)

func TestReporteConsumoInsumosSumaSoloSalidasDelPeriodo(t *testing.T) {
	f := newSprint3Fixture(t)
	var detergenteID, cloroID int
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO insumos(nombre, unidad_medida, stock_actual) VALUES('Detergente reporte','L',100) RETURNING id`).Scan(&detergenteID))
	mustSprint3(t, f.db.QueryRow(f.ctx, `INSERT INTO insumos(nombre, unidad_medida, stock_actual) VALUES('Cloro reporte','L',100) RETURNING id`).Scan(&cloroID))

	type movimientoFecha struct {
		insumoID int
		tipo     string
		cantidad float64
		fecha    time.Time
	}
	movimientos := []movimientoFecha{
		{insumoID: detergenteID, tipo: "salida", cantidad: 1, fecha: time.Date(2026, 8, 31, 23, 59, 59, 0, time.UTC)},
		{insumoID: detergenteID, tipo: "salida", cantidad: 4, fecha: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)},
		{insumoID: detergenteID, tipo: "salida", cantidad: 2.25, fecha: time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)},
		{insumoID: cloroID, tipo: "salida", cantidad: 2.5, fecha: time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC)},
		{insumoID: cloroID, tipo: "entrada", cantidad: 10, fecha: time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)},
		{insumoID: detergenteID, tipo: "salida", cantidad: 3, fecha: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
	}
	movimientoRepo := NewMovimientoInventarioRepository(f.db)
	for _, item := range movimientos {
		movimiento, err := movimientoRepo.Create(f.ctx, &models.MovimientoInventarioCreateRequest{
			InsumoID: item.insumoID, TipoMovimiento: item.tipo, Cantidad: item.cantidad,
		}, f.userID)
		mustSprint3(t, err)
		_, err = f.db.Exec(f.ctx, `UPDATE movimientos_inventario SET fecha_movimiento=$1 WHERE id=$2`, item.fecha, movimiento.ID)
		mustSprint3(t, err)
	}

	reporte, err := NewReporteConsumoInsumosRepository(f.db).GetByPeriodo(
		f.ctx,
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
	)
	mustSprint3(t, err)
	if reporte.CantidadMovimientos != 3 || math.Abs(reporte.TotalConsumido-8.75) > 0.00001 {
		t.Fatalf("resumen inesperado: %+v", reporte)
	}
	if len(reporte.ConsumoPorInsumo) != 2 {
		t.Fatalf("consumo por insumo inesperado: %+v", reporte.ConsumoPorInsumo)
	}
	consumos := make(map[string]models.ReporteConsumoInsumoDetalle)
	for _, item := range reporte.ConsumoPorInsumo {
		consumos[item.Nombre] = item
	}
	if consumos["Detergente reporte"].CantidadConsumida != 6.25 || consumos["Detergente reporte"].CantidadMovimientos != 2 {
		t.Fatalf("consumo de detergente inesperado: %+v", consumos["Detergente reporte"])
	}
	if consumos["Cloro reporte"].CantidadConsumida != 2.5 {
		t.Fatalf("consumo de cloro inesperado: %+v", consumos["Cloro reporte"])
	}
}
