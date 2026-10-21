package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/ESJ0/selava-backend/internal/promociones"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sort"
	"time"
)

var ErrPromocionNoEncontrada = errors.New("promocion no encontrada")
var ErrPromocionInvalida = errors.New("promocion no elegible: verifica vigencia, modalidad, medida y seleccion unica por tipo")

type PromocionRepository struct{ db *pgxpool.Pool }

func NewPromocionRepository(db *pgxpool.Pool) *PromocionRepository {
	return &PromocionRepository{db: db}
}

const promocionColumns = `p.id,p.nombre,p.descripcion,p.tipo_prenda_id,p.modalidad_cobro,p.medida_requerida,p.precio_promocional,p.fecha_inicio,p.fecha_fin,p.activo,p.created_at,p.updated_at,t.nombre`

func scanPromocion(row pgx.Row) (models.Promocion, error) {
	var p models.Promocion
	err := row.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.TipoPrendaID, &p.ModalidadCobro, &p.MedidaRequerida, &p.PrecioPromocional, &p.FechaInicio, &p.FechaFin, &p.Activo, &p.CreatedAt, &p.UpdatedAt, &p.TipoPrendaNombre)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrPromocionNoEncontrada
	}
	return p, err
}
func (r *PromocionRepository) List(ctx context.Context) ([]models.Promocion, error) {
	rows, err := r.db.Query(ctx, `SELECT `+promocionColumns+` FROM promociones p JOIN tipos_prenda t ON t.id=p.tipo_prenda_id ORDER BY p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.Promocion, 0)
	for rows.Next() {
		p, err := scanPromocion(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
func (r *PromocionRepository) GetByID(ctx context.Context, id int) (*models.Promocion, error) {
	p, err := scanPromocion(r.db.QueryRow(ctx, `SELECT `+promocionColumns+` FROM promociones p JOIN tipos_prenda t ON t.id=p.tipo_prenda_id WHERE p.id=$1`, id))
	return &p, err
}

func (r *PromocionRepository) Save(ctx context.Context, id int, req *models.PromocionRequest) (*models.Promocion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var mode models.ModalidadCobro
	err = tx.QueryRow(ctx, `SELECT modalidad_cobro FROM tipos_prenda WHERE id=$1 AND activo FOR SHARE`, req.TipoPrendaID).Scan(&mode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTipoPrendaNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	if mode != req.ModalidadCobro {
		return nil, ErrPromocionInvalida
	}
	args := []any{req.Nombre, req.Descripcion, req.TipoPrendaID, req.ModalidadCobro, req.MedidaRequerida, req.PrecioPromocional, req.FechaInicio, req.FechaFin, req.Activo}
	if id == 0 {
		err = tx.QueryRow(ctx, `INSERT INTO promociones(nombre,descripcion,tipo_prenda_id,modalidad_cobro,medida_requerida,precio_promocional,fecha_inicio,fecha_fin,activo) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, args...).Scan(&id)
	} else {
		args = append(args, id)
		tag, e := tx.Exec(ctx, `UPDATE promociones SET nombre=$1,descripcion=$2,tipo_prenda_id=$3,modalidad_cobro=$4,medida_requerida=$5,precio_promocional=$6,fecha_inicio=$7,fecha_fin=$8,activo=$9,updated_at=NOW() WHERE id=$10`, args...)
		err = e
		if err == nil && tag.RowsAffected() == 0 {
			return nil, ErrPromocionNoEncontrada
		}
	}
	if err != nil {
		return nil, err
	}
	p, err := scanPromocion(tx.QueryRow(ctx, `SELECT `+promocionColumns+` FROM promociones p JOIN tipos_prenda t ON t.id=p.tipo_prenda_id WHERE p.id=$1`, id))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *PromocionRepository) SetActivo(ctx context.Context, id int, activo bool) (*models.Promocion, error) {
	tag, err := r.db.Exec(ctx, `UPDATE promociones SET activo=$2,updated_at=NOW() WHERE id=$1`, id, activo)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrPromocionNoEncontrada
	}
	return r.GetByID(ctx, id)
}

func (r *PromocionRepository) Cotizar(ctx context.Context, req *models.PedidoCreateRequest) (*models.PromocionCotizacion, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	quote, err := cotizarPromociones(ctx, tx, req)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return quote, nil
}

