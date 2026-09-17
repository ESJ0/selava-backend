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
	ErrInsumoNoEncontrado = errors.New("insumo no encontrado")
	ErrNombreInsumoEnUso  = errors.New("ya existe un insumo con ese nombre")
)

type InsumoRepository struct {
	db *pgxpool.Pool
}

func NewInsumoRepository(db *pgxpool.Pool) *InsumoRepository {
	return &InsumoRepository{db: db}
}

func (r *InsumoRepository) Create(ctx context.Context, req *models.InsumoCreateRequest) (*models.Insumo, error) {
	const query = `
		INSERT INTO insumos (nombre, descripcion, unidad_medida, stock_actual, stock_minimo)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, nombre, descripcion, unidad_medida, stock_actual, stock_minimo,
		          activo, created_at, updated_at`

	var insumo models.Insumo
	err := r.db.QueryRow(ctx, query,
		req.Nombre, req.Descripcion, req.UnidadMedida, req.StockActual, req.StockMinimo,
	).Scan(
		&insumo.ID, &insumo.Nombre, &insumo.Descripcion, &insumo.UnidadMedida,
		&insumo.StockActual, &insumo.StockMinimo, &insumo.Activo,
		&insumo.CreatedAt, &insumo.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return nil, ErrNombreInsumoEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error creando insumo: %w", err)
	}
	return &insumo, nil
}

func (r *InsumoRepository) GetByID(ctx context.Context, id int) (*models.Insumo, error) {
	const query = `
		SELECT id, nombre, descripcion, unidad_medida, stock_actual, stock_minimo,
		       activo, created_at, updated_at
		FROM insumos WHERE id = $1`

	var insumo models.Insumo
	err := r.db.QueryRow(ctx, query, id).Scan(
		&insumo.ID, &insumo.Nombre, &insumo.Descripcion, &insumo.UnidadMedida,
		&insumo.StockActual, &insumo.StockMinimo, &insumo.Activo,
		&insumo.CreatedAt, &insumo.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInsumoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo insumo: %w", err)
	}
	return &insumo, nil
}

func (r *InsumoRepository) List(ctx context.Context) ([]models.Insumo, error) {
	const query = `
		SELECT id, nombre, descripcion, unidad_medida, stock_actual, stock_minimo,
		       activo, created_at, updated_at
		FROM insumos ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando insumos: %w", err)
	}
	defer rows.Close()

	insumos := make([]models.Insumo, 0)
	for rows.Next() {
		var insumo models.Insumo
		if err := rows.Scan(
			&insumo.ID, &insumo.Nombre, &insumo.Descripcion, &insumo.UnidadMedida,
			&insumo.StockActual, &insumo.StockMinimo, &insumo.Activo,
			&insumo.CreatedAt, &insumo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error leyendo insumo: %w", err)
		}
		insumos = append(insumos, insumo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando insumos: %w", err)
	}
	return insumos, nil
}

func (r *InsumoRepository) Update(ctx context.Context, id int, req *models.InsumoUpdateRequest) (*models.Insumo, error) {
	const query = `
		UPDATE insumos SET
			nombre = COALESCE($1, nombre),
			descripcion = COALESCE($2, descripcion),
			unidad_medida = COALESCE($3, unidad_medida),
			stock_actual = COALESCE($4, stock_actual),
			stock_minimo = COALESCE($5, stock_minimo),
			activo = COALESCE($6, activo),
			updated_at = NOW()
		WHERE id = $7
		RETURNING id, nombre, descripcion, unidad_medida, stock_actual, stock_minimo,
		          activo, created_at, updated_at`

	var insumo models.Insumo
	err := r.db.QueryRow(ctx, query,
		req.Nombre, req.Descripcion, req.UnidadMedida, req.StockActual,
		req.StockMinimo, req.Activo, id,
	).Scan(
		&insumo.ID, &insumo.Nombre, &insumo.Descripcion, &insumo.UnidadMedida,
		&insumo.StockActual, &insumo.StockMinimo, &insumo.Activo,
		&insumo.CreatedAt, &insumo.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInsumoNoEncontrado
	}
	if isUniqueViolation(err) {
		return nil, ErrNombreInsumoEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error actualizando insumo: %w", err)
	}
	return &insumo, nil
}

func (r *InsumoRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE insumos SET activo = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error eliminando insumo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInsumoNoEncontrado
	}
	return nil
}
