package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportePedidosEstadoRepository struct{ db *pgxpool.Pool }

func NewReportePedidosEstadoRepository(db *pgxpool.Pool) *ReportePedidosEstadoRepository {
	return &ReportePedidosEstadoRepository{db: db}
}

func (r *ReportePedidosEstadoRepository) GetByPeriodo(ctx context.Context, inicio, finExclusivo time.Time) (*models.ReportePedidosEstado, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.id, e.nombre, e.codigo, e.orden,
		       e.es_obligatorio, e.es_terminal, e.es_cancelacion,
		       e.created_at, e.updated_at, COUNT(p.id)
		FROM estados_pedido e
		LEFT JOIN pedidos p
		       ON p.estado_actual_id = e.id
		      AND p.activo = TRUE
		      AND p.fecha_recibido >= $1
		      AND p.fecha_recibido < $2
		GROUP BY e.id, e.nombre, e.codigo, e.orden, e.es_obligatorio,
		         e.es_terminal, e.es_cancelacion, e.created_at, e.updated_at
		ORDER BY e.orden, e.id`, inicio, finExclusivo)
	if err != nil {
		return nil, fmt.Errorf("error consultando pedidos por estado: %w", err)
	}
	defer rows.Close()

	reporte := &models.ReportePedidosEstado{
		FechaInicio:      inicio.Format(time.DateOnly),
		FechaFin:         finExclusivo.AddDate(0, 0, -1).Format(time.DateOnly),
		PedidosPorEstado: make([]models.ReportePedidosEstadoDetalle, 0),
	}
	for rows.Next() {
		var cantidad int64
		var item models.ReportePedidosEstadoDetalle
		if err := rows.Scan(
			&item.Estado.ID, &item.Estado.Nombre, &item.Estado.Codigo,
			&item.Estado.Orden, &item.Estado.EsObligatorio,
			&item.Estado.EsTerminal, &item.Estado.EsCancelacion,
			&item.Estado.CreatedAt, &item.Estado.UpdatedAt, &cantidad,
		); err != nil {
			return nil, fmt.Errorf("error leyendo pedidos por estado: %w", err)
		}
		item.CantidadPedidos = int(cantidad)
		reporte.TotalPedidos += item.CantidadPedidos
		reporte.PedidosPorEstado = append(reporte.PedidosPorEstado, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo pedidos por estado: %w", err)
	}
	return reporte, nil
}
