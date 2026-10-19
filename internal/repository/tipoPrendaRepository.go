package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTipoPrendaSinPrecio    = errors.New("el tipo de prenda no tiene precio configurado")
	ErrTipoPrendaNoEncontrado = errors.New("tipo de prenda no encontrado")
	ErrNombreTipoPrendaEnUso  = errors.New("ya existe un tipo de prenda con ese nombre")
)

type TipoPrendaRepository struct {
	db *pgxpool.Pool
}

func NewTipoPrendaRepository(db *pgxpool.Pool) *TipoPrendaRepository {
	return &TipoPrendaRepository{db: db}
}

func (r *TipoPrendaRepository) Create(ctx context.Context, req *models.TipoPrendaCreateRequest) (*models.TipoPrenda, error) {
	const query = `
		INSERT INTO tipos_prenda (nombre, descripcion, precio)
		VALUES ($1, $2, $3)
		RETURNING id, nombre, descripcion, activo, created_at, updated_at, precio`

	var tipo models.TipoPrenda
	err := r.db.QueryRow(ctx, query, req.Nombre, req.Descripcion, req.Precio).Scan(
		&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt, &tipo.Precio,
	)
	if isUniqueViolation(err) {
		return nil, ErrNombreTipoPrendaEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error creando tipo de prenda: %w", err)
	}
	return &tipo, nil
}

func (r *TipoPrendaRepository) GetByID(ctx context.Context, id int) (*models.TipoPrenda, error) {
	const query = `
		SELECT id, nombre, descripcion, activo, created_at, updated_at, precio
		FROM tipos_prenda WHERE id = $1`

	var tipo models.TipoPrenda
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt, &tipo.Precio,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTipoPrendaNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo tipo de prenda: %w", err)
	}
	return &tipo, nil
}

func (r *TipoPrendaRepository) List(ctx context.Context) ([]models.TipoPrenda, error) {
	const query = `
		SELECT id, nombre, descripcion, activo, created_at, updated_at, precio
		FROM tipos_prenda ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando tipos de prenda: %w", err)
	}
	defer rows.Close()

	tipos := make([]models.TipoPrenda, 0)
	for rows.Next() {
		var tipo models.TipoPrenda
		if err := rows.Scan(
			&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt, &tipo.Precio,
		); err != nil {
			return nil, fmt.Errorf("error leyendo tipo de prenda: %w", err)
		}
		tipos = append(tipos, tipo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando tipos de prenda: %w", err)
	}
	return tipos, nil
}

func (r *TipoPrendaRepository) Update(ctx context.Context, id int, req *models.TipoPrendaUpdateRequest) (*models.TipoPrenda, error) {
	const query = `
		UPDATE tipos_prenda SET
			nombre = COALESCE($1, nombre),
			descripcion = COALESCE($2, descripcion),
			activo = COALESCE($3, activo),
			precio = COALESCE($4, precio),
			updated_at = NOW()
		WHERE id = $5
		RETURNING id, nombre, descripcion, activo, created_at, updated_at, precio`

	var tipo models.TipoPrenda
	err := r.db.QueryRow(ctx, query, req.Nombre, req.Descripcion, req.Activo, req.Precio, id).Scan(
		&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt, &tipo.Precio,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTipoPrendaNoEncontrado
	}
	if isUniqueViolation(err) {
		return nil, ErrNombreTipoPrendaEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error actualizando tipo de prenda: %w", err)
	}
	return &tipo, nil
}

func (r *TipoPrendaRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE tipos_prenda SET activo = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error eliminando tipo de prenda: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTipoPrendaNoEncontrado
	}
	return nil
}

// ListActiveWithPrice excluye tipos pendientes de configurar durante la transicion.
func (r *TipoPrendaRepository) ListActiveWithPrice(ctx context.Context) ([]models.TipoPrendaPublico, error) {
	rows, err := r.db.Query(ctx, "SELECT id, nombre, descripcion, precio FROM tipos_prenda WHERE activo = TRUE AND precio IS NOT NULL ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("error listando precios publicos: %w", err)
	}
	defer rows.Close()
	tipos := make([]models.TipoPrendaPublico, 0)
	for rows.Next() {
		var tipo models.TipoPrendaPublico
		if err := rows.Scan(&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Precio); err != nil {
			return nil, err
		}
		tipos = append(tipos, tipo)
	}
	return tipos, rows.Err()
}
