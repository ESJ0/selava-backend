package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPedidoNoEncontrado = errors.New("pedido no encontrado")

type PrendaRepository struct {
	db *pgxpool.Pool
}

func NewPrendaRepository(db *pgxpool.Pool) *PrendaRepository {
	return &PrendaRepository{db: db}
}

func (r *PrendaRepository) CreateMany(ctx context.Context, pedidoID int, reqs []models.PrendaCreateRequest) ([]models.Prenda, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando transaccion de prendas: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var pedidoExiste bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pedidos WHERE id = $1 AND activo = TRUE)`,
		pedidoID,
	).Scan(&pedidoExiste); err != nil {
		return nil, fmt.Errorf("error verificando pedido: %w", err)
	}
	if !pedidoExiste {
		return nil, ErrPedidoNoEncontrado
	}

	const query = `
		INSERT INTO prendas (pedido_id, tipo_prenda_id, descripcion, cantidad, color)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, pedido_id, tipo_prenda_id, descripcion, cantidad, color,
		          created_at, updated_at`

	prendas := make([]models.Prenda, 0, len(reqs))
	for _, req := range reqs {
		var prenda models.Prenda
		err := tx.QueryRow(ctx, query,
			pedidoID, req.TipoPrendaID, req.Descripcion, req.Cantidad, req.Color,
		).Scan(
			&prenda.ID, &prenda.PedidoID, &prenda.TipoPrendaID,
			&prenda.Descripcion, &prenda.Cantidad, &prenda.Color,
			&prenda.CreatedAt, &prenda.UpdatedAt,
		)
		if err != nil {
			if mappedErr := prendaForeignKeyError(err); mappedErr != nil {
				return nil, mappedErr
			}
			return nil, fmt.Errorf("error creando prenda: %w", err)
		}
		prendas = append(prendas, prenda)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando prendas: %w", err)
	}
	return prendas, nil
}

func prendaForeignKeyError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "prendas_pedido_id_fkey":
		return ErrPedidoNoEncontrado
	case "prendas_tipo_prenda_id_fkey":
		return ErrTipoPrendaNoEncontrado
	default:
		return nil
	}
}
