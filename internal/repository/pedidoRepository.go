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
	ErrEstadoPedidoNoEncontrado          = errors.New("estado inicial de pedido no encontrado")
	ErrEstadoPedidoDestinoNoEncontrado   = errors.New("estado de pedido no encontrado")
	ErrEstadoPedidoCanceladoNoEncontrado = errors.New("estado cancelado de pedido no encontrado")
	ErrPedidoNoCancelable                = errors.New("el pedido no se puede cancelar en su estado actual")
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

func (r *PedidoRepository) GetDetalle(ctx context.Context, pedidoID int) (*models.PedidoDetalle, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando consulta de detalle del pedido: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	detalle := &models.PedidoDetalle{
		Prendas: make([]models.Prenda, 0),
		Pagos:   make([]models.PagoDetalle, 0),
	}
	const pedidoQuery = `
		SELECT p.id, p.cliente_id, p.usuario_id, p.estado_actual_id,
		       p.fecha_recibido, p.fecha_entrega_estimada, p.fecha_entrega_real,
		       p.total, p.observaciones, p.activo, p.created_at, p.updated_at,
		       c.id, c.nombre, c.apellido, c.telefono,
		       COALESCE(c.email, ''), COALESCE(c.direccion, ''),
		       c.activo, c.created_at, c.updated_at,
		       e.id, e.nombre, e.orden, e.created_at, e.updated_at
		FROM pedidos p
		JOIN clientes c ON c.id = p.cliente_id
		JOIN estados_pedido e ON e.id = p.estado_actual_id
		WHERE p.id = $1 AND p.activo = TRUE`
	if err := tx.QueryRow(ctx, pedidoQuery, pedidoID).Scan(
		&detalle.ID, &detalle.ClienteID, &detalle.UsuarioID, &detalle.EstadoActualID,
		&detalle.FechaRecibido, &detalle.FechaEntregaEstimada, &detalle.FechaEntregaReal,
		&detalle.Total, &detalle.Observaciones, &detalle.Activo,
		&detalle.CreatedAt, &detalle.UpdatedAt,
		&detalle.Cliente.ID, &detalle.Cliente.Nombre, &detalle.Cliente.Apellido,
		&detalle.Cliente.Telefono, &detalle.Cliente.Email, &detalle.Cliente.Direccion,
		&detalle.Cliente.Activo, &detalle.Cliente.CreatedAt, &detalle.Cliente.UpdatedAt,
		&detalle.EstadoActual.ID, &detalle.EstadoActual.Nombre, &detalle.EstadoActual.Orden,
		&detalle.EstadoActual.CreatedAt, &detalle.EstadoActual.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPedidoNoEncontrado
		}
		return nil, fmt.Errorf("error obteniendo detalle del pedido: %w", err)
	}

	const prendasQuery = `
		SELECT p.id, p.pedido_id, p.tipo_prenda_id, p.descripcion,
		       p.cantidad, p.color, p.created_at, p.updated_at,
		       tp.id, tp.nombre, tp.descripcion, tp.activo,
		       tp.created_at, tp.updated_at
		FROM prendas p
		JOIN tipos_prenda tp ON tp.id = p.tipo_prenda_id
		WHERE p.pedido_id = $1
		ORDER BY p.id`
	rows, err := tx.Query(ctx, prendasQuery, pedidoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando prendas del pedido: %w", err)
	}
	for rows.Next() {
		prenda := models.Prenda{Servicios: make([]models.PrendaServicio, 0)}
		tipoPrenda := &models.TipoPrenda{}
		if err := rows.Scan(
			&prenda.ID, &prenda.PedidoID, &prenda.TipoPrendaID, &prenda.Descripcion,
			&prenda.Cantidad, &prenda.Color, &prenda.CreatedAt, &prenda.UpdatedAt,
			&tipoPrenda.ID, &tipoPrenda.Nombre, &tipoPrenda.Descripcion,
			&tipoPrenda.Activo, &tipoPrenda.CreatedAt, &tipoPrenda.UpdatedAt,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("error leyendo prendas del pedido: %w", err)
		}
		prenda.TipoPrenda = tipoPrenda
		detalle.Prendas = append(detalle.Prendas, prenda)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("error recorriendo prendas del pedido: %w", err)
	}
	rows.Close()
	for i := range detalle.Prendas {
		if err := r.loadServiciosDePrenda(ctx, tx, &detalle.Prendas[i]); err != nil {
			return nil, err
		}
	}

	const pagosQuery = `
		SELECT p.id, p.pedido_id, p.metodo_pago_id, p.usuario_id,
		       p.monto, p.referencia, p.fecha_pago, p.created_at,
		       mp.id, mp.nombre, mp.activo
		FROM pagos p
		JOIN metodos_pago mp ON mp.id = p.metodo_pago_id
		WHERE p.pedido_id = $1
		ORDER BY p.fecha_pago ASC, p.id ASC`
	rows, err = tx.Query(ctx, pagosQuery, pedidoID)
	if err != nil {
		return nil, fmt.Errorf("error consultando pagos del pedido: %w", err)
	}
	for rows.Next() {
		pago := models.PagoDetalle{}
		if err := rows.Scan(
			&pago.ID, &pago.PedidoID, &pago.MetodoPagoID, &pago.UsuarioID,
			&pago.Monto, &pago.Referencia, &pago.FechaPago, &pago.CreatedAt,
			&pago.MetodoPago.ID, &pago.MetodoPago.Nombre, &pago.MetodoPago.Activo,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("error leyendo pagos del pedido: %w", err)
		}
		detalle.Pagos = append(detalle.Pagos, pago)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("error recorriendo pagos del pedido: %w", err)
	}
	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando consulta de detalle del pedido: %w", err)
	}
	return detalle, nil
}

