package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPagoExcedeSaldo     = errors.New("el monto del pago excede el saldo pendiente")
	ErrPedidoSinSaldo      = errors.New("el pedido no tiene saldo pendiente")
	ErrPagoPedidoCancelado = errors.New("no se pueden registrar pagos para un pedido cancelado")
)

type PagoRepository struct{ db *pgxpool.Pool }

func NewPagoRepository(db *pgxpool.Pool) *PagoRepository { return &PagoRepository{db: db} }

func (r *PagoRepository) Create(ctx context.Context, pedidoID int, req *models.PagoCreateRequest, usuarioID int) (*models.PagoDetalle, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando registro de pago: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var total, pagado float64
	var estado string
	err = tx.QueryRow(ctx, `
		SELECT p.total, e.nombre, COALESCE((SELECT SUM(pg.monto) FROM pagos pg WHERE pg.pedido_id = p.id), 0)
		FROM pedidos p
		JOIN estados_pedido e ON e.id = p.estado_actual_id
		WHERE p.id = $1 AND p.activo = TRUE
		FOR UPDATE OF p`, pedidoID).Scan(&total, &estado, &pagado)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPedidoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error verificando pedido para pago: %w", err)
	}
	if estado == "Cancelado" {
		return nil, ErrPagoPedidoCancelado
	}
	saldo := math.Round((total-pagado)*100) / 100
	if saldo <= 0 {
		return nil, ErrPedidoSinSaldo
	}
	if req.Monto > saldo {
		return nil, ErrPagoExcedeSaldo
	}

	var metodo models.MetodoPago
	err = tx.QueryRow(ctx, `SELECT id, nombre, activo FROM metodos_pago WHERE id = $1 AND activo = TRUE FOR KEY SHARE`, req.MetodoPagoID).Scan(&metodo.ID, &metodo.Nombre, &metodo.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMetodoPagoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error verificando metodo de pago: %w", err)
	}

	detalle := &models.PagoDetalle{MetodoPago: metodo}
	err = tx.QueryRow(ctx, `
		INSERT INTO pagos (pedido_id, metodo_pago_id, usuario_id, monto, referencia)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, pedido_id, metodo_pago_id, usuario_id, monto, referencia, fecha_pago, created_at`,
		pedidoID, req.MetodoPagoID, usuarioID, req.Monto, req.Referencia,
	).Scan(&detalle.ID, &detalle.PedidoID, &detalle.MetodoPagoID, &detalle.UsuarioID, &detalle.Monto, &detalle.Referencia, &detalle.FechaPago, &detalle.CreatedAt)
	if err != nil {
		if mapped := pagoForeignKeyError(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("error registrando pago: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando registro de pago: %w", err)
	}
	return detalle, nil
}

func pagoForeignKeyError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return nil
	}
	switch pgErr.ConstraintName {
	case "pagos_usuario_id_fkey":
		return ErrUsuarioNoEncontrado
	case "pagos_pedido_id_fkey":
		return ErrPedidoNoEncontrado
	case "pagos_metodo_pago_id_fkey":
		return ErrMetodoPagoNoEncontrado
	default:
		return nil
	}
}
