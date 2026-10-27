package repository

import (
	"context"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// cargarFlujoOperativo obtiene únicamente las etapas obligatorias y aquellas
// habilitadas por al menos un servicio contratado en las prendas del pedido.
func cargarFlujoOperativo(ctx context.Context, tx pgx.Tx, pedidoID int, actual models.EstadoPedido) (models.FlujoOperativoPedido, error) {
	const query = `
		WITH servicios_contratados AS (
			SELECT DISTINCT ps.servicio_id
			FROM prendas p
			JOIN prenda_servicios ps ON ps.prenda_id = p.id
			WHERE p.pedido_id = $1
		)
		SELECT e.id, e.nombre, e.codigo, e.orden,
		       e.es_obligatorio, e.es_terminal, e.es_cancelacion,
		       e.created_at, e.updated_at,
		       s.id, s.nombre
		FROM estados_pedido e
		LEFT JOIN estado_pedido_servicios eps ON eps.estado_pedido_id = e.id
		LEFT JOIN servicios_contratados sc ON sc.servicio_id = eps.servicio_id
		LEFT JOIN servicios s ON s.id = sc.servicio_id
		WHERE NOT e.es_cancelacion
		  AND (e.es_obligatorio OR sc.servicio_id IS NOT NULL)
		ORDER BY e.orden, e.id, s.nombre, s.id`

	rows, err := tx.Query(ctx, query, pedidoID)
	if err != nil {
		return models.FlujoOperativoPedido{}, fmt.Errorf("error consultando flujo operativo del pedido: %w", err)
	}
	defer rows.Close()

	flujo := models.FlujoOperativoPedido{Estados: make([]models.EstadoFlujoPedido, 0)}
	porID := make(map[int]int)
	for rows.Next() {
		var estado models.EstadoPedido
		var servicioID *int
		var servicioNombre *string
		if err := rows.Scan(
			&estado.ID, &estado.Nombre, &estado.Codigo, &estado.Orden,
			&estado.EsObligatorio, &estado.EsTerminal, &estado.EsCancelacion,
			&estado.CreatedAt, &estado.UpdatedAt, &servicioID, &servicioNombre,
		); err != nil {
			return models.FlujoOperativoPedido{}, fmt.Errorf("error leyendo flujo operativo del pedido: %w", err)
		}

		index, existe := porID[estado.ID]
		if !existe {
			index = len(flujo.Estados)
			porID[estado.ID] = index
			flujo.Estados = append(flujo.Estados, models.EstadoFlujoPedido{
				EstadoPedido: estado,
				Servicios:    make([]models.ServicioEtapa, 0),
			})
		}
		if servicioID != nil && servicioNombre != nil {
			flujo.Estados[index].Servicios = append(flujo.Estados[index].Servicios, models.ServicioEtapa{
				ID: *servicioID, Nombre: *servicioNombre,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return models.FlujoOperativoPedido{}, fmt.Errorf("error recorriendo flujo operativo del pedido: %w", err)
	}

	flujo.Cancelado = actual.EsCancelacion
	actualIndex := -1
	for i := range flujo.Estados {
		flujo.Estados[i].Actual = flujo.Estados[i].ID == actual.ID && !flujo.Cancelado
		flujo.Estados[i].Completado = !flujo.Cancelado && flujo.Estados[i].Orden <= actual.Orden
		if flujo.Estados[i].Actual {
			actualIndex = i
		}
	}

	if !flujo.Cancelado && len(flujo.Estados) > 0 {
		completadas := 0
		for i := range flujo.Estados {
			if flujo.Estados[i].Completado {
				completadas++
			}
		}
		flujo.ProgresoPorcentaje = completadas * 100 / len(flujo.Estados)
		if actualIndex >= 0 && actualIndex+1 < len(flujo.Estados) {
			siguiente := flujo.Estados[actualIndex+1].EstadoPedido
			flujo.SiguienteEstado = &siguiente
		}
	}

	return flujo, nil
}
