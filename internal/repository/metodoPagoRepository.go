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

type MetodoPagoRepository struct{ db *pgxpool.Pool }

func NewMetodoPagoRepository(db *pgxpool.Pool) *MetodoPagoRepository {
	return &MetodoPagoRepository{db: db}
}
func scanMetodo(row pgx.Row) (*models.MetodoPago, error) {
	var m models.MetodoPago
	err := row.Scan(&m.ID, &m.Nombre, &m.Activo)
	return &m, err
}
func (r *MetodoPagoRepository) Create(ctx context.Context, req *models.MetodoPagoCreateRequest) (*models.MetodoPago, error) {
	m, err := scanMetodo(r.db.QueryRow(ctx, `INSERT INTO metodos_pago(nombre) VALUES($1) RETURNING id,nombre,activo`, req.Nombre))
	if isUniqueViolation(err) {
		return nil, ErrNombreMetodoPagoEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error creando metodo de pago: %w", err)
	}
	return m, nil
}
func (r *MetodoPagoRepository) GetByID(ctx context.Context, id int) (*models.MetodoPago, error) {
	m, err := scanMetodo(r.db.QueryRow(ctx, `SELECT id,nombre,activo FROM metodos_pago WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMetodoPagoNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo metodo de pago: %w", err)
	}
	return m, nil
}
func (r *MetodoPagoRepository) List(ctx context.Context) ([]models.MetodoPago, error) {
	rows, err := r.db.Query(ctx, `SELECT id,nombre,activo FROM metodos_pago ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("error listando metodos de pago: %w", err)
	}
	defer rows.Close()
	out := []models.MetodoPago{}
	for rows.Next() {
		var m models.MetodoPago
		if err := rows.Scan(&m.ID, &m.Nombre, &m.Activo); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *MetodoPagoRepository) Update(ctx context.Context, id int, req *models.MetodoPagoUpdateRequest) (*models.MetodoPago, error) {
	m, err := scanMetodo(r.db.QueryRow(ctx, `UPDATE metodos_pago SET nombre=COALESCE($1,nombre),activo=COALESCE($2,activo) WHERE id=$3 RETURNING id,nombre,activo`, req.Nombre, req.Activo, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMetodoPagoNoEncontrado
	}
	if isUniqueViolation(err) {
		return nil, ErrNombreMetodoPagoEnUso
	}
	if err != nil {
		return nil, fmt.Errorf("error actualizando metodo de pago: %w", err)
	}
	return m, nil
}
func (r *MetodoPagoRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE metodos_pago SET activo=FALSE WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrMetodoPagoNoEncontrado
	}
	return nil
}
