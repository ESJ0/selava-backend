package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReporteVentasRepository struct{ db *pgxpool.Pool }

func NewReporteVentasRepository(db *pgxpool.Pool) *ReporteVentasRepository {
	return &ReporteVentasRepository{db: db}
}

// GetByPeriodo obtiene una vista consistente de los pagos del periodo. El
// limite superior se recibe como fecha exclusiva para incluir completo el dia
// indicado por el usuario sin depender de la hora almacenada.
func (r *ReporteVentasRepository) GetByPeriodo(ctx context.Context, inicio, finExclusivo time.Time) (*models.ReporteVentas, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("error iniciando reporte de ventas: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	reporte := &models.ReporteVentas{
		FechaInicio:         inicio.Format(time.DateOnly),
		FechaFin:            finExclusivo.AddDate(0, 0, -1).Format(time.DateOnly),
		VentasPorDia:        make([]models.ReporteVentasPorDia, 0),
		VentasPorMetodoPago: make([]models.ReporteVentasPorMetodoPago, 0),
	}

	var cantidadVentas, cantidadPagos int64
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(DISTINCT pedido_id), COUNT(*), COALESCE(SUM(monto), 0)
		FROM pagos
		WHERE fecha_pago >= $1 AND fecha_pago < $2`, inicio, finExclusivo).Scan(
		&cantidadVentas, &cantidadPagos, &reporte.TotalVentas,
	); err != nil {
		return nil, fmt.Errorf("error consultando resumen de ventas: %w", err)
	}
	reporte.CantidadVentas = int(cantidadVentas)
	reporte.CantidadPagos = int(cantidadPagos)
	reporte.TotalVentas = redondearDosDecimales(reporte.TotalVentas)

	rows, err := tx.Query(ctx, `
		SELECT fecha_pago::date, COUNT(DISTINCT pedido_id), COUNT(*), SUM(monto)
		FROM pagos
		WHERE fecha_pago >= $1 AND fecha_pago < $2
		GROUP BY fecha_pago::date
		ORDER BY fecha_pago::date`, inicio, finExclusivo)
	if err != nil {
		return nil, fmt.Errorf("error consultando ventas por dia: %w", err)
	}
	for rows.Next() {
		var fecha time.Time
		var ventas, pagos int64
		var item models.ReporteVentasPorDia
		if err := rows.Scan(&fecha, &ventas, &pagos, &item.TotalVentas); err != nil {
			rows.Close()
			return nil, fmt.Errorf("error leyendo ventas por dia: %w", err)
		}
		item.Fecha = fecha.Format(time.DateOnly)
		item.CantidadVentas = int(ventas)
		item.CantidadPagos = int(pagos)
		item.TotalVentas = redondearDosDecimales(item.TotalVentas)
		reporte.VentasPorDia = append(reporte.VentasPorDia, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("error recorriendo ventas por dia: %w", err)
	}
	rows.Close()

	rows, err = tx.Query(ctx, `
		SELECT mp.id, mp.nombre, mp.activo,
		       COUNT(DISTINCT pg.pedido_id), COUNT(*), SUM(pg.monto)
		FROM pagos pg
		JOIN metodos_pago mp ON mp.id = pg.metodo_pago_id
		WHERE pg.fecha_pago >= $1 AND pg.fecha_pago < $2
		GROUP BY mp.id, mp.nombre, mp.activo
		ORDER BY mp.nombre, mp.id`, inicio, finExclusivo)
	if err != nil {
		return nil, fmt.Errorf("error consultando ventas por metodo de pago: %w", err)
	}
	for rows.Next() {
		var ventas, pagos int64
		var item models.ReporteVentasPorMetodoPago
		if err := rows.Scan(
			&item.MetodoPago.ID, &item.MetodoPago.Nombre, &item.MetodoPago.Activo,
			&ventas, &pagos, &item.TotalVentas,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("error leyendo ventas por metodo de pago: %w", err)
		}
		item.CantidadVentas = int(ventas)
		item.CantidadPagos = int(pagos)
		item.TotalVentas = redondearDosDecimales(item.TotalVentas)
		reporte.VentasPorMetodoPago = append(reporte.VentasPorMetodoPago, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("error recorriendo ventas por metodo de pago: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando reporte de ventas: %w", err)
	}
	return reporte, nil
}

func redondearDosDecimales(value float64) float64 {
	return math.Round(value*100) / 100
}
