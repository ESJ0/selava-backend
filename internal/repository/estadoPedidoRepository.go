package repository

import (
	"context"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EstadoPedidoRepository struct {
	db *pgxpool.Pool
}

func NewEstadoPedidoRepository(db *pgxpool.Pool) *EstadoPedidoRepository {
	return &EstadoPedidoRepository{db: db}
}

// List devuelve el catalogo completo de estados de pedido, ordenado por
// su orden de flujo, con las banderas que definen el comportamiento operativo.
func (r *EstadoPedidoRepository) List(ctx context.Context) ([]models.EstadoPedido, error) {
	const query = `
		SELECT id, nombre, codigo, orden, es_obligatorio, es_terminal,
		       es_cancelacion, created_at, updated_at
		FROM estados_pedido
		ORDER BY orden`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error consultando catalogo de estados de pedido: %w", err)
	}
	defer rows.Close()

	estados := make([]models.EstadoPedido, 0)
	for rows.Next() {
		var estado models.EstadoPedido
		if err := rows.Scan(
			&estado.ID, &estado.Nombre, &estado.Codigo, &estado.Orden,
			&estado.EsObligatorio, &estado.EsTerminal, &estado.EsCancelacion,
			&estado.CreatedAt, &estado.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error leyendo catalogo de estados de pedido: %w", err)
		}
		estados = append(estados, estado)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo catalogo de estados de pedido: %w", err)
	}
	return estados, nil
}
