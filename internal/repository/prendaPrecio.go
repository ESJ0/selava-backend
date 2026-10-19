package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// La tarifa se lee y se congela en la misma transaccion. FOR SHARE impide
// que cambie durante la captura; los servicios son asociaciones sin recargo.
func insertarPrendaConPrecio(ctx context.Context, tx pgx.Tx, pedidoID int, req models.PrendaCreateRequest) (*models.Prenda, error) {
	var precio *float64
	var modalidad models.ModalidadCobro
	err := tx.QueryRow(ctx, "SELECT precio, modalidad_cobro FROM tipos_prenda WHERE id=$1 AND activo=TRUE FOR SHARE", req.TipoPrendaID).Scan(&precio, &modalidad)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTipoPrendaNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando precio de prenda: %w", err)
	}
	if precio == nil {
		return nil, ErrTipoPrendaSinPrecio
	}
	if modalidad == models.ModalidadCobroUnidad && req.Cantidad != math.Trunc(req.Cantidad) {
		return nil, ErrMedidaInvalidaModalidad
	}
	const query = `
		INSERT INTO prendas (pedido_id, tipo_prenda_id, descripcion, cantidad, color, precio_aplicado, modalidad_cobro_aplicada)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, pedido_id, tipo_prenda_id, descripcion, cantidad, color,
		          created_at, updated_at, precio_aplicado, modalidad_cobro_aplicada`
	prenda := models.Prenda{Servicios: make([]models.PrendaServicio, 0)}
	err = tx.QueryRow(ctx, query, pedidoID, req.TipoPrendaID, req.Descripcion, req.Cantidad, req.Color, precio, modalidad).Scan(
		&prenda.ID, &prenda.PedidoID, &prenda.TipoPrendaID, &prenda.Descripcion, &prenda.Cantidad, &prenda.Color, &prenda.CreatedAt, &prenda.UpdatedAt, &prenda.PrecioAplicado, &prenda.ModalidadCobroAplicada)
	if err != nil {
		return nil, fmt.Errorf("error creando prenda con precio: %w", err)
	}
	return &prenda, nil
}
