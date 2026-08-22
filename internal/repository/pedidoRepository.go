package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEstadoPedidoNoEncontrado = errors.New("estado inicial de pedido no encontrado")

type PedidoRepository struct {
	db *pgxpool.Pool
}

func NewPedidoRepository(db *pgxpool.Pool) *PedidoRepository {
	return &PedidoRepository{db: db}
}

func (r *PedidoRepository) Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.Pedido, error) {
	const query = `
		INSERT INTO pedidos (
			cliente_id, usuario_id, estado_actual_id,
			fecha_entrega_estimada, observaciones
		)
		SELECT $1, $2, id, $3, $4
		FROM estados_pedido
		WHERE nombre = 'Recibido'
		RETURNING id, cliente_id, usuario_id, estado_actual_id,
		          fecha_recibido, fecha_entrega_estimada, fecha_entrega_real,
		          total, observaciones, activo, created_at, updated_at`

	var pedido models.Pedido
	err := r.db.QueryRow(ctx, query,
		req.ClienteID, usuarioID, req.FechaEntregaEstimada, req.Observaciones,
	).Scan(
		&pedido.ID, &pedido.ClienteID, &pedido.UsuarioID, &pedido.EstadoActualID,
		&pedido.FechaRecibido, &pedido.FechaEntregaEstimada, &pedido.FechaEntregaReal,
		&pedido.Total, &pedido.Observaciones, &pedido.Activo,
		&pedido.CreatedAt, &pedido.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEstadoPedidoNoEncontrado
	}
	if err != nil {
		if mappedErr := pedidoForeignKeyError(err); mappedErr != nil {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("error creando pedido: %w", err)
	}
	return &pedido, nil
}

func pedidoForeignKeyError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "pedidos_cliente_id_fkey":
		return ErrClienteNoEncontrado
	case "pedidos_usuario_id_fkey":
		return ErrUsuarioNoEncontrado
	case "pedidos_estado_actual_id_fkey":
		return ErrEstadoPedidoNoEncontrado
	default:
		return nil
	}
}
