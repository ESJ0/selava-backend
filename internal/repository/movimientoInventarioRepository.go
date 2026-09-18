package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrStockInsuficiente = errors.New("stock insuficiente para registrar la salida")

type MovimientoInventarioRepository struct {
	db *pgxpool.Pool
}

func NewMovimientoInventarioRepository(db *pgxpool.Pool) *MovimientoInventarioRepository {
	return &MovimientoInventarioRepository{db: db}
}

// Create registra el movimiento y modifica el stock dentro de una misma
// transaccion. El bloqueo del insumo evita que dos salidas concurrentes
// utilicen el mismo stock disponible.
func (r *MovimientoInventarioRepository) Create(ctx context.Context, req *models.MovimientoInventarioCreateRequest, usuarioID int) (*models.MovimientoInventario, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando registro de movimiento: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var usuarioExiste bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM usuarios WHERE id = $1 AND activo = TRUE)`, usuarioID,
	).Scan(&usuarioExiste); err != nil {
		return nil, fmt.Errorf("error verificando usuario para movimiento: %w", err)
	}
	if !usuarioExiste {
		return nil, ErrUsuarioNoEncontrado
	}

	var stockActual float64
	err = tx.QueryRow(ctx,
		`SELECT stock_actual FROM insumos WHERE id = $1 AND activo = TRUE FOR UPDATE`, req.InsumoID,
	).Scan(&stockActual)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInsumoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error verificando insumo para movimiento: %w", err)
	}

	nuevoStock := stockActual
	if req.TipoMovimiento == "entrada" {
		nuevoStock += req.Cantidad
	} else {
		nuevoStock, err = calcularStockSalida(stockActual, req.Cantidad)
		if err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx,
		`UPDATE insumos SET stock_actual = $1, updated_at = NOW() WHERE id = $2`,
		nuevoStock, req.InsumoID,
	); err != nil {
		return nil, fmt.Errorf("error actualizando stock del insumo: %w", err)
	}

	const query = `
		INSERT INTO movimientos_inventario (
			insumo_id, usuario_id, tipo_movimiento, cantidad, motivo
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, insumo_id, usuario_id, tipo_movimiento, cantidad,
		          motivo, fecha_movimiento, created_at`

	var movimiento models.MovimientoInventario
	err = tx.QueryRow(ctx, query,
		req.InsumoID, usuarioID, req.TipoMovimiento, req.Cantidad, req.Motivo,
	).Scan(
		&movimiento.ID, &movimiento.InsumoID, &movimiento.UsuarioID,
		&movimiento.TipoMovimiento, &movimiento.Cantidad, &movimiento.Motivo,
		&movimiento.FechaMovimiento, &movimiento.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInsumoNoEncontrado
		}
		return nil, fmt.Errorf("error registrando movimiento de inventario: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando registro de movimiento: %w", err)
	}
	return &movimiento, nil
}

func calcularStockSalida(stockActual, cantidad float64) (float64, error) {
	if cantidad > stockActual {
		return stockActual, ErrStockInsuficiente
	}
	return stockActual - cantidad, nil
}
