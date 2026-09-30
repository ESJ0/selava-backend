package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReporteConsumoInsumosRepository struct{ db *pgxpool.Pool }

func NewReporteConsumoInsumosRepository(db *pgxpool.Pool) *ReporteConsumoInsumosRepository {
	return &ReporteConsumoInsumosRepository{db: db}
}

func (r *ReporteConsumoInsumosRepository) GetByPeriodo(ctx context.Context, inicio, finExclusivo time.Time) (*models.ReporteConsumoInsumos, error) {
	rows, err := r.db.Query(ctx, `
		SELECT i.id, i.nombre, i.unidad_medida, COUNT(m.id), SUM(m.cantidad)
		FROM movimientos_inventario m
		JOIN insumos i ON i.id = m.insumo_id
		WHERE m.tipo_movimiento = 'salida'
		  AND m.fecha_movimiento >= $1
		  AND m.fecha_movimiento < $2
		GROUP BY i.id, i.nombre, i.unidad_medida
		ORDER BY i.nombre, i.id`, inicio, finExclusivo)
	if err != nil {
		return nil, fmt.Errorf("error consultando consumo de insumos: %w", err)
	}
	defer rows.Close()

	reporte := &models.ReporteConsumoInsumos{
		FechaInicio:      inicio.Format(time.DateOnly),
		FechaFin:         finExclusivo.AddDate(0, 0, -1).Format(time.DateOnly),
		ConsumoPorInsumo: make([]models.ReporteConsumoInsumoDetalle, 0),
	}
	for rows.Next() {
		var movimientos int64
		var item models.ReporteConsumoInsumoDetalle
		if err := rows.Scan(
			&item.InsumoID, &item.Nombre, &item.UnidadMedida,
			&movimientos, &item.CantidadConsumida,
		); err != nil {
			return nil, fmt.Errorf("error leyendo consumo de insumos: %w", err)
		}
		item.CantidadMovimientos = int(movimientos)
		item.CantidadConsumida = redondearDosDecimales(item.CantidadConsumida)
		reporte.CantidadMovimientos += item.CantidadMovimientos
		reporte.TotalConsumido += item.CantidadConsumida
		reporte.ConsumoPorInsumo = append(reporte.ConsumoPorInsumo, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo consumo de insumos: %w", err)
	}
	reporte.TotalConsumido = redondearDosDecimales(reporte.TotalConsumido)
	return reporte, nil
}
