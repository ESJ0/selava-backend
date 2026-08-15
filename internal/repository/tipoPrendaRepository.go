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
		INSERT INTO tipos_prenda (nombre, descripcion)
		VALUES ($1, $2)
		RETURNING id, nombre, descripcion, activo, created_at, updated_at`

	var tipo models.TipoPrenda
	err := r.db.QueryRow(ctx, query, req.Nombre, req.Descripcion).Scan(
		&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt,
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
		SELECT id, nombre, descripcion, activo, created_at, updated_at
		FROM tipos_prenda WHERE id = $1`

	var tipo models.TipoPrenda
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt,
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
		SELECT id, nombre, descripcion, activo, created_at, updated_at
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
			&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt,
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
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, nombre, descripcion, activo, created_at, updated_at`

	var tipo models.TipoPrenda
	err := r.db.QueryRow(ctx, query, req.Nombre, req.Descripcion, req.Activo, id).Scan(
		&tipo.ID, &tipo.Nombre, &tipo.Descripcion, &tipo.Activo, &tipo.CreatedAt, &tipo.UpdatedAt,
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
