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

var (
	ErrEstadoPedidoNoEncontrado        = errors.New("estado inicial de pedido no encontrado")
	ErrEstadoPedidoDestinoNoEncontrado = errors.New("estado de pedido no encontrado")
)

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
		}
		prendas = append(prendas, *prenda)
	}
	pedido.Total, pedido.UpdatedAt, err = recalcularTotalPedido(ctx, tx, pedido.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando transaccion: %w", err)
	}

	return &models.PedidoConPrendas{Pedido: *pedido, Prendas: prendas}, nil
}

func (r *PedidoRepository) UpdateEstado(ctx context.Context, pedidoID int, req *models.PedidoEstadoUpdateRequest, usuarioID int) (*models.PedidoEstadoHistorial, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando cambio de estado: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var pedidoBloqueadoID int
	if err := tx.QueryRow(ctx,
		`SELECT id FROM pedidos WHERE id = $1 AND activo = TRUE FOR UPDATE`,
		pedidoID,
	).Scan(&pedidoBloqueadoID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPedidoNoEncontrado
		}
		return nil, fmt.Errorf("error verificando pedido: %w", err)
	}

	var estadoBloqueadoID int
	if err := tx.QueryRow(ctx,
		`SELECT id FROM estados_pedido WHERE id = $1 FOR KEY SHARE`,
		req.EstadoID,
	).Scan(&estadoBloqueadoID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEstadoPedidoDestinoNoEncontrado
		}
		return nil, fmt.Errorf("error verificando estado de pedido: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE pedidos SET estado_actual_id = $1, updated_at = NOW() WHERE id = $2`,
		req.EstadoID, pedidoID,
	); err != nil {
		return nil, fmt.Errorf("error actualizando estado del pedido: %w", err)
	}

	const query = `
		INSERT INTO pedido_estados_historial (
			pedido_id, estado_id, usuario_id, observaciones
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, pedido_id, estado_id, usuario_id,
		          fecha_cambio, observaciones, created_at`
	var historial models.PedidoEstadoHistorial
	err = tx.QueryRow(ctx, query, pedidoID, req.EstadoID, usuarioID, req.Observaciones).Scan(
		&historial.ID, &historial.PedidoID, &historial.EstadoID,
		&historial.UsuarioID, &historial.FechaCambio,
		&historial.Observaciones, &historial.CreatedAt,
	)
	if err != nil {
		if mappedErr := pedidoEstadoHistorialForeignKeyError(err); mappedErr != nil {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("error registrando historial del pedido: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando cambio de estado: %w", err)
	}
	return &historial, nil
}

func (r *PedidoRepository) GetHistorialEstados(ctx context.Context, pedidoID int) ([]models.PedidoEstadoHistorial, error) {
	var pedidoExiste bool
	if err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pedidos WHERE id = $1 AND activo = TRUE)`,
		pedidoID,
	).Scan(&pedidoExiste); err != nil {
		return nil, fmt.Errorf("error verificando pedido para consultar historial: %w", err)
	}
	if !pedidoExiste {
		return nil, ErrPedidoNoEncontrado
	}

	const query = `
		SELECT id, pedido_id, estado_id, usuario_id,
		       fecha_cambio, observaciones, created_at
		FROM pedido_estados_historial
		WHERE pedido_id = $1
		ORDER BY fecha_cambio ASC, id ASC`
	rows, err := r.db.Query(ctx, query, pedidoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando historial del pedido: %w", err)
	}
	defer rows.Close()

	historial := make([]models.PedidoEstadoHistorial, 0)
	for rows.Next() {
		var movimiento models.PedidoEstadoHistorial
		if err := rows.Scan(
			&movimiento.ID, &movimiento.PedidoID, &movimiento.EstadoID,
			&movimiento.UsuarioID, &movimiento.FechaCambio,
			&movimiento.Observaciones, &movimiento.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error leyendo historial del pedido: %w", err)
		}
		historial = append(historial, movimiento)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo historial del pedido: %w", err)
	}

	return historial, nil
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
		RETURNING id, prenda_id, servicio_id, precio_aplicado, created_at, updated_at`
	var relacion models.PrendaServicio
	err := tx.QueryRow(ctx, query, prendaID, servicioID).Scan(
		&relacion.ID, &relacion.PrendaID, &relacion.ServicioID,
		&relacion.PrecioAplicado, &relacion.CreatedAt, &relacion.UpdatedAt,
	)
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

func pedidoEstadoHistorialForeignKeyError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "pedido_estados_historial_pedido_id_fkey":
		return ErrPedidoNoEncontrado
	case "pedido_estados_historial_estado_id_fkey":
		return ErrEstadoPedidoDestinoNoEncontrado
	case "pedido_estados_historial_usuario_id_fkey":
		return ErrUsuarioNoEncontrado
	default:
		return nil
	}
}