func (r *PedidoRepository) Cancelar(ctx context.Context, pedidoID, usuarioID int) (*models.Pedido, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando cancelacion del pedido: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var estadoActualNombre string
	if err := tx.QueryRow(ctx, `
		SELECT e.nombre
		FROM pedidos p
		JOIN estados_pedido e ON e.id = p.estado_actual_id
		WHERE p.id = $1 AND p.activo = TRUE
		FOR UPDATE OF p`, pedidoID).Scan(&estadoActualNombre); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPedidoNoEncontrado
		}
		return nil, fmt.Errorf("error verificando estado actual del pedido: %w", err)
	}

	if estadoActualNombre != "Recibido" {
		return nil, ErrPedidoNoCancelable
	}

	var estadoCanceladoID int
	if err := tx.QueryRow(ctx,
		`SELECT id FROM estados_pedido WHERE nombre = 'Cancelado' FOR KEY SHARE`,
	).Scan(&estadoCanceladoID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEstadoPedidoCanceladoNoEncontrado
		}
		return nil, fmt.Errorf("error verificando estado cancelado del pedido: %w", err)
	}

	var pedido models.Pedido
	if err := tx.QueryRow(ctx, `
		UPDATE pedidos
		SET estado_actual_id = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, cliente_id, usuario_id, estado_actual_id,
		          fecha_recibido, fecha_entrega_estimada, fecha_entrega_real,
		          total, observaciones, activo, created_at, updated_at`,
		estadoCanceladoID, pedidoID).Scan(
		&pedido.ID, &pedido.ClienteID, &pedido.UsuarioID, &pedido.EstadoActualID,
		&pedido.FechaRecibido, &pedido.FechaEntregaEstimada, &pedido.FechaEntregaReal,
		&pedido.Total, &pedido.Observaciones, &pedido.Activo,
		&pedido.CreatedAt, &pedido.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("error actualizando pedido cancelado: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO pedido_estados_historial (pedido_id, estado_id, usuario_id, observaciones)
		VALUES ($1, $2, $3, $4)`,
		pedidoID, estadoCanceladoID, usuarioID, "Pedido cancelado"); err != nil {
		if mappedErr := pedidoEstadoHistorialForeignKeyError(err); mappedErr != nil {
			return nil, mappedErr
		}
		return nil, fmt.Errorf("error registrando cancelacion en el historial: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando cancelacion del pedido: %w", err)
	}
	return &pedido, nil
}

func (r *PedidoRepository) loadServiciosDePrenda(ctx context.Context, tx pgx.Tx, prenda *models.Prenda) error {
	const query = `
		SELECT ps.id, ps.prenda_id, ps.servicio_id, ps.precio_aplicado,
		       ps.created_at, ps.updated_at,
		       s.id, s.nombre, s.descripcion, s.precio_base,
		       s.tiempo_estimado_horas, s.activo, s.created_at, s.updated_at
		FROM prenda_servicios ps
		JOIN servicios s ON s.id = ps.servicio_id
		WHERE ps.prenda_id = $1
		ORDER BY ps.id`
	rows, err := tx.Query(ctx, query, prenda.ID)
	if err != nil {
		return fmt.Errorf("error consultando servicios de la prenda: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		servicioAplicado := models.PrendaServicio{Servicio: &models.Servicio{}}
		if err := rows.Scan(
			&servicioAplicado.ID, &servicioAplicado.PrendaID, &servicioAplicado.ServicioID,
			&servicioAplicado.PrecioAplicado, &servicioAplicado.CreatedAt,
			&servicioAplicado.UpdatedAt,
			&servicioAplicado.Servicio.ID, &servicioAplicado.Servicio.Nombre,
			&servicioAplicado.Servicio.Descripcion, &servicioAplicado.Servicio.PrecioBase,
			&servicioAplicado.Servicio.TiempoEstimadoHoras, &servicioAplicado.Servicio.Activo,
			&servicioAplicado.Servicio.CreatedAt, &servicioAplicado.Servicio.UpdatedAt,
		); err != nil {
			return fmt.Errorf("error leyendo servicios de la prenda: %w", err)
		}
		prenda.Servicios = append(prenda.Servicios, servicioAplicado)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error recorriendo servicios de la prenda: %w", err)
	}
	return nil
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
