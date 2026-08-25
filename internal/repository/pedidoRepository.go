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

// Create inserta el pedido y todas sus prendas dentro de una misma
// transaccion: si alguna prenda falla (p. ej. un tipo_prenda_id que no
// existe), se revierte tambien el pedido para no dejar registros huerfanos.
func (r *PedidoRepository) Create(ctx context.Context, req *models.PedidoCreateRequest, usuarioID int) (*models.PedidoConPrendas, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando transaccion: %w", err)
	}
	defer tx.Rollback(ctx)

	pedido, err := r.insertPedido(ctx, tx, req, usuarioID)
	if err != nil {
		return nil, err
	}

	prendas := make([]models.Prenda, 0, len(req.Prendas))
	var total float64
	for _, prendaReq := range req.Prendas {
		prenda, err := r.insertPrenda(ctx, tx, pedido.ID, prendaReq)
		if err != nil {
			return nil, err
		}
		for _, servicioReq := range prendaReq.Servicios {
			relacion, err := r.insertPrendaServicio(ctx, tx, prenda.ID, servicioReq.ServicioID)
			if err != nil {
				return nil, err
			}
			prenda.Servicios = append(prenda.Servicios, *relacion)
			total += float64(prenda.Cantidad) * relacion.PrecioAplicado
		}
		prendas = append(prendas, *prenda)
	}
	if err := tx.QueryRow(ctx, `UPDATE pedidos SET total=$1, updated_at=NOW() WHERE id=$2 RETURNING total, updated_at`, total, pedido.ID).Scan(&pedido.Total, &pedido.UpdatedAt); err != nil {
		return nil, fmt.Errorf("error actualizando total del pedido: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando transaccion: %w", err)
	}

	return &models.PedidoConPrendas{Pedido: *pedido, Prendas: prendas}, nil
}

func (r *PedidoRepository) insertPedido(ctx context.Context, tx pgx.Tx, req *models.PedidoCreateRequest, usuarioID int) (*models.Pedido, error) {
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
	err := tx.QueryRow(ctx, query,
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

func (r *PedidoRepository) insertPrenda(ctx context.Context, tx pgx.Tx, pedidoID int, req models.PrendaCreateRequest) (*models.Prenda, error) {
	const query = `
		INSERT INTO prendas (pedido_id, tipo_prenda_id, descripcion, cantidad, color)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, pedido_id, tipo_prenda_id, descripcion, cantidad, color, created_at, updated_at`

	var prenda models.Prenda
	err := tx.QueryRow(ctx, query,
		pedidoID, req.TipoPrendaID, req.Descripcion, req.Cantidad, req.Color,
	).Scan(
		&prenda.ID, &prenda.PedidoID, &prenda.TipoPrendaID, &prenda.Descripcion,
		&prenda.Cantidad, &prenda.Color, &prenda.CreatedAt, &prenda.UpdatedAt,
	)
	if err != nil {
		if mappedErr := pedidoPrendaForeignKeyError(err); mappedErr != nil {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("error creando prenda: %w", err)
	}
	return &prenda, nil
}

func (r *PedidoRepository) insertPrendaServicio(ctx context.Context, tx pgx.Tx, prendaID, servicioID int) (*models.PrendaServicio, error) {
	const query = `INSERT INTO prenda_servicios(prenda_id, servicio_id, precio_aplicado)
		SELECT $1, id, precio_base FROM servicios WHERE id=$2 AND activo=TRUE
		RETURNING id, prenda_id, servicio_id, precio_aplicado`
	var relacion models.PrendaServicio
	err := tx.QueryRow(ctx, query, prendaID, servicioID).Scan(&relacion.ID, &relacion.PrendaID, &relacion.ServicioID, &relacion.PrecioAplicado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrServicioNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error asociando servicio a prenda: %w", err)
	}
	return &relacion, nil
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

func pedidoPrendaForeignKeyError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "prendas_tipo_prenda_id_fkey":
		return ErrTipoPrendaNoEncontrado
	case "prendas_pedido_id_fkey":
		return ErrEstadoPedidoNoEncontrado
	default:
		return nil
	}
}
