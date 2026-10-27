package repository

import (
	"context"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// registrarCambioEstado es el único punto de escritura del historial. La hora
// y el usuario proceden del servidor y de la identidad autenticada; nunca del
// cuerpo enviado por el cliente.
func registrarCambioEstado(ctx context.Context, tx pgx.Tx, pedidoID, estadoID, usuarioID int, observaciones *string) (*models.PedidoEstadoHistorial, error) {
	const query = `
		INSERT INTO pedido_estados_historial (
			pedido_id, estado_id, usuario_id, fecha_cambio, observaciones
		)
		VALUES ($1, $2, $3, clock_timestamp(), $4)
		RETURNING id, pedido_id, estado_id, usuario_id,
		          fecha_cambio, observaciones, created_at`

	var historial models.PedidoEstadoHistorial
	if err := tx.QueryRow(ctx, query, pedidoID, estadoID, usuarioID, observaciones).Scan(
		&historial.ID, &historial.PedidoID, &historial.EstadoID,
		&historial.UsuarioID, &historial.FechaCambio,
		&historial.Observaciones, &historial.CreatedAt,
	); err != nil {
		if mappedErr := pedidoEstadoHistorialForeignKeyError(err); mappedErr != nil {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("error registrando historial del pedido: %w", err)
	}
	return &historial, nil
}
