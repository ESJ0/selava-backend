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

var (
	ErrServicioNoEncontrado = errors.New("servicio no encontrado")
	ErrNombreServicioEnUso  = errors.New("ya existe un servicio con ese nombre")
)

type ServicioRepository struct {
	db *pgxpool.Pool
}

func NewServicioRepository(db *pgxpool.Pool) *ServicioRepository {
	return &ServicioRepository{db: db}
}

func (r *ServicioRepository) Create(ctx context.Context, req *models.ServicioCreateRequest) (*models.Servicio, error) {
	const query = `
		INSERT INTO servicios (nombre, descripcion, precio_base, tiempo_estimado_horas)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nombre, descripcion, precio_base, tiempo_estimado_horas,
		          activo, created_at, updated_at`

	var servicio models.Servicio
	err := r.db.QueryRow(ctx, query, req.Nombre, req.Descripcion, req.PrecioBase, req.TiempoEstimadoHoras).Scan(
		&servicio.ID, &servicio.Nombre, &servicio.Descripcion, &servicio.PrecioBase,
		&servicio.TiempoEstimadoHoras, &servicio.Activo, &servicio.CreatedAt, &servicio.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return nil, ErrNombreServicioEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error creando servicio: %w", err)
	}
	return &servicio, nil
}

func (r *ServicioRepository) GetByID(ctx context.Context, id int) (*models.Servicio, error) {
	const query = `
		SELECT id, nombre, descripcion, precio_base, tiempo_estimado_horas,
		       activo, created_at, updated_at
		FROM servicios WHERE id = $1`

	var servicio models.Servicio
	err := r.db.QueryRow(ctx, query, id).Scan(
		&servicio.ID, &servicio.Nombre, &servicio.Descripcion, &servicio.PrecioBase,
		&servicio.TiempoEstimadoHoras, &servicio.Activo, &servicio.CreatedAt, &servicio.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrServicioNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo servicio: %w", err)
	}
	return &servicio, nil
}

func (r *ServicioRepository) List(ctx context.Context) ([]models.Servicio, error) {
	const query = `
		SELECT id, nombre, descripcion, precio_base, tiempo_estimado_horas,
		       activo, created_at, updated_at
		FROM servicios ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando servicios: %w", err)
	}
	defer rows.Close()

	servicios := make([]models.Servicio, 0)
	for rows.Next() {
		var servicio models.Servicio
		if err := rows.Scan(
			&servicio.ID, &servicio.Nombre, &servicio.Descripcion, &servicio.PrecioBase,
			&servicio.TiempoEstimadoHoras, &servicio.Activo, &servicio.CreatedAt, &servicio.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error leyendo servicio: %w", err)
		}
		servicios = append(servicios, servicio)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando servicios: %w", err)
	}
	return servicios, nil
}

func (r *ServicioRepository) Update(ctx context.Context, id int, req *models.ServicioUpdateRequest) (*models.Servicio, error) {
	const query = `
		UPDATE servicios SET
			nombre = COALESCE($1, nombre),
			descripcion = COALESCE($2, descripcion),
			precio_base = COALESCE($3, precio_base),
			tiempo_estimado_horas = COALESCE($4, tiempo_estimado_horas),
			activo = COALESCE($5, activo),
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, nombre, descripcion, precio_base, tiempo_estimado_horas,
		          activo, created_at, updated_at`

	var servicio models.Servicio
	err := r.db.QueryRow(ctx, query,
		req.Nombre, req.Descripcion, req.PrecioBase, req.TiempoEstimadoHoras, req.Activo, id,
	).Scan(
		&servicio.ID, &servicio.Nombre, &servicio.Descripcion, &servicio.PrecioBase,
		&servicio.TiempoEstimadoHoras, &servicio.Activo, &servicio.CreatedAt, &servicio.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrServicioNoEncontrado
	}
	if isUniqueViolation(err) {
		return nil, ErrNombreServicioEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error actualizando servicio: %w", err)
	}
	return &servicio, nil
}

func (r *ServicioRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE servicios SET activo = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error eliminando servicio: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrServicioNoEncontrado
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
