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
		for _, servicioReq := range req.Servicios {
			var relacion models.PrendaServicio
			err := tx.QueryRow(ctx, `INSERT INTO prenda_servicios(prenda_id,servicio_id,precio_aplicado) SELECT $1,id,precio_base FROM servicios WHERE id=$2 AND activo=TRUE RETURNING id,prenda_id,servicio_id,precio_aplicado,created_at,updated_at`, prenda.ID, servicioReq.ServicioID).Scan(
				&relacion.ID, &relacion.PrendaID, &relacion.ServicioID,
				&relacion.PrecioAplicado, &relacion.CreatedAt, &relacion.UpdatedAt,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrServicioNoEncontrado
			}
			if err != nil {
				return nil, fmt.Errorf("error asociando servicio a prenda: %w", err)
			}
			prenda.Servicios = append(prenda.Servicios, relacion)
		}
		prendas = append(prendas, prenda)
	}
	if _, err := tx.Exec(ctx, `UPDATE pedidos SET total=(SELECT COALESCE(SUM(p.cantidad*ps.precio_aplicado),0) FROM prendas p JOIN prenda_servicios ps ON ps.prenda_id=p.id WHERE p.pedido_id=$1),updated_at=NOW() WHERE id=$1`, pedidoID); err != nil {
		return nil, fmt.Errorf("error actualizando total: %w", err)
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
