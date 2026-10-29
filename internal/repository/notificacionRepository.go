package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotificacionNoEncontrada = errors.New("notificacion no encontrada")

type NotificacionRepository struct {
	db *pgxpool.Pool
}

func NewNotificacionRepository(db *pgxpool.Pool) *NotificacionRepository {
	return &NotificacionRepository{db: db}
}

func (r *NotificacionRepository) ListByCliente(ctx context.Context, clienteID int) (*models.NotificacionesResponse, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, cliente_id, pedido_id, mensaje, leida, created_at,
		       COUNT(*) FILTER (WHERE leida = FALSE) OVER ()
		FROM notificaciones
		WHERE cliente_id = $1
		ORDER BY created_at DESC, id DESC`, clienteID)
	if err != nil {
		return nil, fmt.Errorf("error consultando notificaciones: %w", err)
	}
	defer rows.Close()

	response := &models.NotificacionesResponse{Notificaciones: []models.Notificacion{}}
	for rows.Next() {
		var notificacion models.Notificacion
		if err := rows.Scan(
			&notificacion.ID, &notificacion.ClienteID, &notificacion.PedidoID,
			&notificacion.Mensaje, &notificacion.Leida, &notificacion.CreatedAt,
			&response.NoLeidas,
		); err != nil {
			return nil, fmt.Errorf("error leyendo notificacion: %w", err)
		}
		response.Notificaciones = append(response.Notificaciones, notificacion)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo notificaciones: %w", err)
	}
	return response, nil
}

func (r *NotificacionRepository) Create(ctx context.Context, clienteID int, pedidoID *int, mensaje string) (*models.Notificacion, error) {
	var notificacion models.Notificacion
	err := r.db.QueryRow(ctx, `
		INSERT INTO notificaciones(cliente_id, pedido_id, mensaje)
		VALUES($1, $2, $3)
		RETURNING id, cliente_id, pedido_id, mensaje, leida, created_at`,
		clienteID, pedidoID, mensaje,
	).Scan(
		&notificacion.ID, &notificacion.ClienteID, &notificacion.PedidoID,
		&notificacion.Mensaje, &notificacion.Leida, &notificacion.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creando notificacion: %w", err)
	}
	return &notificacion, nil
}

func (r *NotificacionRepository) MarkReadByCliente(ctx context.Context, notificacionID, clienteID int) (*models.Notificacion, error) {
	var notificacion models.Notificacion
	err := r.db.QueryRow(ctx, `
		UPDATE notificaciones SET leida = TRUE
		WHERE id = $1 AND cliente_id = $2
		RETURNING id, cliente_id, pedido_id, mensaje, leida, created_at`,
		notificacionID, clienteID,
	).Scan(
		&notificacion.ID, &notificacion.ClienteID, &notificacion.PedidoID,
		&notificacion.Mensaje, &notificacion.Leida, &notificacion.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotificacionNoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("error marcando notificacion como leida: %w", err)
	}
	return &notificacion, nil
}

func (r *NotificacionRepository) MarkAllReadByCliente(ctx context.Context, clienteID int) (int64, error) {
	result, err := r.db.Exec(ctx, `
		UPDATE notificaciones SET leida = TRUE
		WHERE cliente_id = $1 AND leida = FALSE`, clienteID)
	if err != nil {
		return 0, fmt.Errorf("error marcando notificaciones como leidas: %w", err)
	}
	return result.RowsAffected(), nil
}

func (r *NotificacionRepository) DeleteByCliente(ctx context.Context, notificacionID, clienteID int) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM notificaciones WHERE id = $1 AND cliente_id = $2`,
		notificacionID, clienteID,
	)
	if err != nil {
		return fmt.Errorf("error eliminando notificacion: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotificacionNoEncontrada
	}
	return nil
}

// insertarNotificacionCambioEstado participa en la misma transaccion que el
// pedido: el cliente nunca recibe un aviso de un cambio que no se confirmo.
func insertarNotificacionCambioEstado(ctx context.Context, tx pgx.Tx, pedidoID, clienteID int, estado models.EstadoPedido) error {
	mensaje := fmt.Sprintf("Tu pedido #%d cambió de estado a %s.", pedidoID, estado.Nombre)
	if estado.Codigo == "FINALIZADO" || estado.Nombre == "Finalizado" {
		mensaje = fmt.Sprintf("Tu pedido #%d ha sido finalizado. Ya puedes consultar su detalle.", pedidoID)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notificaciones(cliente_id, pedido_id, mensaje)
		VALUES($1, $2, $3)`, clienteID, pedidoID, mensaje); err != nil {
		return fmt.Errorf("error creando notificacion de cambio de estado: %w", err)
	}
	return nil
}
