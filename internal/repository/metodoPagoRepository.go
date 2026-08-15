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
	ErrMetodoPagoNoEncontrado = errors.New("metodo de pago no encontrado")
	ErrNombreMetodoPagoEnUso  = errors.New("ya existe un metodo de pago con ese nombre")
)

type MetodoPagoRepository struct {
	db *pgxpool.Pool
}

func NewMetodoPagoRepository(db *pgxpool.Pool) *MetodoPagoRepository {
	return &MetodoPagoRepository{db: db}
}

func (r *MetodoPagoRepository) Create(ctx context.Context, req *models.MetodoPagoCreateRequest) (*models.MetodoPago, error) {
	const query = `
		INSERT INTO metodos_pago (nombre)
		VALUES ($1)
		RETURNING id, nombre, activo`

	var metodo models.MetodoPago
	err := r.db.QueryRow(ctx, query, req.Nombre).Scan(&metodo.ID, &metodo.Nombre, &metodo.Activo)
	if isUniqueViolation(err) {
		return nil, ErrNombreMetodoPagoEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error creando metodo de pago: %w", err)
	}
	return &metodo, nil
}

func (r *MetodoPagoRepository) GetByID(ctx context.Context, id int) (*models.MetodoPago, error) {
	const query = `SELECT id, nombre, activo FROM metodos_pago WHERE id = $1`

	var metodo models.MetodoPago
	err := r.db.QueryRow(ctx, query, id).Scan(&metodo.ID, &metodo.Nombre, &metodo.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMetodoPagoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo metodo de pago: %w", err)
	}
	return &metodo, nil
}

func (r *MetodoPagoRepository) List(ctx context.Context) ([]models.MetodoPago, error) {
	const query = `SELECT id, nombre, activo FROM metodos_pago ORDER BY id`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando metodos de pago: %w", err)
	}
	defer rows.Close()

	metodos := make([]models.MetodoPago, 0)
	for rows.Next() {
		var metodo models.MetodoPago
		if err := rows.Scan(&metodo.ID, &metodo.Nombre, &metodo.Activo); err != nil {
			return nil, fmt.Errorf("error leyendo metodo de pago: %w", err)
		}
		metodos = append(metodos, metodo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando metodos de pago: %w", err)
	}
	return metodos, nil
}

func (r *MetodoPagoRepository) Update(ctx context.Context, id int, req *models.MetodoPagoUpdateRequest) (*models.MetodoPago, error) {
	const query = `
		UPDATE metodos_pago SET
			nombre = COALESCE($1, nombre),
			activo = COALESCE($2, activo)
		WHERE id = $3
		RETURNING id, nombre, activo`

	var metodo models.MetodoPago
	err := r.db.QueryRow(ctx, query, req.Nombre, req.Activo, id).Scan(&metodo.ID, &metodo.Nombre, &metodo.Activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMetodoPagoNoEncontrado
	}
	if isUniqueViolation(err) {
		return nil, ErrNombreMetodoPagoEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error actualizando metodo de pago: %w", err)
	}
	return &metodo, nil
}

func (r *MetodoPagoRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE metodos_pago SET activo = FALSE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error eliminando metodo de pago: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMetodoPagoNoEncontrado
	}
	return nil
}