// Precio y modalidad se bloquean en orden estable y se leen de la misma
// transacción que inserta las prendas. Nunca se aceptan importes del navegador.
func cotizarPromociones(ctx context.Context, tx pgx.Tx, req *models.PedidoCreateRequest) (*models.PromocionCotizacion, error) {
	byType := map[int][]models.PrendaCreateRequest{}
	for _, p := range req.Prendas {
		byType[p.TipoPrendaID] = append(byType[p.TipoPrendaID], p)
	}
	ids := make([]int, 0, len(byType))
	for id := range byType {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	selected := map[int]int{}
	for _, s := range req.Promociones {
		if s.TipoPrendaID <= 0 || s.PromocionID <= 0 || selected[s.TipoPrendaID] != 0 || len(byType[s.TipoPrendaID]) == 0 {
			return nil, ErrPromocionInvalida
		}
		selected[s.TipoPrendaID] = s.PromocionID
	}
	var now time.Time
	if err := tx.QueryRow(ctx, `SELECT CURRENT_TIMESTAMP`).Scan(&now); err != nil {
		return nil, err
	}
	result := &models.PromocionCotizacion{Grupos: make([]models.PromocionCotizacionGrupo, 0)}
	var total int64
	for _, id := range ids {
		g := models.PromocionCotizacionGrupo{TipoPrendaID: id, Elegibles: make([]models.PromocionAplicada, 0)}
		var price *float64
		err := tx.QueryRow(ctx, `SELECT nombre,modalidad_cobro,precio FROM tipos_prenda WHERE id=$1 AND activo FOR SHARE`, id).Scan(&g.TipoPrendaNombre, &g.ModalidadCobro, &price)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTipoPrendaNoEncontrado
		}
		if err != nil {
			return nil, err
		}
		if price == nil {
			return nil, ErrTipoPrendaSinPrecio
		}
		cents, err := promociones.Escalar(*price, 100)
		if err != nil {
			return nil, ErrPromocionInvalida
		}
		g.PrecioNormal = *price
		var measure, base int64
		for _, p := range byType[id] {
			m, err := promociones.Escalar(p.Cantidad, 1000)
			if err != nil || m <= 0 || m > 99999999999 {
				return nil, ErrPromocionInvalida
			}
			if g.ModalidadCobro == models.ModalidadCobroUnidad && m%1000 != 0 {
				return nil, ErrMedidaInvalidaModalidad
			}
			line, err := promociones.Subtotal(cents, m)
			if err != nil {
				return nil, ErrTotalPedidoFueraDeRango
			}
			measure += m
			base += line
			if base > promociones.MaxCentavos {
				return nil, ErrTotalPedidoFueraDeRango
			}
		}
		g.Medida = float64(measure) / 1000
		g.SubtotalBase = float64(base) / 100
		g.SubtotalTotal = g.SubtotalBase
		rows, err := tx.Query(ctx, `SELECT `+promocionColumns+` FROM promociones p JOIN tipos_prenda t ON t.id=p.tipo_prenda_id WHERE p.tipo_prenda_id=$1 AND p.activo AND p.fecha_inicio<=$2 AND p.fecha_fin>=$2 ORDER BY p.id FOR SHARE OF p`, id, now)
		if err != nil {
			return nil, err
		}
		candidates := make([]models.Promocion, 0)
		for rows.Next() {
			p, err := scanPromocion(rows)
			if err != nil {
				rows.Close()
				return nil, err
			}
			candidates = append(candidates, p)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		final := base
		for _, p := range candidates {
			a, err := promociones.Aplicar(p, g.ModalidadCobro, id, measure, cents, base, now)
			if err != nil {
				continue
			}
			g.Elegibles = append(g.Elegibles, *a)
			if selected[id] == p.ID {
				g.Aplicada = a
				g.SubtotalTotal = a.SubtotalTotal
				final, _ = promociones.Escalar(a.SubtotalTotal, 100)
			}
		}
		if selected[id] != 0 && g.Aplicada == nil {
			return nil, ErrPromocionInvalida
		}
		total += final
		if total > promociones.MaxCentavos {
			return nil, ErrTotalPedidoFueraDeRango
		}
		result.Grupos = append(result.Grupos, g)
	}
	result.Total = float64(total) / 100
	return result, nil
}

func guardarPromociones(ctx context.Context, tx pgx.Tx, pedidoID int, quote *models.PromocionCotizacion, prendas []models.Prenda) ([]models.PromocionAplicada, error) {
	result := make([]models.PromocionAplicada, 0)
	for _, g := range quote.Grupos {
		if g.Aplicada == nil {
			continue
		}
		a := *g.Aplicada
		for _, p := range prendas {
			if p.TipoPrendaID == a.TipoPrendaID {
				a.PrendaIDs = append(a.PrendaIDs, p.ID)
			}
		}
		data, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `INSERT INTO pedido_promociones(pedido_id,promocion_id,tipo_prenda_id,subtotal_base,subtotal_total,snapshot) VALUES($1,$2,$3,$4,$5,$6)`, pedidoID, a.PromocionID, a.TipoPrendaID, a.SubtotalBase, a.SubtotalTotal, data)
		if err != nil {
			return nil, fmt.Errorf("guardando promocion aplicada: %w", err)
		}
		result = append(result, a)
	}
	return result, nil
}
func cargarPromociones(ctx context.Context, tx pgx.Tx, pedidoID int) ([]models.PromocionAplicada, error) {
	rows, err := tx.Query(ctx, `SELECT snapshot FROM pedido_promociones WHERE pedido_id=$1 ORDER BY id`, pedidoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.PromocionAplicada, 0)
	for rows.Next() {
		var data []byte
		var a models.PromocionAplicada
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
