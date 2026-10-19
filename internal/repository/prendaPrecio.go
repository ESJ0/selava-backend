package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// La tarifa se lee y se congela en la misma transaccion. FOR SHARE impide
// que cambie durante la captura; los servicios son asociaciones sin recargo.
func insertarPrendaConPrecio(ctx context.Context, tx pgx.Tx, pedidoID int, req models.PrendaCreateRequest) (*models.Prenda, error) {
	var precio *float64
	err := tx.QueryRow(ctx, "SELECT precio FROM tipos_prenda WHERE id=$1 AND activo=TRUE FOR SHARE", req.TipoPrendaID).Scan(&precio)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTipoPrendaNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando precio de prenda: %w", err)
	}
	if precio == nil {
		return nil, ErrTipoPrendaSinPrecio
	}
	const query = `
		INSERT INTO prendas (pedido_id, tipo_prenda_id, descripcion, cantidad, color, precio_aplicado)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, pedido_id, tipo_prenda_id, descripcion, cantidad, color,
		          created_at, updated_at, precio_aplicado`
	prenda := models.Prenda{Servicios: make([]models.PrendaServicio, 0)}
	err = tx.QueryRow(ctx, query, pedidoID, req.TipoPrendaID, req.Descripcion, req.Cantidad, req.Color, precio).Scan(
		&prenda.ID, &prenda.PedidoID, &prenda.TipoPrendaID, &prenda.Descripcion, &prenda.Cantidad, &prenda.Color, &prenda.CreatedAt, &prenda.UpdatedAt, &prenda.PrecioAplicado)
	if err != nil {
		return nil, fmt.Errorf("error creando prenda con precio: %w", err)
	}
	return &prenda, nil
}
